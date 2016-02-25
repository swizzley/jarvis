package jarvis

import (
	"fmt"
	"strings"
	"time"

	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
	"github.com/blendlabs/go-util/linq"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/jarvis/external"
)

// NewBot returns a new Bot instance.
func NewBot(token string) *Bot {
	return &Bot{Token: token, JobManager: chronometer.NewJobManager(), Configuration: map[string]interface{}{}, OptionPassive: true}
}

// Bot is the main primitive.
type Bot struct {
	Token            string
	BotID            string
	OrganizationName string

	UsersLookup    map[string]slack.User
	ChannelsLookup map[string]slack.Channel

	OptionPassive bool

	Configuration map[string]interface{}

	JobManager *chronometer.JobManager
	Client     *slack.Client
}

// Init connects the bot to Slack.
func (b *Bot) Init() error {
	client := slack.Connect(b.Token)
	b.Client = client
	b.Client.Listen(slack.EventHello, func(m *slack.Message, c *slack.Client) {
		b.Log("slack is connected")
	})
	b.Client.Listen(slack.EventMessage, func(m *slack.Message, c *slack.Client) {
		resErr := b.DoResponse(m)
		if resErr != nil {
			b.Log(resErr)
		}
	})
	b.JobManager.LoadJob(jobs.NewClock(b))
	b.JobManager.DisableJob("clock")
	return nil
}

func (b *Bot) Start() error {
	session, err := b.Client.Start()
	if err != nil {
		return err
	}

	b.BotId = session.Self.Id
	b.OrganizationName = session.Team.Name
	b.ChannelsLookup = b.createChannelLookup(session)
	b.UsersLookup = b.createUsersLookup(session)
	b.JobManager.Start()
	return nil
}

func (b *Bot) MentionCommands() []JarvisAction {
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
		JarvisAction{"^config:passive", "Enables or disables passive commands.", j.DoConfigPassive},
		JarvisAction{"^config", "Prints the current config", j.DoConfig},
		JarvisAction{"(.*)", "I'll do the best I can.", j.DoOtherResponse},
	}
}

func (b *Bot) PassiveCommands() []JarvisAction {
	return []JarvisAction{
		JarvisAction{"(DSP-[0-9]+)", "Fetch jira task info.", j.DoJira},
		JarvisAction{"(BUGS-[0-9]+)", "Fetch jira task info.", j.DoJira},
		JarvisAction{"(.*)", "I'll do the best I can.", j.DoOtherPassiveResponse},
	}
}

func (b *Bot) DoResponse(m *slack.Message) error {
	b.LogIncomingMessage(m)
	user := b.FindUser(m.User)
	if user != nil {
		if m.User != "slackbot" && m.User != b.BotId && !user.IsBot {
			messageText := util.TrimWhitespace(LessMentions(m.Text))
			if IsUserMention(m.Text, b.BotId) || IsDM(m.Channel) {
				for _, actionHandler := range b.MentionCommands() {
					if Like(messageText, actionHandler.Expr) {
						return actionHandler.Handler(m)
					}
				}
			} else {
				if b.OptionPassive {
					for _, actionHandler := range b.PassiveCommands() {
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

func (b *Bot) DoHelp(m *slack.Message) error {
	responseText := "Here are the commands that are currently configured:"
	for _, actionHandler := range b.MentionCommands() {
		responseText = responseText + fmt.Sprintf("\n>`%s` - %s", actionHandler.Expr, actionHandler.Desc)
	}
	responseText = responseText + "\nWith the following passive commands:"
	for _, actionHandler := range b.PassiveCommands() {
		responseText = responseText + fmt.Sprintf("\n>`%s` - %s", actionHandler.Expr, actionHandler.Desc)
	}
	return b.Say(m.Channel, responseText)
}

func (b *Bot) DoTime(m *slack.Message) error {
	now := time.Now().UTC()
	return b.AnnounceTime(m.Channel, now)
}

func (b *Bot) DoTell(m *slack.Message) error {
	messageText := LessSpecificMention(m.Text, b.BotId)
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
	tellMessage = ReplaceAny(tellMessage, "you are", "shes", "she's", "she is", "hes", "he's", "he is", "theyre", "they're", "they are")
	resultMessage := fmt.Sprintf("%s %s", destinationUser, tellMessage)
	return b.Say(m.Channel, resultMessage)
}

func (b *Bot) DoChannels(m *slack.Message) error {
	if len(b.Client.ActiveChannels) == 0 {
		return b.Say(m.Channel, "currently listening to *no* channels.")
	}
	activeChannelsText := "currently listening to the following channels:\n"
	for _, channelId := range b.Client.ActiveChannels {
		if channel := b.FindChannel(channelId); channel != nil {
			activeChannelsText = activeChannelsText + fmt.Sprintf(">#%s (id:%s)\n", channel.Name, channel.Id)
		}
	}
	return b.Say(m.Channel, activeChannelsText)
}

func (b *Bot) DoJobsStatus(m *slack.Message) error {
	statusText := "current job statuses:\n"
	for _, status := range b.JobManager.Status() {
		if len(status.RunningFor) != 0 {
			statusText = statusText + fmt.Sprintf(">`%s` - state: %s running for: %s\n", status.Name, status.State, status.RunningFor)
		} else {
			statusText = statusText + fmt.Sprintf(">`%s` - state: %s\n", status.Name, status.State)
		}
	}
	return b.Say(m.Channel, statusText)
}

func (b *Bot) DoJobsRun(m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(LessMentions(m.Text))
	pieces := strings.Split(messageWithoutMentions, " ")
	if len(pieces) > 1 {
		jobName := pieces[len(pieces)-1]
		b.JobManager.RunJob(jobName)
		return b.Sayf(m.Channel, "ran job `%s`", jobName)
	} else {
		b.JobManager.RunAllJobs()
		return b.Say(m.Channel, "ran all jobs")
	}
}

func (b *Bot) DoJobsCancel(m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(LessMentions(m.Text))
	pieces := strings.Split(messageWithoutMentions, " ")
	if len(pieces) > 1 {
		taskName := pieces[len(pieces)-1]
		b.JobManager.CancelTask(taskName)
		return b.Sayf(m.Channel, "canceled task `%s`", taskName)
	}
	return b.DoUnknown(m)
}

func (b *Bot) DoJobEnable(m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(LessMentions(m.Text))
	pieces := strings.Split(messageWithoutMentions, " ")
	if len(pieces) > 1 {
		taskName := pieces[len(pieces)-1]
		b.JobManager.EnableJob(taskName)
		return b.Sayf(m.Channel, "enabled job `%s`", taskName)
	}
	return b.DoUnknown(m)
}

func (b *Bot) DoJobDisable(m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(LessMentions(m.Text))
	pieces := strings.Split(messageWithoutMentions, " ")
	if len(pieces) > 1 {
		taskName := pieces[len(pieces)-1]
		b.JobManager.DisableJob(taskName)
		return b.Sayf(m.Channel, "disabled job `%s`", taskName)
	}
	return b.DoUnknown(m)
}

func (b *Bot) DoConfig(m *slack.Message) error {
	configText := "current config:\n"
	if b.OptionPassive {
		configText = configText + "> `passive` : enabled\n"
	} else {
		configText = configText + "> `passive` : disabled\n"
	}

	return b.Say(m.Channel, configText)
}

func (b *Bot) DoConfigPassive(m *slack.Message) error {
	messageWithoutMentions := util.TrimWhitespace(LessMentions(m.Text))
	passiveValue := linq.LastOfString(strings.Split(messageWithoutMentions, " "), nil)

	if passiveValue != nil {
		passiveSetting := false
		if LikeAny(*passiveValue, "true", "yes", "on", "1") {
			passiveSetting = true
		} else if LikeAny(*passiveValue, "false", "off", "0") {
			passiveSetting = false
		} else {
			return b.Sayf(m.Channel, "invalid %T option value: %q", passiveSetting, *passiveValue)
		}
		b.OptionPassive = passiveSetting
		if passiveSetting {
			return b.Say(m.Channel, "config: enabled passive responses")
		} else {
			return b.Say(m.Channel, "config: disabled passive responses")
		}
	}
	return b.DoUnknown(m)
}

func (b *Bot) DoStockPrice(m *slack.Message) error {
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
		return b.AnnounceStocks(m.Channel, stockInfo)
	}
	return b.DoUnknown(m)
}

func (b *Bot) DoJira(m *slack.Message) error {
	text := LessMentions(m.Text)

	issueIds := b.extractJiraIssues(text)
	if len(issueIds) == 0 {
		return nil
	}

	issues, issuesErr := b.fetchJiraIssues(issueIds)
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

	_, messageErr := b.Client.ChatPostMessage(message)
	if messageErr != nil {
		fmt.Printf("issue posting message: %v\n", messageErr)
	}
	return messageErr
}

func (b *Bot) extractJiraIssues(text string) []string {
	issueIds := []string{}
	issueIds = append(issueIds, Extract(text, "(DSP-[0-9]+)")...)
	issueIds = append(issueIds, Extract(text, "(BUGS-[0-9]+)")...)
	return issueIds
}

func (b *Bot) fetchJiraIssues(issueIds []string) ([]*external.JiraIssue, error) {
	issues := []*external.JiraIssue{}
	rawCredentials, hasCredentials := b.Configuration["JIRA_CREDENTIALS"]

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

	jiraHost, hasJiraHost := b.Configuration["JIRA_HOST"]
	if !hasJiraHost {
		return issues, exception.New("Jarvis is not configured with a Jira host.")
	}

	host, hostIsString := jiraHost.(string)
	if !hostIsString {
		return issues, exception.New("Jira host is not a string.")
	}

	for _, issueID := range issueIds {
		issue, issueErr := external.GetJiraIssue(jiraUser, jiraPassword, host, issueID)
		if issueErr == nil {
			issues = append(issues, issue)
		} else {
			return issues, issueErr
		}
	}

	return issues, nil
}

func (b *Bot) DoOtherResponse(m *slack.Message) error {
	message := util.TrimWhitespace(LessMentions(m.Text))
	if IsSalutation(message) {
		return b.DoSalutation(m)
	} else {
		return b.DoUnknown(m)
	}
}

func (b *Bot) DoOtherPassiveResponse(m *slack.Message) error {
	message := util.TrimWhitespace(LessMentions(m.Text))
	if IsAngry(message) {
		user := b.FindUser(m.User)
		response := []string{"slow down %s", "maybe calm down %s", "%s you should really relax", "chill %s", "it's ok %s, let it out"}
		return b.Sayf(m.Channel, Random(response), strings.ToLower(user.Profile.FirstName))
	}
	return nil
}

func (b *Bot) DoSalutation(m *slack.Message) error {
	user := b.FindUser(m.User)
	salutation := []string{"hey %s", "hi %s", "hello %s", "ohayo gozaimasu %s", "salut %s", "bonjour %s", "yo %s", "sup %s"}
	return b.Sayf(m.Channel, Random(salutation), strings.ToLower(user.Profile.FirstName))
}

func (b *Bot) DoUnknown(m *slack.Message) error {
	return b.Sayf(m.Channel, "I don't know how to respond to this\n>%s", m.Text)
}

func (b *Bot) AnnounceStocks(destinationId string, stockInfo []external.StockInfo) error {
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
				return b.Sayf(destinationId, "There was an issue with `%s`", stock.Ticker)
			}
		}
	}
	return b.Say(destinationId, stockText)
}

func (b *Bot) AnnounceTime(destinationId string, currentTime time.Time) error {
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

	_, messageErr := b.Client.ChatPostMessage(message)
	if messageErr != nil {
		fmt.Printf("issue posting message: %v\n", messageErr)
	}
	return messageErr
}

func (b *Bot) FindUser(userID string) *slack.User {
	if user, hasUser := b.UsersLookup[userID]; hasUser {
		return &user
	}
	return nil
}

func (b *Bot) FindChannel(channelId string) *slack.Channel {
	if channel, hasChannel := b.ChannelsLookup[channelId]; hasChannel {
		return &channel
	}
	return nil
}

func (b *Bot) createUsersLookup(session *slack.Session) map[string]slack.User {
	lookup := map[string]slack.User{}
	for x := 0; x < len(session.Users); x++ {
		user := session.Users[x]
		lookup[user.Id] = user
	}
	return lookup
}

func (b *Bot) createChannelLookup(session *slack.Session) map[string]slack.Channel {
	lookup := map[string]slack.Channel{}
	for x := 0; x < len(session.Channels); x++ {
		channel := session.Channels[x]
		lookup[channel.Id] = channel
	}
	return lookup
}

func (b *Bot) Say(destinationID string, components ...interface{}) error {
	b.LogOutgoingMessage(destinationID, components...)
	return b.Client.Say(destinationID, components...)
}

func (b *Bot) Sayf(destinationID string, format string, components ...interface{}) error {
	message := fmt.Sprintf(format, components...)
	b.LogOutgoingMessage(destinationID, message)
	return b.Client.Sayf(destinationID, format, components...)
}

func (b *Bot) LogIncomingMessage(m *slack.Message) {
	user := b.FindUser(m.User)
	channel := b.FindChannel(m.Channel)

	userName := "system"
	if user != nil {
		userName = user.Name
	}

	if channel != nil {
		b.Logf("=> #%s (%s) - %s: %s", channel.Name, channel.Id, userName, m.Text)
	} else {
		b.Logf("=> PM - %s: %s", userName, m.Text)
	}
}

func (b *Bot) LogOutgoingMessage(destinationId string, components ...interface{}) {
	if Like(destinationId, "^C") {
		channel := b.FindChannel(destinationId)
		b.Logf("<= #%s (%s) - jarvis: %s", channel.Name, channel.Id, fmt.Sprint(components...))
	} else {
		b.Logf("<= PM - jarvis: %s", fmt.Sprint(components...))
	}
}

func (b *Bot) Log(components ...interface{}) {
	message := fmt.Sprint(components...)
	fmt.Printf("%s - %s - %s\n", b.OrganizationName, time.Now().UTC().Format(time.RFC3339), message)
}

func (b *Bot) Logf(format string, components ...interface{}) {
	message := fmt.Sprintf(format, components...)
	fmt.Printf("%s - %s - %s\n", b.OrganizationName, time.Now().UTC().Format(time.RFC3339), message)
}
