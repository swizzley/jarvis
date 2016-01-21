package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/wcharczuk/go-slack"
)

var _usersLookup map[string]slack.User

func TOKEN() string {
	return os.Getenv("SLACK_API_TOKEN")
}

func main() {
	client := slack.Connect(TOKEN())

	client.Listen(slack.EVENT_HELLO, func(m *slack.Message, c *slack.Client) {
		log("Slack: Connected")
	})

	client.Listen(slack.EVENT_MESSAGE, func(m *slack.Message, c *slack.Client) {
		user := findUser(m.User)
		logf("Slack Message: %s - %s", user.Name, m.Text)
	})

	session, err := client.Start()
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	_usersLookup = createUsersLookup(session)

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func findUser(userId string) *slack.User {
	if user, hasUser := _usersLookup[userId]; hasUser {
		return &user
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

func log(components ...interface{}) {
	message := fmt.Sprint(components...)
	fmt.Printf("%s - %s\n", time.Now().UTC().Format(time.RFC3339), message)
}

func logf(format string, components ...interface{}) {
	message := fmt.Sprintf(format, components...)
	fmt.Printf("%s - %s\n", time.Now().UTC().Format(time.RFC3339), message)
}
