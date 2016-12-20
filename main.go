package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	logger "github.com/blendlabs/go-logger"
	"github.com/blendlabs/go-util"
	"github.com/dlintw/goconf"
	"github.com/wcharczuk/jarvis/jarvis"
	"github.com/wcharczuk/jarvis/jarvis/core"
)

func key() []byte {
	keyBlob := os.Getenv("JARVIS_KEY")
	key, keyErr := util.String.Base64Decode(keyBlob)
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
	}
	return envPort
}

func main() {
	var bots []*jarvis.Bot
	var configFile = flag.String("config", "", "config file to read from")
	if configFile != nil && len(*configFile) != 0 {
		bots = initializeBotsFromConfig(*configFile)
	} else {
		bot, err := intializeBotFromEnvironment()
		if err != nil {
			fmt.Printf("Error Initializing Bot From Environment: %v\n", err)
			os.Exit(1)
		}

		bots = []*jarvis.Bot{bot}
	}

	startStatusServer(bots)
}

func intializeBotFromEnvironment() (*jarvis.Bot, error) {
	b, err := jarvis.NewBotFromEnvironment()
	if err != nil {
		return nil, err
	}
	err = b.Init()
	if err != nil {
		return nil, err
	}
	err = b.Start()
	if err != nil {
		return nil, err
	}
	return b, nil
}

func initializeBotsFromConfig(configPath string) []*jarvis.Bot {
	bots := []*jarvis.Bot{}
	config, err := goconf.ReadConfigFile(configPath)
	if err != nil {
		fmt.Printf("error reading config: %v\n", err)
		os.Exit(1)
	}

	for _, section := range config.GetSections() {
		tokenRaw, err := config.GetString(section, "SLACK_API_TOKEN")
		if err == nil {
			decryptedToken, err := decryptValue(tokenRaw)
			if err != nil {
				fmt.Printf("error decrypting slack token: %v\n", err)
				os.Exit(1)
			}
			j := jarvis.NewBot(decryptedToken)

			options, _ := config.GetOptions(section)
			for _, option := range options {
				if value, err := config.GetString(section, option); err == nil {
					decryptedValue, err := decryptValue(value)
					if err == nil {
						j.Configuration()[strings.ToUpper(option)] = decryptedValue
					} else {
						j.Configuration()[strings.ToUpper(option)] = value
					}
				}
			}

			j.Init()
			j.Start()
			bots = append(bots, j)
		} else {
			fmt.Printf("Error Reading `SLACK_API_TOKEN`: %v\n", err)
		}
	}
	return bots
}

func startStatusServer(bots []*jarvis.Bot) {
	http.HandleFunc("/", injectBots(bots, statusHandler))
	label := logger.ColorBlue.Apply("jarvis-cli")
	ts := logger.ColorLightBlack.Apply(time.Now().UTC().Format(time.RFC3339))
	fmt.Printf("%s - %s - starting status server, listening on: %s\n", label, ts, port())
	http.ListenAndServe(":"+port(), nil)
}

func injectBots(bots []*jarvis.Bot, h botAwareHTTPHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h(bots, w, r)
	}
}

type botAwareHTTPHandlerFunc func(bots []*jarvis.Bot, w http.ResponseWriter, r *http.Request)

func statusHandler(bots []*jarvis.Bot, w http.ResponseWriter, r *http.Request) {
	for _, bot := range bots {
		statusText := fmt.Sprintf("Jarvis is running and listening to the following channels (%s):\n", bot.OrganizationName())
		for _, channelID := range bot.Client().ActiveChannels {
			channel := bot.FindChannel(channelID)
			statusText = statusText + fmt.Sprintf("> #%s (%s)\n", channel.Name, channel.ID)
		}
		statusText = statusText + "\n"
		fmt.Fprintf(w, statusText)
	}
}

func encryptValue(value string) (string, error) {
	encrypted, encryptError := core.Encrypt(key(), value)
	if encryptError != nil {
		return util.StringEmpty, encryptError
	}

	return util.String.Base64Encode(encrypted), nil
}

func decryptValue(cipherText string) (string, error) {
	tokenBlob, tokenBlobErr := util.String.Base64Decode(cipherText)
	if tokenBlobErr != nil {
		return util.StringEmpty, tokenBlobErr
	}
	decrypted, decryptedErr := core.Decrypt(key(), tokenBlob)
	if decryptedErr != nil {
		return util.StringEmpty, decryptedErr
	}
	return decrypted, nil
}
