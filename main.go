package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/blendlabs/go-util"
	"github.com/wcharczuk/jarvis-cli/Godeps/_workspace/src/github.com/dlintw/goconf"
	"github.com/wcharczuk/jarvis-cli/jarvis"
)

func key() []byte {
	keyBlob := os.Getenv("JARVIS_KEY")
	key, keyErr := util.Base64Decode(keyBlob)
	if keyErr != nil {
		fmt.Printf("error reading key: %v\n", keyErr)
		os.Exit(1)
	}
	return key
}

func port() string {
	envPort := os.Getenv("PORT")
	if len(envPort) == 0 {
		return "8888"
	} else {
		return envPort
	}
}

func main() {
	args := os.Args
	if len(args) == 1 {
		bots := initializeBotsFromConfig("jarvis.conf")
		startStatusServer(bots)
	} else {
		command := args[1]
		switch strings.ToLower(command) {
		case "key":
			fmt.Printf("JARVIS_KEY=%s\n", util.Base64Encode(jarvis.CreateKey(32)))
			os.Exit(0)
		case "encrypt":
			if len(args) < 3 {
				fmt.Println("need to provide a value to `encrypt`")
				os.Exit(1)
			}
			value := args[2]
			fmt.Printf("%s\n", encryptValue(value))
			os.Exit(0)
		}
	}
}

func initializeBotsFromConfig(configPath string) []*jarvis.JarvisBot {
	bots := []*jarvis.JarvisBot{}
	config, err := goconf.ReadConfigFile(configPath)
	if err != nil {
		fmt.Printf("error reading config: %v\n", err)
		os.Exit(1)
	}

	for _, section := range config.GetSections() {
		tokenRaw, tokenErr := config.GetString(section, "SLACK_API_TOKEN")
		if tokenErr == nil {
			j := jarvis.NewJarvisBot(decryptValue(tokenRaw))

			if jiraCredentials, jiraCredentialsErr := config.GetString(section, "JIRA_CREDENTIALS"); jiraCredentialsErr == nil {
				j.Configuration["JIRA_CREDENTIALS"] = decryptValue(jiraCredentials)
			}

			if jiraHost, jiraHostErr := config.GetString(section, "JIRA_HOST"); jiraHostErr == nil {
				j.Configuration["JIRA_HOST"] = jiraHost
			}

			j.Init()
			j.Start()
			bots = append(bots, j)
		}
	}
	return bots
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

func encryptValue(value string) string {
	encrypted, encryptError := jarvis.Encrypt(key(), value)
	if encryptError != nil {
		fmt.Printf("error encrypting value: %v\n", encryptError)
		os.Exit(1)
	}

	return util.Base64Encode(encrypted)
}

func decryptValue(cipherText string) string {
	tokenBlob, tokenBlobErr := util.Base64Decode(cipherText)
	if tokenBlobErr != nil {
		fmt.Printf("error reading value: %v\n", tokenBlobErr)
		os.Exit(1)
	}
	decrypted, decryptedErr := jarvis.Decrypt(key(), tokenBlob)
	if decryptedErr != nil {
		fmt.Printf("error decrypting value: %v\n", decryptedErr)
		os.Exit(1)
	}
	return decrypted
}
