package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Mutex struct {
	c chan interface{}
}

func (m *Mutex) Lock() {
	m.c <- struct{}{}
}

func (m *Mutex) UnLock() {
	<-m.c
}

func (m *Mutex) TryLock(seconds int) bool {

	tm := time.NewTimer(time.Second)

	select {
	case m.c <- struct{}{}:
		tm.Stop()
		return true
	case <-time.After(time.Second):
		return false
	}

	return false
}

type frk struct {
	mut *Mutex
}

func (f *frk) peekUp() bool {
	f.mut.Lock()
	return true
}

func (f *frk) pushDown() bool {
	f.mut.UnLock()
	return true
}

type philosopher struct {
	left, right *frk
	number      int
}

func (ph *philosopher) eat() {
	fmt.Printf("Philsosopher %d picks up left\n", ph.number)
	leftLock := ph.left.mut.TryLock(1)

	time.Sleep(time.Second * 3)

	if leftLock {
		rightLock := ph.right.mut.TryLock(1)
		if !rightLock {
			ph.left.mut.UnLock()

			return
		}
	}

	fmt.Printf("Philosopher %d eating\n", ph.number)

	ph.right.mut.UnLock()
	ph.left.mut.UnLock()
	fmt.Printf("Philosopher %d fead up\n", ph.number)

	return
}

func main() {
	var count = flag.Int("count", 3, "helper")
	flag.Parse()

	if *count > 10 || *count < 0 {
		fmt.Println("Count must be between 0 and 10")
		os.Exit(0)
	}

	fmt.Printf("Count of phils: %d\n", *count)

	phils := make([]*philosopher, *count)
	frks := make([]frk, *count)

	for i := 0; i < *count; i++ {
		frks[i].mut = &Mutex{c: make(chan interface{}, 1)}
	}

	wg := sync.WaitGroup{}

	for i := 0; i < *count; i++ {

		leftNum := i
		rightNum := i + 1
		if rightNum >= *count {
			rightNum = 0
		}

		phils[i] = &philosopher{
			left:   &frks[leftNum],
			right:  &frks[rightNum],
			number: i,
		}
	}

	wg.Add(*count)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	stopChan := make(chan interface{}, 1)

	go func() {

		for {
			select {
			case sig := <-c:
				if sig == syscall.SIGINT {
					fmt.Printf("SIGINT!!!!\n")
					close(stopChan)
				}
			}
		}
	}()

	for _, p := range phils {
		//fmt.Printf("Want eat: %d\n", p.number)

		go func(p *philosopher) {
			defer wg.Done()

			for {
				select {
				case <-stopChan:
					fmt.Printf("CLOSED CHAN FOR PHILOSOPHER %d\n", p.number)
					return
				default:
					p.eat()
				}
			}

		}(p)
	}

	//fmt.Printf("Start waiting\n")
	wg.Wait()
	fmt.Println("HI")
}
