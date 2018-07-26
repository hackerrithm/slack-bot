package main

import (
	"log"
	"os"

	"github.com/Krognol/go-wolfram"
	"github.com/christianrondeau/go-wit"

	"github.com/nlopes/slack"
)

const confidenceThreshold = 0.5

// external client variables
var (
	slackClient   *slack.Client
	witClient     *wit.Client
	wolframClient *wolfram.Client
)

func main() {
	// get environment variable for slack access token
	slackClient = slack.New(os.Getenv("SLACK_ACCESS_TOKEN"))
	// get environment variable for wit.ai access token
	witClient = wit.NewClient(os.Getenv("WIT_AI_ACCESS_TOKEN"))
	// get environment variable for wolfram app ID
	wolframClient = &wolfram.Client{AppID: os.Getenv("WOLFRAM_APP_ID")}

	// passes socket connection established for slack client
	rtm := slackClient.NewRTM()
	// starts this connection in a goroutine
	go rtm.ManageConnection()

	// iterates over messages comming in and handles messages accordingly
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if len(ev.BotID) == 0 {
				go handleMessage(ev)
			}
		}
	}
}

// handleMessage
func handleMessage(ev *slack.MessageEvent) {
	result, err := witClient.Message(ev.Msg.Text)
	if err != nil {
		log.Printf("unable to get wit.ai result: %v", err)
		return
	}

	var (
		topEntity    wit.MessageEntity
		topEntityKey string
	)

	for key, entityList := range result.Entities {
		for _, entity := range entityList {
			if entity.Confidence > confidenceThreshold && entity.Confidence > topEntity.Confidence {
				topEntity = entity
				topEntityKey = key
			}
		}
	}

	replyToUser(ev, topEntity, topEntityKey)
}

func replyToUser(ev *slack.MessageEvent, topEntity wit.MessageEntity, topEntityKey string) {
	switch topEntityKey {
	case "greetings":
		slackClient.PostMessage(ev.User, "Hello user! How can I help you?", slack.PostMessageParameters{
			AsUser: true,
		})
		return
	case "bye":
		slackClient.PostMessage(ev.User, "Bye user! See you later?", slack.PostMessageParameters{
			AsUser: true,
		})
		return
	case "thanks":
		slackClient.PostMessage(ev.User, "Thank you", slack.PostMessageParameters{
			AsUser: true,
		})
		return
	case "wolfram_search_query":
		res, err := wolframClient.GetSpokentAnswerQuery(topEntity.Value.(string), wolfram.Metric, 1000)
		if err == nil {
			slackClient.PostMessage(ev.User, res, slack.PostMessageParameters{
				AsUser: true,
			})
			return
		}

		log.Printf("unable to get data from wolfram: %v", err)
	}

	slackClient.PostMessage(ev.User, "¯\\_(o_o)_//¯", slack.PostMessageParameters{
		AsUser: true,
	})
}
