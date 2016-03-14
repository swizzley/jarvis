package jobs

import (
	"github.com/blendlabs/go-chronometer"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/jarvis/core"
)

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
	return chronometer.OnTheHour{}
}
