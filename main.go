package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/go-slack"
	"github.com/wcharczuk/jarvis/lib"
	"github.com/wcharczuk/jarvis/lib/jobs"
)

var _usersLookup map[string]slack.User
var _channelsLookup map[string]slack.Channel
var _botId string

type messageHandler func(m *slack.Message, c *slack.Client) error

type action struct {
	expr    string
	handler messageHandler
}

var commands = []action{
	action{"^time", doTime},
	action{"^debug:channels", doDebugChannels},
	action{"^jobs", doJobsStatus},
	action{"^job:run", doJobsRun},
	action{"^job:cancel", doJobsCancel},
	action{"^stock:price", doStockPrice},
	action{"^stock:track", doStockTrack},
	action{"^stock:remove", doStockRemove},
	action{"^stocks:price", doStocksPrice},
	action{"^stocks", doStocks},
	action{"(.*)", doOtherResponse},
}

func TOKEN() string {
	return os.Getenv("SLACK_API_TOKEN")
}

func main() {
	client := slack.Connect(TOKEN())

	client.Listen(slack.EVENT_HELLO, func(m *slack.Message, c *slack.Client) {
		log("slack is connected")
	})

	client.Listen(slack.EVENT_MESSAGE, func(m *slack.Message, c *slack.Client) {
		resErr := doResponse(m, c)
		if resErr != nil {
			log(resErr)
		}
	})

	lib.DbInit()
	migrateErr := lib.MigrateModel()
	if migrateErr != nil {
		fmt.Printf("Error migrating model: %s\n", migrateErr)
		os.Exit(1)
	}

	chronometer.Default().LoadJob(jobs.NewStock(client))
	chronometer.Default().LoadJob(jobs.NewClock(client))
	chronometer.Default().Start()

	session, err := client.Start()
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	_botId = session.Self.Id
	_usersLookup = createUsersLookup(session)
	_channelsLookup = createChannelLookup(session)

	startStatusServer(client)
}

func startStatusServer(c *slack.Client) {
	http.HandleFunc("/", injectClient(c, statusHandler))
	logf("starting status server, listening on: %s", os.Getenv("PORT"))
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}

func injectClient(c *slack.Client, h clientAwareHttpHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h(c, w, r)
	}
}

type clientAwareHttpHandlerFunc func(c *slack.Client, w http.ResponseWriter, r *http.Request)

func statusHandler(c *slack.Client, w http.ResponseWriter, r *http.Request) {
	statusText := "Jarvis is running and listening to the following channels:\n"
	for _, channelId := range c.ActiveChannels {
		channel := findChannel(channelId)
		statusText = statusText + fmt.Sprintf("> #%s (%s)\n", channel.Name, channel.Id)
	}
	fmt.Fprintf(w, statusText)
}

func doResponse(m *slack.Message, c *slack.Client) error {
	logIncomingMessage(m, c)
	if m.User != "slackbot" && m.User != _botId && (lib.IsMention(m.Text, _botId) || (lib.IsDM(m.Channel))) {
		messageText := util.TrimWhitespace(lib.LessMentions(m.Text))
		for _, actionHandler := range commands {
			if lib.Like(messageText, actionHandler.expr) {
				return actionHandler.handler(m, c)
			}
		}
	}
	return nil
}

func doTime(m *slack.Message, c *slack.Client) error {
	now := time.Now().UTC()
	return lib.AnnounceTime(c, m.Channel, now)
}

func doDebugChannels(m *slack.Message, c *slack.Client) error {
	if len(c.ActiveChannels) == 0 {
		return say(c, m.Channel, "currently listening to *no* channels.")
	}
	activeChannelsText := "currently listening to the following channels:\n"
	for _, channelId := range c.ActiveChannels {
		if channel := findChannel(channelId); channel != nil {
			activeChannelsText = activeChannelsText + fmt.Sprintf(">#%s (id:%s)\n", channel.Name, channel.Id)
		}
	}
	return say(c, m.Channel, activeChannelsText)
}

func doJobsStatus(m *slack.Message, c *slack.Client) error {
	statusText := "current job statuses:\n"
	for _, status := range chronometer.Default().Status() {
		if len(status.RunningFor) != 0 {
			statusText = statusText + fmt.Sprintf(">`%s` - state: %s running for: %s\n", status.Name, status.State, status.RunningFor)
		} else {
			statusText = statusText + fmt.Sprintf(">`%s` - state: %s\n", status.Name, status.State)
		}
	}
	return say(c, m.Channel, statusText)
}

func doJobsRun(m *slack.Message, c *slack.Client) error {
	messageWithoutMentions := util.TrimWhitespace(lib.LessMentions(m.Text))
	pieces := strings.Split(messageWithoutMentions, " ")
	if len(pieces) > 1 {
		jobName := pieces[len(pieces)-1]
		chronometer.Default().RunJob(jobName)
		return sayf(c, m.Channel, "ran job `%s`", jobName)
	} else {
		chronometer.Default().RunAllJobs()
		return say(c, m.Channel, "ran all jobs")
	}
}

func doJobsCancel(m *slack.Message, c *slack.Client) error {
	messageWithoutMentions := util.TrimWhitespace(lib.LessMentions(m.Text))
	pieces := strings.Split(messageWithoutMentions, " ")
	if len(pieces) > 1 {
		taskName := pieces[len(pieces)-1]
		chronometer.Default().CancelTask(taskName)
		return sayf(c, m.Channel, "canceled task `%s`", taskName)
	}
	return doUnknown(m, c)
}

func doStocks(m *slack.Message, c *slack.Client) error {
	if job, hasJob := chronometer.Default().LoadedJobs["stocks"]; hasJob {
		if typedJob, isStocksJob := job.(*jobs.Stocks); isStocksJob {
			if len(typedJob.Tickers) == 0 {
				return say(c, m.Channel, "currently tracking *no* stocks.")
			}
			tickersLabels := []string{}
			for _, stock := range typedJob.Tickers {
				tickersLabels = append(tickersLabels, fmt.Sprintf("`%s`", stock))
			}
			tickersLabel := strings.Join(tickersLabels, " ")
			stocksText := fmt.Sprintf("currently tracking the following stocks: %s", tickersLabel)
			return say(c, m.Channel, stocksText)
		}
	}
	return doUnknown(m, c)
}

func doStocksPrice(m *slack.Message, c *slack.Client) error {
	if job, hasJob := chronometer.Default().LoadedJobs["stocks"]; hasJob {
		if typedJob, isStocksJob := job.(*jobs.Stocks); isStocksJob {
			if len(typedJob.Tickers) == 0 {
				return say(c, m.Channel, "currently tracking *no* stocks.")
			}
			stockInfo, stockErr := lib.StockPrice(typedJob.Tickers, lib.STOCK_DEFAULT_FORMAT)
			if stockErr != nil {
				return stockErr
			}
			return lib.AnnounceStocks(c, m.Channel, stockInfo)
		}
	}
	return doUnknown(m, c)
}

func doStockPrice(m *slack.Message, c *slack.Client) error {
	messageWithoutMentions := util.TrimWhitespace(lib.LessMentions(m.Text))
	pieces := strings.Split(messageWithoutMentions, " ")
	if len(pieces) > 1 {
		rawTicker := pieces[len(pieces)-1]
		tickers := []string{}
		if strings.Contains(rawTicker, ",") {
			tickers = strings.Split(rawTicker, ",")
		} else {
			tickers = []string{rawTicker}
		}
		stockInfo, stockErr := lib.StockPrice(tickers, lib.STOCK_DEFAULT_FORMAT)
		if stockErr != nil {
			return stockErr
		}
		return lib.AnnounceStocks(c, m.Channel, stockInfo)
	}
	return doUnknown(m, c)
}

func doStockTrack(m *slack.Message, c *slack.Client) error {
	messageWithoutMentions := util.TrimWhitespace(lib.LessMentions(m.Text))
	pieces := strings.Split(messageWithoutMentions, " ")
	ticker := pieces[len(pieces)-1]

	if job, hasJob := chronometer.Default().LoadedJobs["stocks"]; hasJob {
		if typedJob, isStocksJob := job.(*jobs.Stocks); isStocksJob {
			typedJob.Track(ticker, m.User)
			return sayf(c, m.Channel, "tracking `%s`", ticker)
		} else {
			logf("job `%s` is of type %#v", "stocks", job)
			return sayf(c, m.Channel, "job `%s` could not be marshalled", "stocks")
		}
	} else {
		return sayf(c, m.Channel, "job `%s` is not loaded", "stocks")
	}
	return doUnknown(m, c)
}

func doStockRemove(m *slack.Message, c *slack.Client) error {
	messageWithoutMentions := util.TrimWhitespace(lib.LessMentions(m.Text))
	pieces := strings.Split(messageWithoutMentions, " ")
	ticker := pieces[len(pieces)-1]

	if job, hasJob := chronometer.Default().LoadedJobs["stocks"]; hasJob {
		if typedJob, isStocksJob := job.(*jobs.Stocks); isStocksJob {
			typedJob.StopTracking(ticker)
			return sayf(c, m.Channel, "stopped tracking `%s`", ticker)
		} else {
			logf("job `%s` is of type %#v", "stocks", job)
			return sayf(c, m.Channel, "job `%s` could not be marshalled", "stocks")
		}
	} else {
		return sayf(c, m.Channel, "job `%s` is not loaded", "stocks")
	}
	return doUnknown(m, c)
}

func doOtherResponse(m *slack.Message, c *slack.Client) error {
	message := util.TrimWhitespace(lib.LessMentions(m.Text))
	if lib.IsSalutation(message) {
		return doSalutation(m, c)
	} else {
		return doUnknown(m, c)
	}
}

func doSalutation(m *slack.Message, c *slack.Client) error {
	user := findUser(m.User)
	salutation := []string{"hey %s", "hi %s", "hello %s", "ohayo gozaimasu %s", "salut %s", "bonjour %s", "yo %s", "sup %s"}
	return sayf(c, m.Channel, lib.Random(salutation), strings.ToLower(user.Profile.FirstName))
}

func doUnknown(m *slack.Message, c *slack.Client) error {
	return sayf(c, m.Channel, "I don't know how to respond to this\n>%s", m.Text)
}

func logIncomingMessage(m *slack.Message, c *slack.Client) {
	user := findUser(m.User)
	channel := findChannel(m.Channel)

	userName := "system"
	if user != nil {
		userName = user.Name
	}

	if channel != nil {
		logf("=> #%s (%s) - %s: %s", channel.Name, channel.Id, userName, m.Text)
	} else {
		logf("=> PM - %s: %s", userName, m.Text)
	}
}

func logOutgoingMessage(c *slack.Client, destinationId string, components ...interface{}) {
	if lib.Like(destinationId, "^C") {
		channel := findChannel(destinationId)
		logf("<= #%s (%s) - jarvis: %s", channel.Name, channel.Id, fmt.Sprint(components...))
	} else {
		logf("<= PM - jarvis: %s", fmt.Sprint(components...))
	}
}

func findUser(userId string) *slack.User {
	if user, hasUser := _usersLookup[userId]; hasUser {
		return &user
	}
	return nil
}

func findChannel(channelId string) *slack.Channel {
	if channel, hasChannel := _channelsLookup[channelId]; hasChannel {
		return &channel
	}
	return nil
}

func createUsersLookup(session *slack.Session) map[string]slack.User {
	lookup := map[string]slack.User{}
	for x := 0; x < len(session.Users); x++ {
		user := session.Users[x]
		lookup[user.Id] = user
	}
	return lookup
}

func createChannelLookup(session *slack.Session) map[string]slack.Channel {
	lookup := map[string]slack.Channel{}
	for x := 0; x < len(session.Channels); x++ {
		channel := session.Channels[x]
		lookup[channel.Id] = channel
	}
	return lookup
}

func say(c *slack.Client, destinationId string, components ...interface{}) error {
	logOutgoingMessage(c, destinationId, components...)
	return c.Say(destinationId, components...)
}

func sayf(c *slack.Client, destinationId string, format string, components ...interface{}) error {
	message := fmt.Sprintf(format, components...)
	logOutgoingMessage(c, destinationId, message)
	return c.Sayf(destinationId, format, components...)
}

func log(components ...interface{}) {
	message := fmt.Sprint(components...)
	fmt.Printf("%s - %s\n", time.Now().UTC().Format(time.RFC3339), message)
}

func logf(format string, components ...interface{}) {
	message := fmt.Sprintf(format, components...)
	fmt.Printf("%s - %s\n", time.Now().UTC().Format(time.RFC3339), message)
}

func writeToFile(path, contents string) error {
	f, fErr := os.Create(path)
	if fErr != nil {
		return exception.Wrap(fErr)
	}
	f.WriteString(contents)
	return nil
}
