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
	"github.com/blendlabs/go-exception"
	"github.com/blendlabs/go-request"
	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/go-slack"
)

var _usersLookup map[string]slack.User
var _channelsLookup map[string]slack.Channel
var _botId string

func TOKEN() string {
	return os.Getenv("SLACK_API_TOKEN")
}

func main() {
	client := slack.Connect(TOKEN())

	client.Listen(slack.EVENT_HELLO, func(m *slack.Message, c *slack.Client) {
		log("connected")
	})

	client.Listen(slack.EVENT_MESSAGE, func(m *slack.Message, c *slack.Client) {
		logf("message hook fired")
		resErr := doResponse(m, c)
		if resErr != nil {
			log(resErr)
		}
	})

	session, err := client.Start()
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	_botId = session.Self.Id
	_usersLookup = createUsersLookup(session)
	_channelsLookup = createChannelLookup(session)

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func doResponse(m *slack.Message, c *slack.Client) error {
	user := findUser(m.User)
	channel := findChannel(m.Channel)
	fullMessage := m.Text
	message := lessMentions(fullMessage)

	if user == nil || channel == nil {
		return nil
	}

	logf("incoming message: #%s - %s - %s", channel.Name, user.Name, message)
	if isMention(fullMessage) {
		if strings.HasPrefix(message, "food") {
			results := amazonFoodSearch(lessFirst(message))
			if len(results) != 0 {
				firstProduct := results[0]
				return sayf(c, channel.Id, "I found the following product, it costs %s\n>%s", firstProduct.price, firstProduct.url)
			} else {
				return sayf(c, channel.Id, "No Results for Food Query\n>%s", message)
			}
		} else if isSalutation(message) {
			salutation := []string{"Hey %s", "Hi %s", "Hello %s", "Ohayo Gozamaisu %s", "Salut %s", "Yo %s"}
			return sayf(c, channel.Id, random(salutation), user.Profile.FirstName)
		} else {
			return sayf(c, channel.Id, "I don't know how to respond to this\n>%s", message)
		}
	}
	return nil
}

func random(messages []string) string {
	return messages[rand.Intn(len(messages))]
}

func isMention(message string) bool {
	return like(message, fmt.Sprintf("<@%s>", _botId))
}

func isSalutation(message string) bool {
	return likeAny(message, []string{"^hello", "^hi", "^greetings", "^hey"})
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

func sayf(c *slack.Client, channelId, format string, components ...interface{}) error {
	channel := findChannel(channelId)
	logf("outgoing message: #%s - %s", channel.Name, fmt.Sprintf(format, components...))
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

type amazonProduct struct {
	name     string
	price    string
	url      string
	is_prime bool
}

func amazonFoodSearch(query string) []amazonProduct {
	products := []amazonProduct{}
	queryEscaped := url.QueryEscape(query)
	queryFormat := "http://www.amazon.com/s/?url=search-alias%%3Dgrocery&field-keywords=%s"
	fullQuery := fmt.Sprintf(queryFormat, queryEscaped)

	results, fetchErr := request.NewRequest().AsGet().WithHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.82 Safari/537.36").WithUrl(fullQuery).FetchString()
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
		href, _ := link.Attr("href")
		price := s.Find("span.a-color-price")

		prime := s.Find("i.a-icon-prime").First()

		product := amazonProduct{}
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

func writeToFile(path, contents string) error {
	f, fErr := os.Create(path)
	if fErr != nil {
		return exception.Wrap(fErr)
	}
	f.WriteString(contents)
	return nil
}
