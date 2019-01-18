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
}

func randomDisease() string {
	rand.Seed(time.Now().UTC().UnixNano())
	disease := diseases[rand.Intn(len(diseases))]
	return disease
}

func initializeApplication() error {
	users, err := allUsers()

	if err != nil {
		return err
	}

	if len(users) != 0 {
		return errors.New("I expected the user table to be emtpy but it's not")
	}

	slackUsers, err := slackClient.GetUsers()

	if err != nil {
		return err
	}

	for _, slackUser := range slackUsers {
		err := createUser(slackUser)

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
	slackUsers, err := slackClient.GetUsers()

	if err != nil {
		return err
	}

	knownUsers, err := allUsers()

	if err != nil {
		return err
	}

	babies, err := registerBabies(slackUsers, knownUsers)

	if err != nil {
		return err
	}

	err = announceBabies(babies)

	if err != nil {
		return err
	}

	corpses := fillMorgue(slackUsers, knownUsers)

	announceDeathsAndBury(corpses)

	return nil
}

func registerBabies(slackUsers []slack.User, knownUsers []User) ([]slack.User, error) {
	lookup := make(map[string]bool)

	for _, knownUser := range knownUsers {
		lookup[knownUser.ID] = true
	}

	babies := []slack.User{}

	for _, slackUser := range slackUsers {
		if !lookup[slackUser.ID] {
			babies = append(babies, slackUser)

			err := createUser(slackUser)

			if err != nil {
				return nil, err
			}
		}
	}

	return babies, nil
}

func announceBabies(babies []slack.User) error {
	for _, baby := range babies {
		text := ""
		if baby.Deleted {
			text = "I'm sorry for your loss, %s was stillborn"
		} else {
			text = "Congratulations, you have a beautiful new baby named %s"
		}

		name := ""
		if baby.RealName == "" {
			name = baby.Name
		} else {
			name = baby.RealName
		}

		_, _, err := slackClient.PostMessage(
			slackChannelID,
			slack.MsgOptionUser("trail"),
			slack.MsgOptionIconEmoji(":baby:"),
			slack.MsgOptionText(fmt.Sprintf(text, name), false),
		)

		if err != nil {
			return err
		}
	}

	return nil
}

func fillMorgue(slackUsers []slack.User, knownUsers []User) []User {
	deadLookup := map[string]User{}

	for _, user := range knownUsers {
		deadLookup[user.ID] = user
	}

	corpses := []User{}

	for _, slackUser := range slackUsers {
		user, known := deadLookup[slackUser.ID]

		// If known is false, that means we didn't know about the user yet. This condition is handled by
		// announcing new user (stillborn), so we skip it here
		if known && slackUser.Deleted && !user.Deleted {
			corpses = append(corpses, user)
		}
	}

	return corpses

}

func announceDeathsAndBury(corpses []User) error {
	for _, corpse := range corpses {
		err := corpse.AnnounceDeath()

		if err != nil {
			return err
		}

		err = corpse.Bury()

		if err != nil {
			return err
		}
	}

	return nil
}
