package mpsc

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQueue_PushPop(t *testing.T) {
	q := New()

	q.Push(1)
	q.Push(2)
	assert.Equal(t, 1, q.Pop())
	assert.Equal(t, 2, q.Pop())
	assert.True(t, q.Empty())
}

func TestQueue_Empty(t *testing.T) {
	q := New()
	assert.True(t, q.Empty())
	q.Push(1)
	assert.False(t, q.Empty())
}

func TestMpscQueueConsistency(t *testing.T) {
	max := 1000000
	c := 100
	var wg sync.WaitGroup
	wg.Add(1)
	q := New()

	var alreadySeen int32
	go func() {
		i := 0
		seen := make(map[string]string)
		for {
			r := q.Pop()
			if r == nil {
				runtime.Gosched()
				continue
			}
			i++
			s := r.(string)
			_, present := seen[s]
			if present {
				log.Printf("item have already been seen %v", s)
				atomic.StoreInt32(&alreadySeen, 1)
				return
			}
			seen[s] = s
			if i == max {
				wg.Done()
				return
			}
		}
	}()

	for j := 0; j < c; j++ {
		jj := j
		cmax := max / c
		go func() {
			for i := 0; i < cmax; i++ {
				if rand.Intn(10) == 0 {
					time.Sleep(time.Duration(rand.Intn(1000)))
				}
				q.Push(fmt.Sprintf("%v %v", jj, i))
			}
		}()
	}

	wg.Wait()
	time.Sleep(500 * time.Millisecond)

	if atomic.LoadInt32(&alreadySeen) == 1 {
		t.FailNow()
	}
	// queue should be empty
	for i := 0; i < 100; i++ {
		r := q.Pop()
		if r != nil {
			log.Printf("unexpected result %+v", r)
			t.FailNow()
		}
	}
}
