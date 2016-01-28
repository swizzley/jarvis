package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/go-slack"
)

var ADMINS = []string{"isadora", "will"}

var _usersLookup map[string]slack.User
var _channelsLookup map[string]slack.Channel
var _botId string
var _orders []order

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

	chronometer.Default().LoadJob(TimeJob{Client: client})
	chronometer.Default().Start()

	session, err := client.Start()
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	_botId = session.Self.Id
	_usersLookup = createUsersLookup(session)
	_channelsLookup = createChannelLookup(session)
	_orders = []order{}

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

	if isMention(m.Text) || channel == nil && isAdminUser(user.Name) {
		return processMessage(m, c)
	}
	return nil
}

func processMessage(m *slack.Message, c *slack.Client) error {
	message := lessMentions(m.Text)
	if likeAny(message, []string{"^order ", "^add", "^include"}) {
		return doAddOrder(m, c)
	} else if likeAny(message, []string{"^orders", "^list orders", "^show orders"}) {
		return doListOrders(m, c)
	} else if likeAny(message, []string{"^purge orders", "^clear orders", "^empty orders"}) {
		return doClearOrders(m, c)
	} else if isSalutation(message) {
		return doSalutation(m, c)
	} else if like(message, "^debug") {
		return doDebug(m, c)
	} else if like(message, "^time") {
		return doTime(m, c)
	} else {
		return doUnknown(m, c)
	}
}

func doDebug(m *slack.Message, c *slack.Client) error {
	message := lessMentions(m.Text)
	if like(message, "^debug list channels") {
		activeChannelsText := "Currently listening to the following channels:\n"
		for _, channelId := range c.ActiveChannels {
			if channel := findChannel(channelId); channel != nil {
				activeChannelsText = activeChannelsText + fmt.Sprintf(">#%s (id:%s)\n", channel.Name, channel.Id)
			}
		}
		return say(c, m.Channel, activeChannelsText)
	} else if like(message, "^debug jobs run") {
		chronometer.Default().RunAllJobs()
		return say(c, m.Channel, "Running Jobs")
	} else if like(message, "^debug jobs next-run-times") {
		nextRunTimes := "Here are the loaded jobs and their next run times:\n"
		for k, v := range chronometer.Default().NextRunTimes {
			nextRunTimes = nextRunTimes + fmt.Sprintf("> job: %s at %s", k, v.Format(time.RFC3339))
		}
		return say(c, m.Channel, nextRunTimes)
	}
	return say(c, m.Channel, "I'm not sure how to run that debugging command.")
}

func doAddOrder(m *slack.Message, c *slack.Client) error {
	messageLast := last(m.Text)
	productUrlRaw := extractTags(messageLast)
	productUrl, urlErr := url.Parse(productUrlRaw)
	if urlErr != nil {
		return say(c, m.Channel, "That is not a valid product url.")
	}
	if !likeAny(productUrl.Host, []string{"amazon.com$", "instacart.com$", "freshdirect.com$", "jet.com$"}) {
		return say(c, m.Channel, "That url is not from an approved online retailer.")
	}

	addOrder(m.User, productUrl.String())
	return sayf(c, m.Channel, "Adding new product to order:\n>%s", productUrl.String())
}

func doClearOrders(m *slack.Message, c *slack.Client) error {
	_orders = []order{}
	return say(c, m.Channel, "Removed all orders")
}

func doListOrders(m *slack.Message, c *slack.Client) error {
	if len(_orders) == 0 {
		return say(c, m.Channel, "I have no products to order.")
	}
	output := fmt.Sprintf("I have %d products listed to order:\n", len(_orders))
	for _, order := range _orders {
		user := findUser(order.orderedBy)
		output = output + fmt.Sprintf("%s has asked that we order:\n>%s\n", user.Profile.FirstName, order.productUrl)
	}
	return say(c, m.Channel, output)
}

func doTime(m *slack.Message, c *slack.Client) error {
	now := time.Now().UTC()
	return announceTime(c, m.Channel, now)
}

func doSalutation(m *slack.Message, c *slack.Client) error {
	user := findUser(m.User)
	salutation := []string{"Hey %s", "Hi %s", "Hello %s", "Ohayo Gozaimasu %s", "Salut %s", "Bonjour %s", "yo %s", "sup %s"}
	return sayf(c, m.Channel, random(salutation), user.Profile.FirstName)
}

func doUnknown(m *slack.Message, c *slack.Client) error {
	return sayf(c, m.Channel, "I don't know how to respond to this\n>%s", m.Text)
}

func random(messages []string) string {
	return messages[rand.Intn(len(messages))]
}

func isMention(message string) bool {
	return like(message, fmt.Sprintf("<@%s>", _botId))
}

func isAdminUser(userName string) bool {
	return any(userName, ADMINS)
}

func isDebugChannel(channel *slack.Channel) bool {
	return like(channel.Name, "bot-test")
}

func isSalutation(message string) bool {
	return likeAny(message, []string{"^hello", "^hi", "^greetings", "^hey"})
}

func isAsking(message string) bool {
	return likeAny(message, []string{"would it be possible", "can you", "would you", "is it possible", "([^.?!]*)\\?"})
}

func isPolite(message string) bool {
	return likeAny(message, []string{"please", "thanks"})
}

func isVulgar(message string) bool {
	return likeAny(message, []string{"fuck", "shit", "ass", "cunt"}) //yep.
}

func isAngry(message string) bool {
	return likeAny(message, []string{"stupid", "worst", "terrible", "horrible", "cunt"}) //yep.
}

func lessMentions(message string) string {
	output := ""
	state := 0
	for _, c := range message {
		switch state {
		case 0:
			if c == rune("<"[0]) {
				state = 1
			} else {
				output = output + string(c)
			}
		case 1:
			if c == rune(">"[0]) {
				state = 2
			}
		case 2:
			if c == rune(":"[0]) {
				state = 2
			} else if c == rune(" "[0]) {
				state = 0
			} else {
				state = 0
				output = output + string(c)
			}
		}
	}
	return output
}

func extractTags(message string) string {
	output := ""
	for _, c := range message {
		if !(c == rune("<"[0]) || c == rune(">"[0])) {
			output = output + string(c)
		}
	}
	return output
}

func lessFirst(message string) string {
	queryPieces := strings.Split(message, " ")[1:]
	return strings.Join(queryPieces, " ")
}

func last(message string) string {
	pieces := strings.Split(message, " ")
	if len(pieces) != 0 {
		return pieces[len(pieces)-1]
	} else {
		return ""
	}
}

func like(corpus, expr string) bool {
	matched, _ := regexp.Match(expr, []byte(corpus))
	return matched
}

func likeAny(corpus string, exprs []string) bool {
	for _, expr := range exprs {
		if like(corpus, expr) {
			return true
		}
	}
	return false
}

func any(value string, values []string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
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

func addOrder(userId string, productUrl string) {
	_orders = append(_orders, order{id: util.UUID_v4().ToShortString(), timestamp: time.Now().UTC(), orderedBy: userId, productUrl: productUrl})
}

func removeOrder(u *slack.User, orderId string) []order {
	newOrders := []order{}
	for _, order := range _orders {
		if order.id != orderId {
			newOrders = append(newOrders, order)
		}
	}
	return newOrders
}

func announceTime(c *slack.Client, channelId string, currentTime time.Time) error {
	timeText := fmt.Sprintf("%s UTC", currentTime.Format(time.Kitchen))
	message := slack.NewChatMessage(channelId, "")
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

	_, messageErr := c.ChatPostMessage(message)
	if messageErr != nil {
		fmt.Printf("issue posting message: %v\n", messageErr)
	}
	return messageErr
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
	if like(destinationId, "^C") {
		channel := findChannel(destinationId)
		logf("<= #%s (%s) - jarvis: %s", channel.Name, channel.Id, fmt.Sprint(components...))
	} else {
		logf("<= PM - jarvis: %s", fmt.Sprint(components...))
	}

	return c.Say(destinationId, components...)
}

func sayf(c *slack.Client, destinationId string, format string, components ...interface{}) error {
	if like(destinationId, "^C") {
		channel := findChannel(destinationId)
		logf("<= #%s (%s) - jarvis: %s", channel.Name, channel.Id, fmt.Sprintf(format, components...))
	} else {
		logf("<= PM - jarvis: %s", fmt.Sprintf(format, components...))
	}
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

type order struct {
	id         string
	timestamp  time.Time
	orderedBy  string `json:"ordered_by"`
	productUrl string `json:"product_url"`
}

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

type TimeJob struct {
	Client *slack.Client
}

func (t TimeJob) Name() string {
	return "Clock"
}

func (t TimeJob) Execute(ct *chronometer.CancellationToken) error {
	logf("job `%s` running", t.Name())
	currentTime := time.Now().UTC()

	for x := 0; x < len(t.Client.ActiveChannels); x++ {
		channelId := t.Client.ActiveChannels[x]
		return announceTime(t.Client, channelId, currentTime)
	}
	return nil
}

func (t TimeJob) Schedule() chronometer.Schedule {
	return OnTheQuarterHour{}
}

func writeToFile(path, contents string) error {
	f, fErr := os.Create(path)
	if fErr != nil {
		return exception.Wrap(fErr)
	}
	f.WriteString(contents)
	return nil
}
