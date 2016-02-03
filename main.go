package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/wcharczuk/jarvis-cli/Godeps/_workspace/src/github.com/dlintw/goconf"
	"github.com/wcharczuk/jarvis-cli/jarvis"
)

func port() string {
	envPort := os.Getenv("PORT")
	if len(envPort) == 0 {
		return "8888"
	} else {
		return envPort
	}
}

func main() {
	bots := []*jarvis.JarvisBot{}

	config, err := goconf.ReadConfigFile("jarvis.conf")
	if err != nil {
		fmt.Printf("error reading config: %v\n", err)
		os.Exit(1)
	}

	for _, section := range config.GetSections() {
		token, tokenErr := config.GetString(section, "SLACK_API_TOKEN")
		if tokenErr == nil {
			j := jarvis.NewJarvisBot(token)
			j.Init()
			j.Start()
			bots = append(bots, j)
		}
	}

	//start up the bots.
	startStatusServer(bots)
}

func startStatusServer(bots []*jarvis.JarvisBot) {
	http.HandleFunc("/", injectBots(bots, statusHandler))
	fmt.Printf("jarvis-cli - %s - starting status server, listening on: %s\n", time.Now().UTC().Format(time.RFC3339), port())
	http.ListenAndServe(":"+port(), nil)
}

func injectBots(bots []*jarvis.JarvisBot, h botAwareHttpHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h(bots, w, r)
	}
}

type botAwareHttpHandlerFunc func(bots []*jarvis.JarvisBot, w http.ResponseWriter, r *http.Request)

func statusHandler(bots []*jarvis.JarvisBot, w http.ResponseWriter, r *http.Request) {
	for _, bot := range bots {
		statusText := fmt.Sprintf("Jarvis is running and listening to the following channels (%s):\n", bot.OrganizationName)
		for _, channelId := range bot.Client.ActiveChannels {
			channel := bot.FindChannel(channelId)
			statusText = statusText + fmt.Sprintf("> #%s (%s)\n", channel.Name, channel.Id)
		}
		statusText = statusText + "\n"
		fmt.Fprintf(w, statusText)
	}
}
