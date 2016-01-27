package main

import (
	"testing"
	"time"

	"github.com/blendlabs/go-assert"
)

func TestOnTheQuarter(t *testing.T) {
	a := assert.New(t)

	s := OnTheQuarterHour{}

	now := time.Now().UTC()
	notNow := s.GetNextRunTime(nil)
	diff := notNow.Sub(now)
	a.True(diff > time.Millisecond)

	justBeforeNoon := time.Date(2016, 01, 01, 11, 46, 00, 00, time.UTC)
	noon := s.GetNextRunTime(&justBeforeNoon)
	a.Equal(12, noon.Hour())
	a.Equal(00, noon.Minute())
	a.Equal(00, noon.Second())
	a.Equal(justBeforeNoon.Year(), noon.Year())
	a.Equal(justBeforeNoon.Month(), noon.Month())
	a.Equal(justBeforeNoon.Day(), noon.Day())

	justBeforeQuarterTo := time.Date(2016, 01, 01, 11, 44, 00, 00, time.UTC)
	quarterTo := s.GetNextRunTime(&justBeforeQuarterTo)
	a.Equal(11, quarterTo.Hour())
	a.Equal(45, quarterTo.Minute())
	a.Equal(00, quarterTo.Second())
	a.Equal(justBeforeQuarterTo.Year(), quarterTo.Year())
	a.Equal(justBeforeQuarterTo.Month(), quarterTo.Month())
	a.Equal(justBeforeQuarterTo.Day(), quarterTo.Day())

	justBeforeThirty := time.Date(2016, 01, 01, 11, 25, 00, 00, time.UTC)
	thirty := s.GetNextRunTime(&justBeforeThirty)
	a.Equal(11, thirty.Hour())
	a.Equal(30, thirty.Minute())
	a.Equal(00, thirty.Second())
	a.Equal(justBeforeThirty.Year(), thirty.Year())
	a.Equal(justBeforeThirty.Month(), thirty.Month())
	a.Equal(justBeforeThirty.Day(), thirty.Day())

	onATick := time.Date(2016, 01, 01, 11, 30, 00, 00, time.UTC)
	quarterTo = s.GetNextRunTime(&onATick)
	a.Equal(11, quarterTo.Hour())
	a.Equal(45, quarterTo.Minute())
	a.Equal(00, quarterTo.Second())
	a.Equal(onATick.Year(), quarterTo.Year())
	a.Equal(onATick.Month(), quarterTo.Month())
	a.Equal(onATick.Day(), quarterTo.Day())
}
