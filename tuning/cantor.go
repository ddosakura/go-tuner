package tuning

import (
	"fmt"
	"sync"
)

// Cantor - then controller
type Cantor struct {
	wg       *sync.WaitGroup
	MaxTrack int
	trackNum int
}

var cantor = newCantor()

// GetCantor - get cantor
func GetCantor() *Cantor {
	return cantor
}

func newCantor() *Cantor {
	return &Cantor{
		wg:       &sync.WaitGroup{},
		MaxTrack: 0,
		trackNum: 0,
	}
}

// Wait - wait for all goroutine
func Wait() {
	cantor.wait()
}
func (c *Cantor) wait() {
	c.wg.Wait()
}

// RunTrack - init track
func RunTrack(t *Track) *Clock {
	return cantor.runTrack(t)
}
func (c *Cantor) runTrack(t *Track) *Clock {
	return &Clock{
		callback: func(super bool) {
			if super || c.MaxTrack == 0 || c.trackNum < c.MaxTrack {
				cantor.wg.Add(1)
				cantor.trackNum++
				fmt.Println("wg", cantor.trackNum)
				go func() {
					defer func() {
						cantor.trackNum--
						fmt.Println("wg", cantor.trackNum)
						cantor.wg.Done()
					}()
					// fmt.Println("run", *t.melody)
					t.run()
				}()
			}
		},
		super: false,
	}
}
