package modules

import (
	"fmt"
	"os"
	"strings"

	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/jarvis/core"
	"github.com/wcharczuk/jarvis/jarvis/external"
)

const (
	// EnvironmentJiraCredentials is the environment variable name for jira credentials.
	EnvironmentJiraCredentials = "JIRA_CREDENTIALS"

	// EnvironmentJiraHost is the environment variable name for jira credentials.
	EnvironmentJiraHost = "JIRA_HOST"

	// ConfigJiraCredentials is the jira credentials bot config entry.
	ConfigJiraCredentials = "jira_credentials"

	// ConfigJiraHost is the jira host bot config entry.
	ConfigJiraHost = "jira_host"

	// ModuleJira is the name of the jira module.
	ModuleJira = "jira"

	// ActionJiraDSP is the name of the DSP action.
	ActionJiraDSP = "jira.dsp"

	// ActionJiraBUGS is the name of the bugs action.
	ActionJiraBUGS = "jira.bugs"

	// ActionJiraIMP is the name of the bugs action.
	ActionJiraIMP = "jira.imp"

	// ActionJiraREL is the name of the bugs action.
	ActionJiraREL = "jira.rel"
)

// Jira is the jira module.
type Jira struct{}

// Init for this module does nothing.
func (j *Jira) Init(b core.Bot) error {
	if _, hasEntry := b.Configuration()[ConfigJiraCredentials]; !hasEntry {
		envCredentials := os.Getenv(EnvironmentJiraCredentials)
		if len(envCredentials) != 0 {
			b.Configuration()[ConfigJiraCredentials] = envCredentials
		} else {
			return exception.Newf("No `%s` provided, module `%s` cannot load.", EnvironmentJiraCredentials, ModuleJira)
		}
	}

	if _, hasEntry := b.Configuration()[ConfigJiraHost]; !hasEntry {
		envHost := os.Getenv(EnvironmentJiraHost)
		if len(envHost) != 0 {
			b.Configuration()[ConfigJiraHost] = envHost
		} else {
			return exception.Newf("No `%s` provided, module `%s` cannot load.", EnvironmentJiraHost, ModuleJira)
		}
	}

	return nil
}

// Name returns the name of the module.
func (j *Jira) Name() string {
	return ModuleJira
}

// Actions returns the action for the module.
func (j *Jira) Actions() []core.Action {
	return []core.Action{
		core.Action{ID: ActionJiraDSP, Passive: true, MessagePattern: "(DSP-[0-9]+)", Description: "Fetch jira DSP task info.", Handler: j.handleJira},
		core.Action{ID: ActionJiraBUGS, Passive: true, MessagePattern: "(BUGS-[0-9]+)", Description: "Fetch jira BUGS task info.", Handler: j.handleJira},
		core.Action{ID: ActionJiraIMP, Passive: true, MessagePattern: "(IMP-[0-9]+)", Description: "Fetch jira IMP task info.", Handler: j.handleJira},
		core.Action{ID: ActionJiraREL, Passive: true, MessagePattern: "(REL-[0-9]+)", Description: "Fetch jira REL task info.", Handler: j.handleJira},
	}
}

func (j *Jira) handleJira(b core.Bot, m *slack.Message) error {
	text := core.LessMentions(m.Text)

	issueIds := j.extractJiraIssues(text)
	if len(issueIds) == 0 {
		return nil
	}

	issues, issuesErr := j.fetchJiraIssues(b, issueIds)
	if issuesErr != nil {
		return issuesErr
	}
	if len(issues) == 0 {
		return nil
	}

	user := b.FindUser(m.User)

	leadText := fmt.Sprintf("*%s* has mentioned the following jira issues (%d): ", user.Profile.FirstName, len(issues))
	message := slack.NewChatMessage(m.Channel, leadText)
	message.AsUser = slack.OptionalBool(true)
	message.UnfurlLinks = slack.OptionalBool(false)
	for _, issue := range issues {
		if !util.IsEmpty(issue.Key) {
			var itemText string
			if issue.Fields != nil {
				assignee := "Unassigned"
				if issue.Fields.Assignee != nil {
					assignee = issue.Fields.Assignee.DisplayName
				}
				itemText = fmt.Sprintf("%s %s\nAssigned To: %s",
					fmt.Sprintf("https://%s/browse/%s", b.Configuration()[ConfigJiraHost], issue.Key),
					issue.Fields.Summary,
					assignee,
				)
			} else {
				itemText = fmt.Sprintf("%s\n%s", issue.Key, fmt.Sprintf("https://%s/browse/%s", b.Configuration()[ConfigJiraHost], issue.Key))
			}

			item := slack.ChatMessageAttachment{
				Color: slack.OptionalString("#3572b0"),
				Text:  slack.OptionalString(itemText),
			}
			message.Attachments = append(message.Attachments, item)
		}
	}

	_, messageErr := b.Client().ChatPostMessage(message)
	if messageErr != nil {
		fmt.Printf("issue posting message: %v\n", messageErr)
	}
	return messageErr
}

func (j *Jira) extractJiraIssues(text string) []string {
	issueIds := []string{}
	issueIds = append(issueIds, core.Extract(text, "(DSP-[0-9]+)")...)
	issueIds = append(issueIds, core.Extract(text, "(BUGS-[0-9]+)")...)
	issueIds = append(issueIds, core.Extract(text, "(IMP-[0-9]+)")...)
	issueIds = append(issueIds, core.Extract(text, "(REL-[0-9]+)")...)
	return issueIds
}

func (j *Jira) fetchJiraIssues(b core.Bot, issueIds []string) ([]*external.JiraIssue, error) {
	issues := []*external.JiraIssue{}
	credentials, hasCredentials := b.Configuration()[ConfigJiraCredentials]

	if !hasCredentials {
		return issues, exception.New("Jarvis is not configured with Jira credentials.")
	}

	credentialPieces := strings.Split(credentials, ":")

	if len(credentialPieces) != 2 {
		return issues, exception.New("Jira credentials are not formatted correctly.")
	}

	jiraUser := credentialPieces[0]
	jiraPassword := credentialPieces[1]

	jiraHost, hasJiraHost := b.Configuration()[ConfigJiraHost]
	if !hasJiraHost {
		return issues, exception.New("Jarvis is not configured with a Jira host.")
	}

	for _, issueID := range issueIds {
		issue, issueErr := external.GetJiraIssue(jiraUser, jiraPassword, jiraHost, issueID)
		if issueErr == nil {
			issues = append(issues, issue)
		} else {
			return issues, issueErr
		}
	}

	return issues, nil
}
