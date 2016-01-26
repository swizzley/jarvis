package main

import (
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/blendlabs/go-chronometer"
	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-request"
	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/go-slack"
)

var _usersLookup map[string]slack.User
var _channelsLookup map[string]slack.Channel
var _botId string
var _orders []order

var _searchResults map[string][]amazonProduct
var _scrollIndicies map[string]int

func TOKEN() string {
	return os.Getenv("SLACK_API_TOKEN")
}

func main() {
	client := slack.Connect(TOKEN())

	client.Listen(slack.EVENT_HELLO, func(m *slack.Message, c *slack.Client) {
		log("connected")
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
	_searchResults = map[string][]amazonProduct{}
	_scrollIndicies = map[string]int{}

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func doResponse(m *slack.Message, c *slack.Client) error {
	user := findUser(m.User)
	channel := findChannel(m.Channel)
	fullMessage := m.Text
	message := lessMentions(fullMessage)

	if channel == nil {
		return nil
	}

	userName := "system"
	if user != nil {
		userName = user.Name
	}

	logf("=> #%s (%s) - %s: %s", channel.Name, channel.Id, userName, message)

	if isMention(fullMessage) {
		if likeAny(message, []string{"^run jobs"}) {
			return doRunJobs(c, channel, user, message)
		} else if likeAny(message, []string{"^search", "^find"}) {
			return doSearch(c, channel, user, message)
		} else if likeAny(message, []string{"^next$", "^more$"}) {
			return doSearchNextResults(c, channel, user, message)
		} else if likeAny(message, []string{"^order ", "^add", "^include"}) {
			return doAddOrder(c, channel, user, message)
		} else if likeAny(message, []string{"^orders", "^list orders", "^show orders"}) {
			return doListOrders(c, channel, user, message)
		} else if likeAny(message, []string{"^purge orders", "^clear orders", "^empty orders"}) {
			return doClearOrders(c, channel, user, message)
		} else if isSalutation(message) {
			return doSalutation(c, channel, user, message)
		} else {
			return doUnknown(c, channel, user, message)
		}
	}
	return nil
}

func doRunJobs(c *slack.Client, channel *slack.Channel, user *slack.User, message string) error {
	chronometer.Default().RunAllJobs()
	return say(c, channel.Id, "Running Jobs")
}

func doSearch(c *slack.Client, channel *slack.Channel, user *slack.User, message string) error {
	results := amazonSearch(lessFirst(message))
	if len(results) != 0 {
		_searchResults[user.Id] = results
		_scrollIndicies[user.Id] = 2
		products := results[:3]

		resultsText := "I found the following product(s)\n"
		index := 1
		for _, product := range products {
			resultsText = resultsText + fmt.Sprintf("> #%d (%s) %s\n", index, product.price, product.url)
			index++
		}

		return say(c, channel.Id, resultsText)
	} else {
		return sayf(c, channel.Id, "No Results for Food Query\n>%s", message)
	}
}

func doSearchNextResults(c *slack.Client, channel *slack.Channel, user *slack.User, message string) error {
	if results, hasResults := _searchResults[user.Id]; hasResults {
		scrollIndex := _scrollIndicies[user.Id]
		_scrollIndicies[user.Id] = scrollIndex + 3
		products := results[scrollIndex : scrollIndex+3]
		resultsText := "I also found the following product(s)\n"
		index := scrollIndex + 1
		for _, product := range products {
			resultsText = resultsText + fmt.Sprintf("> #%d (%s) %s\n", index, product.price, product.url)
			index++
		}

		return say(c, channel.Id, resultsText)
	}
	return sayf(c, channel.Id, "No Search Results for %s", user.Name)
}

func doAddOrder(c *slack.Client, channel *slack.Channel, user *slack.User, message string) error {
	itemId := util.ParseInt(last(message))
	userId := user.Id

	searchResults := _searchResults[userId]
	result := searchResults[itemId-1]
	addOrder(user, result)
	return sayf(c, channel.Id, "Adding new product to order:\n>%s", result.name)
}

func doClearOrders(c *slack.Client, channel *slack.Channel, user *slack.User, message string) error {
	_orders = []order{}
	return say(c, channel.Id, "Removed all orders")
}

func doListOrders(c *slack.Client, channel *slack.Channel, user *slack.User, message string) error {
	if len(_orders) == 0 {
		return say(c, channel.Id, "I have no products to order.")
	}
	output := fmt.Sprintf("I have %d products listed to order:\n", len(_orders))
	for _, order := range _orders {
		user := findUser(order.ordered_by)
		output = output + fmt.Sprintf("%s has asked that we order:\n>%s (%s)\n", user.Profile.FirstName, order.product.name, order.product.price)
	}
	return say(c, channel.Id, output)
}

func doSalutation(c *slack.Client, channel *slack.Channel, user *slack.User, message string) error {
	salutation := []string{"Hey %s", "Hi %s", "Hello %s", "Ohayo Gozaimasu %s", "Salut %s", "Bonjour %s", "yo %s", "sup %s"}
	return sayf(c, channel.Id, random(salutation), user.Profile.FirstName)
}

func doUnknown(c *slack.Client, channel *slack.Channel, user *slack.User, message string) error {
	return sayf(c, channel.Id, "I don't know how to respond to this\n>%s", message)
}

func random(messages []string) string {
	return messages[rand.Intn(len(messages))]
}

func isMention(message string) bool {
	return like(message, fmt.Sprintf("<@%s>", _botId))
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

func addOrder(u *slack.User, product amazonProduct) {
	_orders = append(_orders, order{id: util.UUID_v4().ToShortString(), timestamp: time.Now().UTC(), ordered_by: u.Id, product: product})
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

func say(c *slack.Client, channelId string, components ...interface{}) error {
	channel := findChannel(channelId)
	logf("<= #%s (%s) - jarvis: %s", channel.Name, channel.Id, fmt.Sprint(components...))
	return c.Say(channelId, components...)
}

func sayf(c *slack.Client, channelId, format string, components ...interface{}) error {
	channel := findChannel(channelId)
	logf("<= #%s (%s) - jarvis: %s", channel.Name, channel.Id, fmt.Sprintf(format, components...))
	return c.Sayf(channelId, format, components...)
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
	ordered_by string
	product    amazonProduct
}

type amazonProduct struct {
	id       string
	name     string
	price    string
	url      string
	is_prime bool
}

func amazonSearch(query string) []amazonProduct {
	products := []amazonProduct{}

	results, fetchErr := request.NewRequest().
		AsGet().
		WithHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.82 Safari/537.36").
		WithScheme("http").
		WithHost("www.amazon.com").
		WithPath("s/").
		WithQueryString("field-keywords", query).
		FetchString()
	if fetchErr != nil {
		log(fetchErr)
		return products
	}

	if util.IsEmpty(results) {
		return products
	}

	doc, docErr := goquery.NewDocumentFromReader(strings.NewReader(results))

	if docErr != nil {
		log(docErr)
		return products
	}

	doc.Find("li.s-result-item").Each(func(i int, s *goquery.Selection) {
		link := s.Find("a.s-access-detail-page").First()
		price := s.Find("span.a-color-price").First()
		prime := s.Find("i.a-icon-prime").First()

		href, _ := link.Attr("href")

		_, urlErr := url.Parse(href)
		if urlErr != nil {
			return
		}

		product := amazonProduct{}
		product.id = util.UUID_v4().ToShortString()
		product.url = href
		product.name = link.Text()
		product.price = price.Text()
		product.is_prime = prime != nil

		if product.is_prime {
			products = append(products, product)
		}
	})

	return products
}

type OnTheQuarterHour struct{}

func (o OnTheQuarterHour) GetNextRunTime(after *time.Time) time.Time {
	if after == nil {
		now := time.Now().UTC()
		if now.Minute() >= 45 {
			return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 45, 0, 0, time.UTC).Add(15 * time.Minute)
		} else if now.Minute() >= 30 {
			return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 30, 0, 0, time.UTC).Add(15 * time.Minute)
		} else if now.Minute() >= 15 {
			return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 15, 0, 0, time.UTC).Add(15 * time.Minute)
		} else {
			return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC).Add(15 * time.Minute)
		}
	} else {
		if after.Minute() >= 45 {
			return time.Date(after.Year(), after.Month(), after.Day(), after.Hour(), 45, 0, 0, time.UTC).Add(15 * time.Minute)
		} else if after.Minute() >= 30 {
			return time.Date(after.Year(), after.Month(), after.Day(), after.Hour(), 30, 0, 0, time.UTC).Add(15 * time.Minute)
		} else if after.Minute() >= 15 {
			return time.Date(after.Year(), after.Month(), after.Day(), after.Hour(), 15, 0, 0, time.UTC).Add(15 * time.Minute)
		} else {
			return time.Date(after.Year(), after.Month(), after.Day(), after.Hour(), 0, 0, 0, time.UTC).Add(15 * time.Minute)
		}
	}
}

type TimeJob struct {
	Client *slack.Client
}

func (t TimeJob) Name() string {
	return "Clock"
}

func (t TimeJob) Execute(ct *chronometer.CancellationToken) error {
	currentTime := time.Now().UTC()

	for x := 0; x < len(t.Client.ActiveChannels); x++ {
		channelId := t.Client.ActiveChannels[x]
		return t.announceTime(channelId, currentTime)
	}
	return nil
}

func (t TimeJob) announceTime(channelId string, currentTime time.Time) error {
	timeText := fmt.Sprintf("%d:%d UTC", currentTime.Hour(), currentTime.Minute())
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

	_, messageErr := t.Client.ChatPostMessage(message)
	if messageErr != nil {
		fmt.Printf("issue posting message: %v\n", messageErr)
	}
	return messageErr
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
