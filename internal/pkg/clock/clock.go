package clock

import "time"

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (c RealClock) Now() time.Time {
	return time.Now()
}

type FakeClock struct {
	currentTime time.Time
}

func NewFakeClock(initialTime time.Time) *FakeClock {
	return &FakeClock{currentTime: initialTime}
}

func (c *FakeClock) Now() time.Time {
	return c.currentTime
}

func (c *FakeClock) Advance(d time.Duration) {
	c.currentTime = c.currentTime.Add(d)
}