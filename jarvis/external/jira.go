package external

import (
	"encoding/json"
	"io/ioutil"

	"github.com/blendlabs/go-exception"
	"github.com/wcharczuk/jarvis/jarvis/core"
)

// JiraIssue represents JIRA metadata.
type JiraIssue struct {
	ID     string           `json:"id"`
	Expand string           `json:"expand"`
	Self   string           `json:"self"`
	Key    string           `json:"key"`
	Fields *JiraIssueFields `json:"fields"`
}

// JiraIssueFields represents JIRA metadata.
type JiraIssueFields struct {
	LastViewed  string       `json:"lastViewed"`
	Created     string       `json:"created"`
	Updated     string       `json:"updated"`
	Summary     string       `json:"summary"`
	Description string       `json:"description"`
	Labels      []string     `json:"labels"`
	Project     *JiraProject `json:"project"`
	Assignee    *JiraUser    `json:"assignee"`
	Creator     *JiraUser    `json:"creator"`
	Status      *JiraStatus  `json:"status"`
}

// JiraUser represents JIRA metadata.
type JiraUser struct {
	ID           string            `json:"id"`
	Self         string            `json:"self"`
	Name         string            `json:"name"`
	Key          string            `json:"key"`
	EmailAddress string            `json:"emailAddress"`
	DisplayName  string            `json:"displayName"`
	Active       bool              `json:"active"`
	TimeZone     string            `json:"timeZone"`
	AvatarUrls   map[string]string `json:"avatarUrls"`
}

// JiraProject represents JIRA metadata.
type JiraProject struct {
	ID         string            `json:"id"`
	Self       string            `json:"self"`
	Name       string            `json:"name"`
	Key        string            `json:"key"`
	AvatarUrls map[string]string `json:"avatarUrls"`
}

// JiraStatus represents JIRA metadata.
type JiraStatus struct {
	ID             string              `json:"id"`
	Self           string              `json:"self"`
	Description    string              `json:"description"`
	IconURL        string              `json:"iconUrl"`
	Name           string              `json:"name"`
	StatusCategory *JiraStatusCategory `json:"statusCategory"`
}

// JiraStatusCategory represents JIRA metadata.
type JiraStatusCategory struct {
	ID        int    `json:"id"`
	Self      string `json:"self"`
	Name      string `json:"name"`
	Key       string `json:"key"`
	ColorName string `json:"colorName"`
}

// JiraPriority represents JIRA metadata.
type JiraPriority struct {
	ID      string `json:"id"`
	Self    string `json:"self"`
	IconURL string `json:"iconUrl"`
	Name    string `json:"name"`
}

type JiraError struct {
	ErrorMessages []string
}

// GetJiraIssue gets the metadata for a given issueID.
func GetJiraIssue(user, password, host, issueID string) (*JiraIssue, error) {
	res, err := core.NewExternalRequest().AsGet().WithBasicAuth(user, password).WithScheme("https").WithHost(host).WithPathf("rest/api/2/issue/%s", issueID).FetchRawResponse()
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	var je JiraError
	err = json.Unmarshal(body, &je)
	if err == nil && len(je.ErrorMessages) != 0 {
		return nil, exception.Newf("Errors returned from jira: %s\n", je.ErrorMessages[0])
	}

	var ji JiraIssue
	err = json.Unmarshal(body, &ji)
	return &ji, err
}
