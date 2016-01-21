package main

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/blendlabs/go-request"
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
	message := m.Text
	messageLessMentions := withoutMentions(message)

	logf("incoming message: #%s - %s - %s", channel.Name, user.Name, message)
	if channel.Name == "bot-test" {
		if isMention(message) {
			if strings.HasPrefix(messageLessMentions, "food") {
				queryPieces := strings.Split(messageLessMentions, " ")[1:]
				results := amazonFoodSearch(strings.Join(queryPieces, " "))
				if len(results) != 0 {
					return sayf(c, channel.Id, results[0])
				} else {
					return sayf(c, channel.Id, "No Results for Food Query\n>%s", messageLessMentions)
				}

			} else {
				return sayf(c, channel.Id, "I don't know how to respond to this\n>%s", messageLessMentions)
			}
		}
	}
	return nil
}

func isMention(message string) bool {
	return like(message, fmt.Sprintf("<@%s>", _botId))
}

func withoutMentions(message string) string {
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

func like(corpus, expr string) bool {
	matched, _ := regexp.Match(expr, []byte(corpus))
	return matched
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

func amazonFoodSearch(query string) []string {
	queryEscaped := url.QueryEscape(query)
	queryFormat := "http://www.amazon.com/s/ref=nb_sb_noss_1?url=search-alias%3Dgrocery&field-keywords=%s"
	fullQuery := fmt.Sprintf(queryFormat, queryEscaped)

	results, fetchErr := request.NewRequest().AsGet().WithHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.82 Safari/537.36").WithUrl(fullQuery).FetchString()
	if fetchErr != nil {
		log(fetchErr)
		return []string{}
	}

	doc, docErr := goquery.NewDocumentFromReader(strings.NewReader(results))

	if docErr != nil {
		log(docErr)
		return []string{}
	}

	products := []string{}
	doc.Find(".a-link-normal").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		products = append(products, href)
	})

	return products
}
