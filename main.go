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

	if err != nil {
		return err
	}

	for _, zombie := range change.Zombies {
		err := zombie.Necromance()

		if err != nil {
			return err
		}
	}

	return nil
}

func registerAndAnnounceBabies(babies []slack.User) error {
	for _, baby := range babies {
		err := createUser(baby)

		if err != nil {
			return err
		}

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

		err = message(fmt.Sprintf(text, name), ":baby:")

		if err != nil {
			return err
		}
	}

	return nil
}

type DiffResult struct {
	Babies  []slack.User
	Corpses []User
	Zombies []User
}

func (d *DiffResult) AddBaby(user slack.User) {
	d.Babies = append(d.Babies, user)
}

func (d *DiffResult) AddCorpse(user User) {
	d.Corpses = append(d.Corpses, user)
}

func (d *DiffResult) AddZombie(user User) {
	d.Zombies = append(d.Zombies, user)
}

func diff(users []User, slackUsers []slack.User) DiffResult {
	diff := DiffResult{}

	lookup := make(map[string]User)

	for _, user := range users {
		lookup[user.ID] = user
	}

	for _, slackUser := range slackUsers {
		if user, ok := lookup[slackUser.ID]; ok {
			if slackUser.Deleted != user.Deleted {
				if slackUser.Deleted {
					diff.AddCorpse(user)
				} else {
					diff.AddZombie(user)
				}
			}
		} else {
			diff.AddBaby(slackUser)
		}
	}

	return diff
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
