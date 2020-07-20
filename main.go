package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/getsentry/raven-go"
	"github.com/pkg/errors"
	"github.com/slack-go/slack"
)

/*
Concept:

Fetch all users from slack and compare the slack user list to our user list stored in the database

Any user who is in the slack list, but not our list is a new user, so announce those new users, and
store them in the database

Any user who is in the slack list as deleted, but in our list is not deleted, is a user who has been
recently deleted. Announce them as a deleted user, and update their deleted state in the database.

*/

var (
	slackChannelID string
	slackClient    *slack.Client
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	slackChannelID = os.Getenv("SLACK_CHANNEL_ID")
	slackClient = slack.New(os.Getenv("SLACK_TOKEN"))
}

// FIXME: Should just move these to cli instead of having to remember to setup and iteration in 2
// places
func main() {
	if _, exists := os.LookupEnv("LAMBDA"); exists {
		switch os.Getenv("COMMAND") {
		case "users":
			lambda.Start(withSentry(runUsersIteration))
		case "emojis":
			lambda.Start(withSentry(runEmojisIteration))
		case "employees":
			lambda.Start(withSentry(runEmployeesIteration))
		default:
			lambda.Start(func() error {
				return fmt.Errorf("env LAMBDA is set but no env COMMAND is missing. You must specify a COMMAND!")
			},
			)
		}
	} else {
		runCLI()
	}
}

// Oregon Trail Diseases:
var diseases = []string{
	"Dysentery",
	"Typhoid Fever",
	"Cholera",
	"Diphtheria",
	"Measles",
	"Thirst Traps",
}

func randomDisease() string {
	rand.Seed(time.Now().UTC().UnixNano())
	disease := diseases[rand.Intn(len(diseases)+1)]
	return disease
}

func message(text string, emoji string, attachments ...slack.Attachment) error {
	_, _, err := slackClient.PostMessage(
		slackChannelID,
		slack.MsgOptionUsername("trail"),
		slack.MsgOptionIconEmoji(emoji),
		slack.MsgOptionText(text, false),
		slack.MsgOptionAttachments(attachments...),
	)

	return errors.Wrap(err, "sending a slack message")
}

func withSentry(f func() error) func() error {
	function := f
	return func() error {

		fmt.Println("Starting iteration...")
		var err error

		raven.CapturePanic(func() {
			err = function()

			if err != nil {
				fmt.Printf("Something broke :(\n%s\n", err.Error())
				raven.CaptureErrorAndWait(err, nil)
			}
		}, nil)

		fmt.Println("Finished iteration.")

		return err
	}
}
