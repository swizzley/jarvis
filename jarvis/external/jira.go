package external

import "github.com/wcharczuk/jarvis/jarvis/core"

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

// GetJiraIssue gets the metadata for a given issueID.
func GetJiraIssue(user, password, host, issueID string) (*JiraIssue, error) {
	var issue JiraIssue
	fetchErr := core.NewExternalRequest().AsGet().WithBasicAuth(user, password).WithScheme("https").WithHost(host).WithPathf("rest/api/2/issue/%s", issueID).FetchJSONToObject(&issue)
	return &issue, fetchErr
}
