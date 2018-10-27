package tuning

import (
	"time"

	"github.com/robfig/cron"
)

// Clock - clock
type Clock struct {
	callback func(bool)
	super    bool
}

var cronJob *cron.Cron

func initClock() {
	if cronJob == nil {
		cronJob = cron.New()
		cronJob.Start()
	}
}

// Super - switck `super` flag
func (c *Clock) Super() *Clock {
	c.super = !c.super
	return c
}

// Immediately - immediately run
func (c *Clock) Immediately() {
	c.callback(c.super)
}

// Cron -
// `spec` see: github.com/robfig/cron
// usage: (*tuning.Clock).Cron(spec) == (*cron.Cron).AddFunc(spec, callback)
func (c *Clock) Cron(spec string) {
	initClock()
	cronJob.AddFunc(spec, func() {
		c.callback(c.super)
	})
}

// After - run after
func (c *Clock) After(d time.Duration) {
	timer := time.NewTimer(d)
	cantor.wg.Add(1)
	go func() {
		defer func() {
			cantor.wg.Done()
		}()
		<-timer.C
		c.callback(c.super)
	}()
}

// On - run on
func (c *Clock) On(t time.Time) (d time.Duration) {
	d = t.Sub(time.Now())
	c.After(d)
	return
}
