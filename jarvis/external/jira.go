package external

import "github.com/blendlabs/go-request"

type JiraIssue struct {
	Id     string           `json:"id"`
	Expand string           `json:"expand"`
	Self   string           `json:"self"`
	Key    string           `json:"key"`
	Fields *JiraIssueFields `json:"fields"`
}

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

type JiraUser struct {
	Id           string            `json:"id"`
	Self         string            `json:"self"`
	Name         string            `json:"name"`
	Key          string            `json:"key"`
	EmailAddress string            `json:"emailAddress"`
	DisplayName  string            `json:"displayName"`
	Active       bool              `json:"active"`
	TimeZone     string            `json:"timeZone"`
	AvatarUrls   map[string]string `json:"avatarUrls"`
}

type JiraProject struct {
	Id         string            `json:"id"`
	Self       string            `json:"self"`
	Name       string            `json:"name"`
	Key        string            `json:"key"`
	AvatarUrls map[string]string `json:"avatarUrls"`
}

type JiraStatus struct {
	Id             string              `json:"id"`
	Self           string              `json:"self"`
	Description    string              `json:"description"`
	IconUrl        string              `json:"iconUrl"`
	Name           string              `json:"name"`
	StatusCategory *JiraStatusCategory `json:"statusCategory"`
}

type JiraStatusCategory struct {
	Id        int    `json:"id"`
	Self      string `json:"self"`
	Name      string `json:"name"`
	Key       string `json:"key"`
	ColorName string `json:"colorName"`
}

type JiraPriority struct {
	Id      string `json:"id"`
	Self    string `json:"self"`
	IconUrl string `json:"iconUrl"`
	Name    string `json:"name"`
}

func GetJiraIssue(user, password, host, issueId string) (*JiraIssue, error) {
	var issue JiraIssue
	fetchErr := request.NewRequest().AsGet().WithBasicAuth(user, password).WithScheme("https").WithHost(host).WithPath("rest/api/2/issue/%s", issueId).FetchJsonToObject(&issue)
	return &issue, fetchErr
}
