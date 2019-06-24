package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	raven "github.com/getsentry/raven-go"
	"github.com/nlopes/slack"
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

func main() {
	if _, exists := os.LookupEnv("LAMBDA"); exists {
		lambda.Start(runIterationWithSentry)
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

func initializeApplication() error {
	users, err := usersFromDatabase()

	if err != nil {
		return err
	}

	if len(users) != 0 {
		return errors.New("I expected the user table to be emtpy but it's not")
	}

	slackUsers, err := usersFromSlack()

	if err != nil {
		return err
	}

	for _, slackUser := range slackUsers {
		_, err := createUser(&slackUser)

		if err != nil {
			return err
		}
	}

	return nil
}

func runIterationWithSentry() error {
	fmt.Println("Starting iteration...")
	var err error

	raven.CapturePanic(func() {
		err = runIteration()

		if err != nil {
			fmt.Printf("Something broke :(\n%s\n", err.Error())
			raven.CaptureErrorAndWait(err, nil)
		}
	}, nil)

	fmt.Println("Finished iteration.")

	return err
}

func runIteration() error {
	slackUsers, err := usersFromSlack()

	if err != nil {
		return err
	}

	knownUsers, err := usersFromDatabase()

	if err != nil {
		return err
	}

	change := diff(knownUsers, slackUsers)

	err = registerAndAnnounceBabies(change.Babies)

	if err != nil {
		return err
	}

	for _, corpse := range change.Corpses {
		err = corpse.Bury()

		if err != nil {
			return err
		}
	}

	for _, zombie := range change.Zombies {
		err := zombie.Necromance()

		if err != nil {
			return err
		}
	}

	for _, nc := range change.NameChanges {
		err := nc.User.ChangeName(nc.NewName)

		if err != nil {
			return err
		}
	}

	for _, sc := range change.StatusChanges {
		err := sc.User.ChangeStatus(sc.NewStatus)

		if err != nil {
			return err
		}
	}

	return nil
}

func message(text string, emoji string, attachments ...slack.Attachment) error {
	_, _, err := slackClient.PostMessage(
		slackChannelID,
		slack.MsgOptionUsername("trail"),
		slack.MsgOptionIconEmoji(emoji),
		slack.MsgOptionText(text, false),
		slack.MsgOptionAttachments(attachments...),
	)

	return err
}
