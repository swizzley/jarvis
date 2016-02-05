package jarvis

import (
	"fmt"
	"strings"
	"time"

	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis-cli/jarvis/external"
)

type JarvisAction struct {
	Expr    string
	Desc    string
	Handler JarvisMessageHandler
}

type JarvisMessageHandler func(m *slack.Message) error

func NewJarvisBot(token string) *JarvisBot {
	return &JarvisBot{Token: token, JobManager: chronometer.NewJobManager(), Configuration: map[string]interface{}{}, OptionPassive: false}
}

type JarvisBot struct {
	Token            string
	BotId            string
	OrganizationName string

	UsersLookup    map[string]slack.User
	ChannelsLookup map[string]slack.Channel

	OptionPassive bool

	Configuration map[string]interface{}

	JobManager *chronometer.JobManager
	Client     *slack.Client
}

func (jb *JarvisBot) Init() error {
	client := slack.Connect(jb.Token)
	jb.Client = client
	jb.Client.Listen(slack.EVENT_HELLO, func(m *slack.Message, c *slack.Client) {
		jb.Log("slack is connected")
	})
	jb.Client.Listen(slack.EVENT_MESSAGE, func(m *slack.Message, c *slack.Client) {
		resErr := jb.DoResponse(m)
		if resErr != nil {
			jb.Log(resErr)
		}
	})
	jb.JobManager.LoadJob(NewClock(jb))
	jb.JobManager.DisableJob("clock")
	return nil
}

func (jb *JarvisBot) Start() error {
	session, err := jb.Client.Start()
	if err != nil {
		return err
	}

	jb.BotId = session.Self.Id
	jb.OrganizationName = session.Team.Name
	jb.ChannelsLookup = jb.createChannelLookup(session)
	jb.UsersLookup = jb.createUsersLookup(session)
	jb.JobManager.Start()
	return nil
}

func (j *JarvisBot) MentionCommands() []JarvisAction {
	return []JarvisAction{
		JarvisAction{"^help", "Prints help info.", j.DoHelp},
		JarvisAction{"^time", "Prints the current time.", j.DoTime},
		JarvisAction{"^tell", "Tell people things.", j.DoTell},
		JarvisAction{"^channels", "Prints the channels I'm currently listening to.", j.DoChannels},
		JarvisAction{"^jobs", "Prints the current jobs and their statuses.", j.DoJobsStatus},
		JarvisAction{"^job:run", "Runs all jobs", j.DoJobsRun},
		JarvisAction{"^job:cancel", "Cancels a running job.", j.DoJobsCancel},
		JarvisAction{"^job:enable", "Enables a job.", j.DoJobEnable},
		JarvisAction{"^job:disable", "Disables enables a job.", j.DoJobDisable},
		JarvisAction{"^stock:price", "Fetches the current price and volume for a given ticker.", j.DoStockPrice},
		JarvisAction{"(.*)", "I'll do the best I can.", j.DoOtherResponse},
	}
}

func (j *JarvisBot) PassiveCommands() []JarvisAction {
	return []JarvisAction{
		JarvisAction{"(DSP-[0-9]+)", "Fetch jira task info.", j.DoJira},
		JarvisAction{"(BUGS-[0-9]+)", "Fetch jira task info.", j.DoJira},
		JarvisAction{"(.*)", "I'll do the best I can.", j.DoOtherPassiveResponse},
	}
}

func (jb *JarvisBot) DoResponse(m *slack.Message) error {
	jb.LogIncomingMessage(m)
	user := jb.FindUser(m.User)
	if user != nil {
		if m.User != "slackbot" && m.User != jb.BotId && !user.IsBot {
			messageText := util.TrimWhitespace(LessMentions(m.Text))
			if IsUserMention(m.Text, jb.BotId) || IsDM(m.Channel) {
				for _, actionHandler := range jb.MentionCommands() {
					if Like(messageText, actionHandler.Expr) {
						return actionHandler.Handler(m)
					}
				}
			} else {
				if jb.OptionPassive {
					for _, actionHandler := range jb.PassiveCommands() {
						if Like(messageText, actionHandler.Expr) {
							return actionHandler.Handler(m)
						}
					}
				}
			}
		}
	}
	return nil
}

func (jb *JarvisBot) DoHelp(m *slack.Message) error {
	responseText := "Here are the commands that are currently configured:"
	for _, actionHandler := range jb.MentionCommands() {
		responseText = responseText + fmt.Sprintf("\n>`%s` - %s", actionHandler.Expr, actionHandler.Desc)
	}
	responseText = responseText + "\nWith the following passive commands:"
	for _, actionHandler := range jb.PassiveCommands() {
		responseText = responseText + fmt.Sprintf("\n>`%s` - %s", actionHandler.Expr, actionHandler.Desc)
	}
	return jb.Say(m.Channel, responseText)
}

func (jb *JarvisBot) DoTime(m *slack.Message) error {
	now := time.Now().UTC()
	return jb.AnnounceTime(m.Channel, now)
}

func (jb *JarvisBot) DoTell(m *slack.Message) error {
	messageText := LessSpecificMention(m.Text, jb.BotId)
	words := strings.Split(messageText, " ")

	destinationUser := ""
	tellMessage := ""

	for x := 0; x < len(words); x++ {
		word := words[x]
		if Like(word, "tell") {
			continue
		} else if IsMention(word) {
			destinationUser = word
			tellMessage = strings.Join(words[x+1:], " ")
		}
	}
	tellMessage = ReplaceAny(tellMessage, []string{"shes", "she's", "she is", "hes", "he's", "he is", "theyre", "they're", "they are"}, "you are")
	resultMessage := fmt.Sprintf("%s %s", destinationUser, tellMessage)
	return jb.Say(m.Channel, resultMessage)
}

func (jb *JarvisBot) DoChannels(m *slack.Message) error {
	if len(jb.Client.ActiveChannels) == 0 {
		return jb.Say(m.Channel, "currently listening to *no* channels.")
	}
	activeChannelsText := "currently listening to the following channels:\n"
	for _, channelId := range jb.Client.ActiveChannels {
		if channel := jb.FindChannel(channelId); channel != nil {
			activeChannelsText = activeChannelsText + fmt.Sprintf(">#%s (id:%s)\n", channel.Name, channel.Id)
		}
	}
	return jb.Say(m.Channel, activeChannelsText)
}

func (jb *JarvisBot) DoJobsStatus(m *slack.Message) error {
	statusText := "current job statuses:\n"
	for _, status := range jb.JobManager.Status() {
		if len(status.RunningFor) != 0 {
			statusText = statusText + fmt.Sprintf(">`%s` - state: %s running for: %s\n", status.Name, status.State, status.RunningFor)
		} else {
			statusText = statusText + fmt.Sprintf(">`%s` - state: %s\n", status.Name, status.State)
		}
	}
	return jb.Say(m.Channel, statusText)
}

func (jb *JarvisBot) DoJobsRun(m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(LessMentions(m.Text))
	pieces := strings.Split(messageWithoutMentions, " ")
	if len(pieces) > 1 {
		jobName := pieces[len(pieces)-1]
		jb.JobManager.RunJob(jobName)
		return jb.Sayf(m.Channel, "ran job `%s`", jobName)
	} else {
		jb.JobManager.RunAllJobs()
		return jb.Say(m.Channel, "ran all jobs")
	}
}

func (jb *JarvisBot) DoJobsCancel(m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(LessMentions(m.Text))
	pieces := strings.Split(messageWithoutMentions, " ")
	if len(pieces) > 1 {
		taskName := pieces[len(pieces)-1]
		jb.JobManager.CancelTask(taskName)
		return jb.Sayf(m.Channel, "canceled task `%s`", taskName)
	}
	return jb.DoUnknown(m)
}

func (jb *JarvisBot) DoJobEnable(m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(LessMentions(m.Text))
	pieces := strings.Split(messageWithoutMentions, " ")
	if len(pieces) > 1 {
		taskName := pieces[len(pieces)-1]
		jb.JobManager.EnableJob(taskName)
		return jb.Sayf(m.Channel, "enabled job `%s`", taskName)
	}
	return jb.DoUnknown(m)
}

func (jb *JarvisBot) DoJobDisable(m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(LessMentions(m.Text))
	pieces := strings.Split(messageWithoutMentions, " ")
	if len(pieces) > 1 {
		taskName := pieces[len(pieces)-1]
		jb.JobManager.DisableJob(taskName)
		return jb.Sayf(m.Channel, "disabled job `%s`", taskName)
	}
	return jb.DoUnknown(m)
}

func (jb *JarvisBot) DoStockPrice(m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(LessMentions(m.Text))
	pieces := strings.Split(messageWithoutMentions, " ")
	if len(pieces) > 1 {
		rawTicker := pieces[len(pieces)-1]
		tickers := []string{}
		if strings.Contains(rawTicker, ",") {
			tickers = strings.Split(rawTicker, ",")
		} else {
			tickers = []string{rawTicker}
		}
		stockInfo, stockErr := external.StockPrice(tickers, external.STOCK_DEFAULT_FORMAT)
		if stockErr != nil {
			return stockErr
		}
		return jb.AnnounceStocks(m.Channel, stockInfo)
	}
	return jb.DoUnknown(m)
}

func (jb *JarvisBot) DoJira(m *slack.Message) error {
	text := LessMentions(m.Text)

	issueIds := jb.extractJiraIssues(text)
	if len(issueIds) == 0 {
		return nil
	}

	issues, issuesErr := jb.fetchJiraIssues(issueIds)
	if issuesErr != nil {
		return issuesErr
	}
	if len(issues) == 0 {
		return nil
	}

	user := jb.FindUser(m.User)

	leadText := fmt.Sprintf("*%s* has mentioned the following jira issues (%d): ", user.Profile.FirstName, len(issues))
	message := slack.NewChatMessage(m.Channel, leadText)
	message.AsUser = slack.OptionalBool(true)
	message.UnfurlLinks = slack.OptionalBool(false)
	message.UnfurlMedia = slack.OptionalBool(false)
	for _, issue := range issues {
		itemText := fmt.Sprintf("%s - %s\n%s", issue.Key, issue.Fields.Summary, issue.Self)
		item := slack.ChatMessageAttachment{
			Fallback: itemText,
			Color:    slack.OptionalString("#3572b0"),
			Text:     slack.OptionalString(itemText),
		}
		message.Attachments = append(message.Attachments, item)
	}

	_, messageErr := jb.Client.ChatPostMessage(message)
	if messageErr != nil {
		fmt.Printf("issue posting message: %v\n", messageErr)
	}
	return messageErr
}

func (jb *JarvisBot) extractJiraIssues(text string) []string {
	issueIds := []string{}
	issueIds = append(issueIds, Extract(text, "(DSP-[0-9]+)")...)
	issueIds = append(issueIds, Extract(text, "(BUGS-[0-9]+)")...)
	return issueIds
}

func (jb *JarvisBot) fetchJiraIssues(issueIds []string) ([]*external.JiraIssue, error) {
	issues := []*external.JiraIssue{}
	rawCredentials, hasCredentials := jb.Configuration["JIRA_CREDENTIALS"]

	if !hasCredentials {
		return issues, exception.New("Jarvis is not configured with Jira credentials.")
	}
	credentials, isString := rawCredentials.(string)
	if !isString {
		return issues, exception.New("Jira credentials are not a string.")
	}

	credentialPieces := strings.Split(credentials, ":")

	if len(credentialPieces) != 2 {
		return issues, exception.New("Jira credentials are not formatted correctly.")
	}

	jiraUser := credentialPieces[0]
	jiraPassword := credentialPieces[1]

	jiraHost, hasJiraHost := jb.Configuration["JIRA_HOST"]
	if !hasJiraHost {
		return issues, exception.New("Jarvis is not configured with a Jira host.")
	}

	host, hostIsString := jiraHost.(string)
	if !hostIsString {
		return issues, exception.New("Jira host is not a string.")
	}

	for _, issueId := range issueIds {
		issue, issueErr := external.GetJiraIssue(jiraUser, jiraPassword, host, issueId)
		if issueErr == nil {
			issues = append(issues, issue)
		} else {
			return issues, issueErr
		}
	}

	return issues, nil
}

func (jb *JarvisBot) DoOtherResponse(m *slack.Message) error {
	message := util.TrimWhitespace(LessMentions(m.Text))
	if IsSalutation(message) {
		return jb.DoSalutation(m)
	} else {
		return jb.DoUnknown(m)
	}
}

func (jb *JarvisBot) DoOtherPassiveResponse(m *slack.Message) error {
	message := util.TrimWhitespace(LessMentions(m.Text))
	if IsAngry(message) {
		user := jb.FindUser(m.User)
		response := []string{"slow down %s", "maybe calm down %s", "%s you should really relax", "chill %s", "it's ok %s, let it out"}
		return jb.Sayf(m.Channel, Random(response), strings.ToLower(user.Profile.FirstName))
	}
	return nil
}

func (jb *JarvisBot) DoSalutation(m *slack.Message) error {
	user := jb.FindUser(m.User)
	salutation := []string{"hey %s", "hi %s", "hello %s", "ohayo gozaimasu %s", "salut %s", "bonjour %s", "yo %s", "sup %s"}
	return jb.Sayf(m.Channel, Random(salutation), strings.ToLower(user.Profile.FirstName))
}

func (jb *JarvisBot) DoUnknown(m *slack.Message) error {
	return jb.Sayf(m.Channel, "I don't know how to respond to this\n>%s", m.Text)
}

func (jb *JarvisBot) AnnounceStocks(destinationId string, stockInfo []external.StockInfo) error {
	tickersLabels := []string{}
	for _, stock := range stockInfo {
		tickersLabels = append(tickersLabels, fmt.Sprintf("`%s`", stock.Ticker))
	}
	tickersLabel := strings.Join(tickersLabels, " ")
	stockText := fmt.Sprintf("current equity price info for %s\n", tickersLabel)
	for _, stock := range stockInfo {
		if stock.Values != nil && len(stock.Values) > 3 {
			if floatValue, isFloat := stock.Values[2].(float64); isFloat {
				change := floatValue
				changeText := fmt.Sprintf("%.2f", change)
				changePct := stock.Values[3]
				stockText = stockText + fmt.Sprintf("> `%s` - last: *%.2f* vol: *%d* ch: *%s* *%s*\n", stock.Ticker, stock.Values[0], int(stock.Values[1].(float64)), changeText, util.StripQuotes(changePct.(string)))
			} else {
				return jb.Sayf(destinationId, "There was an issue with `%s`", stock.Ticker)
			}
		}
	}
	return jb.Say(destinationId, stockText)
}

func (jb *JarvisBot) AnnounceTime(destinationId string, currentTime time.Time) error {
	timeText := fmt.Sprintf("%s UTC", currentTime.Format(time.Kitchen))
	message := slack.NewChatMessage(destinationId, "")
	message.AsUser = slack.OptionalBool(true)
	message.UnfurlLinks = slack.OptionalBool(false)
	message.UnfurlMedia = slack.OptionalBool(false)
	message.Attachments = []slack.ChatMessageAttachment{
		slack.ChatMessageAttachment{
			Fallback: fmt.Sprintf("The time is now:\n>%s", timeText),
			Color:    slack.OptionalString("#4099FF"),
			Pretext:  slack.OptionalString("The time is now:"),
			Text:     slack.OptionalString(timeText),
		},
	}

	_, messageErr := jb.Client.ChatPostMessage(message)
	if messageErr != nil {
		fmt.Printf("issue posting message: %v\n", messageErr)
	}
	return messageErr
}

func (jb *JarvisBot) FindUser(userId string) *slack.User {
	if user, hasUser := jb.UsersLookup[userId]; hasUser {
		return &user
	}
	return nil
}

func (jb *JarvisBot) FindChannel(channelId string) *slack.Channel {
	if channel, hasChannel := jb.ChannelsLookup[channelId]; hasChannel {
		return &channel
	}
	return nil
}

func (jb *JarvisBot) createUsersLookup(session *slack.Session) map[string]slack.User {
	lookup := map[string]slack.User{}
	for x := 0; x < len(session.Users); x++ {
		user := session.Users[x]
		lookup[user.Id] = user
	}
	return lookup
}

func (jb *JarvisBot) createChannelLookup(session *slack.Session) map[string]slack.Channel {
	lookup := map[string]slack.Channel{}
	for x := 0; x < len(session.Channels); x++ {
		channel := session.Channels[x]
		lookup[channel.Id] = channel
	}
	return lookup
}

func (jb *JarvisBot) Say(destinationId string, components ...interface{}) error {
	jb.LogOutgoingMessage(destinationId, components...)
	return jb.Client.Say(destinationId, components...)
}

func (jb *JarvisBot) Sayf(destinationId string, format string, components ...interface{}) error {
	message := fmt.Sprintf(format, components...)
	jb.LogOutgoingMessage(destinationId, message)
	return jb.Client.Sayf(destinationId, format, components...)
}

func (jb *JarvisBot) LogIncomingMessage(m *slack.Message) {
	user := jb.FindUser(m.User)
	channel := jb.FindChannel(m.Channel)

	userName := "system"
	if user != nil {
		userName = user.Name
	}

	if channel != nil {
		jb.Logf("=> #%s (%s) - %s: %s", channel.Name, channel.Id, userName, m.Text)
	} else {
		jb.Logf("=> PM - %s: %s", userName, m.Text)
	}
}

func (jb *JarvisBot) LogOutgoingMessage(destinationId string, components ...interface{}) {
	if Like(destinationId, "^C") {
		channel := jb.FindChannel(destinationId)
		jb.Logf("<= #%s (%s) - jarvis: %s", channel.Name, channel.Id, fmt.Sprint(components...))
	} else {
		jb.Logf("<= PM - jarvis: %s", fmt.Sprint(components...))
	}
}

func (jb *JarvisBot) Log(components ...interface{}) {
	message := fmt.Sprint(components...)
	fmt.Printf("%s - %s - %s\n", jb.OrganizationName, time.Now().UTC().Format(time.RFC3339), message)
}

func (jb *JarvisBot) Logf(format string, components ...interface{}) {
	message := fmt.Sprintf(format, components...)
	fmt.Printf("%s - %s - %s\n", jb.OrganizationName, time.Now().UTC().Format(time.RFC3339), message)
}
