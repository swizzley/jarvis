package jarvis

import (
	"time"

	"github.com/blendlabs/go-chronometer"
)

type OnTheQuarterHour struct{}

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

type OnTheHour struct{}

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

func NewClock(j *JarvisBot) *Clock {
	return &Clock{Bot: j}
}

type Clock struct {
	Bot *JarvisBot
}

func (t Clock) Name() string {
	return "clock"
}

func (t Clock) Execute(ct *chronometer.CancellationToken) error {
	currentTime := time.Now().UTC()

	for x := 0; x < len(t.Bot.Client.ActiveChannels); x++ {
		channelId := t.Bot.Client.ActiveChannels[x]
		err := t.Bot.AnnounceTime(channelId, currentTime)
		if err != nil {
			t.Bot.Logf("Error announcing time: %v", err)
		}
	}
	return nil
}

func (t Clock) Schedule() chronometer.Schedule {
	return OnTheHour{}
}