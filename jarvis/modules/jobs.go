package modules

import (
	"fmt"
	"strings"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/jarvis/core"
	"github.com/wcharczuk/jarvis/jarvis/jobs"
)

const (
	// ModuleJobs is the name of the jobs module.
	ModuleJobs = "jobs"

	//ActionJobs is the jobs action id.
	ActionJobs = "jobs"

	//ActionJobRun is the jobs run action id.
	ActionJobRun = "job.run"

	//ActionJobCancel is the jobs cancel action id.
	ActionJobCancel = "job.cancel"

	//ActionJobEnable is the jobs enable action id.
	ActionJobEnable = "job.enable"

	//ActionJobDisable is the jobs disable action id.
	ActionJobDisable = "job.disable"
)

// Jobs is the module that governs jobs within a bot.
type Jobs struct{}

// Init does nothing right now.
func (j *Jobs) Init(b core.Bot) error {
	b.JobManager().LoadJob(jobs.NewClock(b))
	b.JobManager().DisableJob("clock")
	return nil
}

// Name returns the name of the module
func (j *Jobs) Name() string {
	return ModuleJobs
}

// Actions are all the actions the module provides.
func (j *Jobs) Actions() []core.Action {
	return []core.Action{
		core.Action{ID: ActionJobs, MessagePattern: "^jobs", Description: "Prints the current jobs and their statuses.", Handler: j.handleJobsStatus},
		core.Action{ID: ActionJobRun, MessagePattern: "^job:run", Description: "Runs all jobs", Handler: j.handleJobRun},
		core.Action{ID: ActionJobCancel, MessagePattern: "^job:cancel", Description: "Cancels a running job.", Handler: j.handleJobCancel},
		core.Action{ID: ActionJobEnable, MessagePattern: "^job:enable", Description: "Enables a job.", Handler: j.handleJobEnable},
		core.Action{ID: ActionJobDisable, MessagePattern: "^job:disable", Description: "Disables enables a job.", Handler: j.handleJobDisable},
	}
}

func (j *Jobs) handleJobsStatus(b core.Bot, m *slack.Message) error {
	statusText := "current job statuses:\n"
	for _, status := range b.JobManager().Status() {
		if len(status.RunningFor) != 0 {
			statusText = statusText + fmt.Sprintf(">`%s` - state: %s running for: %s\n", status.Name, status.State, status.RunningFor)
		} else {
			statusText = statusText + fmt.Sprintf(">`%s` - state: %s\n", status.Name, status.State)
		}
	}
	return b.Say(m.Channel, statusText)
}

func (j *Jobs) handleJobRun(b core.Bot, m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(core.LessMentions(m.Text))
	pieces := strings.Split(messageWithoutMentions, " ")
	if len(pieces) > 1 {
		jobName := pieces[len(pieces)-1]
		b.JobManager().RunJob(jobName)
		return b.Sayf(m.Channel, "ran job `%s`", jobName)
	}

	b.JobManager().RunAllJobs()
	return b.Say(m.Channel, "ran all jobs")
}

func (j *Jobs) handleJobCancel(b core.Bot, m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(core.LessMentions(m.Text))
	pieces := strings.Split(messageWithoutMentions, " ")
	if len(pieces) > 1 {
		taskName := pieces[len(pieces)-1]
		b.JobManager().CancelTask(taskName)
		return b.Sayf(m.Channel, "canceled task `%s`", taskName)
	}
	return exception.New("unhandled response.")
}

func (j *Jobs) handleJobEnable(b core.Bot, m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(core.LessMentions(m.Text))
	pieces := strings.Split(messageWithoutMentions, " ")
	if len(pieces) > 1 {
		taskName := pieces[len(pieces)-1]
		b.JobManager().EnableJob(taskName)
		return b.Sayf(m.Channel, "enabled job `%s`", taskName)
	}
	return exception.New("unhandled response.")
}

func (j *Jobs) handleJobDisable(b core.Bot, m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(core.LessMentions(m.Text))
	pieces := strings.Split(messageWithoutMentions, " ")
	if len(pieces) > 1 {
		taskName := pieces[len(pieces)-1]
		b.JobManager().DisableJob(taskName)
		return b.Sayf(m.Channel, "disabled job `%s`", taskName)
	}
	return exception.New("unhandled response.")
}
