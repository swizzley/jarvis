package jobs

import (
	"time"

	"github.com/blendlabs/go-chronometer"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/jarvis/core"
)

// OnTheQuarterHour is a schedule that fires every 15 minutes, on the quarter hours.
type OnTheQuarterHour struct{}

// GetNextRunTime implements the chronometer Schedule api.
func (o OnTheQuarterHour) GetNextRunTime(after *time.Time) time.Time {
	var returnValue time.Time
	if after == nil {
		now := time.Now().UTC()
		if now.Minute() >= 45 {
			returnValue = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 45, 0, 0, time.UTC).Add(15 * time.Minute)
		} else if now.Minute() >= 30 {
			returnValue = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 30, 0, 0, time.UTC).Add(15 * time.Minute)
		} else if now.Minute() >= 15 {
			returnValue = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 15, 0, 0, time.UTC).Add(15 * time.Minute)
		} else {
			returnValue = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC).Add(15 * time.Minute)
		}
	} else {
		if after.Minute() >= 45 {
			returnValue = time.Date(after.Year(), after.Month(), after.Day(), after.Hour(), 45, 0, 0, time.UTC).Add(15 * time.Minute)
		} else if after.Minute() >= 30 {
			returnValue = time.Date(after.Year(), after.Month(), after.Day(), after.Hour(), 30, 0, 0, time.UTC).Add(15 * time.Minute)
		} else if after.Minute() >= 15 {
			returnValue = time.Date(after.Year(), after.Month(), after.Day(), after.Hour(), 15, 0, 0, time.UTC).Add(15 * time.Minute)
		} else {
			returnValue = time.Date(after.Year(), after.Month(), after.Day(), after.Hour(), 0, 0, 0, time.UTC).Add(15 * time.Minute)
		}
	}
	return returnValue
}

// OnTheHour is a schedule that fires every hour on the 00th minute.
type OnTheHour struct{}

// GetNextRunTime implements the chronometer Schedule api.
func (o OnTheHour) GetNextRunTime(after *time.Time) time.Time {
	var returnValue time.Time
	if after == nil {
		now := time.Now().UTC()
		returnValue = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC).Add(1 * time.Hour)
	} else {
		returnValue = time.Date(after.Year(), after.Month(), after.Day(), after.Hour(), 0, 0, 0, time.UTC).Add(1 * time.Hour)
	}
	return returnValue
}

// NewClock returns a new clock job instance.
func NewClock(j core.Bot) *Clock {
	return &Clock{Bot: j}
}

// Clock is a job that announces the time through a given bot.
type Clock struct {
	Bot core.Bot
}

// Name returns the name of the chronometer job.
func (t Clock) Name() string {
	return "clock"
}

// Execute is the actual code that runs when the job is fired.
func (t Clock) Execute(ct *chronometer.CancellationToken) error {
	for x := 0; x < len(t.Bot.ActiveChannels()); x++ {
		channelID := t.Bot.ActiveChannels()[x]
		err := t.Bot.TriggerAction("time", &slack.Message{Channel: channelID})
		if err != nil {
			return err
		}
	}
	return nil
}

// Schedule returns the job schedule.
func (t Clock) Schedule() chronometer.Schedule {
	return OnTheHour{}
}
