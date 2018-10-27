package tuning

import (
	"container/list"
	"sync"
)

type trackSign int

const (
	_ trackSign = iota
	play
	stop
	finish
)

// Track - the task
type Track struct {
	melody      *Melody
	next        *list.List
	in          chan interface{}
	sign        chan trackSign
	runIsCalled bool
}

// NewTrack - new task
func NewTrack(m *Melody, bufLen int) *Track {
	return &Track{
		melody:      m,
		next:        list.New(),
		in:          make(chan interface{}, bufLen),
		sign:        make(chan trackSign),
		runIsCalled: false,
	}
}

// Run - init
func (t *Track) run() {
	if t.runIsCalled {
		return
	}
	t.runIsCalled = true
	running := true
	unlock := true
	var wg = &sync.WaitGroup{}
DONE:
	for {
		select {
		case sign := <-t.sign:
			switch sign {
			case play:
				running = true
			case stop:
				running = false
			case finish:
				close(t.in)
				close(t.sign)
				break DONE
			default:
				break DONE
			}
		default:
			if running && unlock {
				unlock = false
				// fmt.Println("lock", *t.melody)
				wg.Add(1)
				go func() {
					defer func() {
						// fmt.Println("unlock", *t.melody)
						unlock = true
						wg.Done()
					}()
					data := <-t.in
					if data == nil {
						// fmt.Println("Warning: nil")
						return
					}
					ans := (*t.melody).Play(t, data)
					// fmt.Println("ans", ans)

					for e := t.next.Front(); e != nil; e = e.Next() {
						wg.Add(1)
						go func(e *list.Element) {
							defer func() {
								err := recover()
								if err != nil {
									next := e.Next()
									t.next.Remove(e)
									e = next
								}
								wg.Done()
							}()
							RunTrack(e.Value.(*Track)).Immediately()
							e.Value.(*Track).in <- ans
						}(e)
					}
				}()
			}
			// TODO: time.Sleep ?
		}
	}
	wg.Wait()
}

// Play - continue task
func (t *Track) Play() {
	t.sign <- play
}

// Stop - stop task
func (t *Track) Stop() {
	t.sign <- stop
}

// Finish - finish task
func (t *Track) Finish() {
	t.sign <- finish
}

// Before - t before n
func (t *Track) Before(n *Track) *Track {
	for e := t.next.Front(); e != nil; e = e.Next() {
		if e.Value == n {
			return t
		}
	}
	t.next.PushBack(n)
	return t
}

// After - t after n
func (t *Track) After(n *Track) *Track {
	n.Before(t)
	return t
}

// Inputer - input to track's in chan
func (t *Track) Inputer(in interface{}) {
	t.in <- in
}
