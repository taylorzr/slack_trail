package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	raven "github.com/getsentry/raven-go"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
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
		switch os.Getenv("COMMAND") {
		case "users":
			lambda.Start(withSentry(runUsersIteration))
		case "emojis":
			lambda.Start(withSentry(runEmojisIteration))
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

func initializeUsers() error {
	users, err := usersFromDatabase()

	if err != nil {
		return errors.Wrap(err, "fetching users from the database")
	}

	if len(users) != 0 {
		return errors.New("I expected the user table to be emtpy but it's not")
	}

	slackUsers, err := usersFromSlack()

	if err != nil {
		return errors.Wrap(err, "fetching users from slack")
	}

	for _, slackUser := range slackUsers {
		_, err := createUser(&slackUser)

		if err != nil {
			return errors.Wrapf(err, "creating user %#v", slackUser)
		}
	}

	return nil
}

func initializeEmojis() error {
	emojis, err := emojisFromDatabase()

	if err != nil {
		return errors.Wrap(err, "fetching emojis from the database")
	}

	if len(emojis) != 0 {
		return errors.New("I expected the emojis table to be emtpy but it's not")
	}

	slackEmojis, err := emojisFromSlack()

	if err != nil {
		return errors.Wrap(err, "fetching emojis from slack")
	}

	for _, slackEmoji := range slackEmojis {
		err := createEmoji(&slackEmoji)

		if err != nil {
			return errors.Wrapf(err, "creating emoji %#v", slackEmoji)
		}
	}

	return nil
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

func runEmojisIteration() error {
	slackEmojis, err := emojisFromSlack()

	if err != nil {
		return errors.Wrap(err, "fetching emojis from slack")
	}

	knownEmojis, err := emojisFromDatabase()

	if err != nil {
		return errors.Wrap(err, "fetching emojis from the database")
	}

	err = diffEmojis(knownEmojis, slackEmojis)

	return errors.Wrap(err, "diffing emojis")
}

func runUsersIteration() error {
	slackUsers, err := usersFromSlack()

	if err != nil {
		return errors.Wrap(err, "fetching users from slack")
	}

	knownUsers, err := usersFromDatabase()

	if err != nil {
		return errors.Wrap(err, "fetching users from the database")
	}

	err = diffUsers(knownUsers, slackUsers)

	return errors.Wrap(err, "diffing users")
}

func runMononymIterationWithSentry() error {
	fmt.Println("Starting mononym iteration...")
	var err error

	raven.CapturePanic(func() {
		err = runMononymIteration()

		if err != nil {
			fmt.Printf("Something broke :(\n%s\n", err.Error())
			raven.CaptureErrorAndWait(err, nil)
		}
	}, nil)

	fmt.Println("Finished mononym iteration.")

	return err
}

func runMononymIteration() error {
	users, err := usersFromMononym()
	usersLookup := map[string]bool{}

	if err != nil {
		return errors.Wrap(err, "fetching users from slack")
	}

	for _, user := range users {
		usersLookup[user.ID] = true
		if strings.ContainsAny(user.DisplayName, " .") {
			fmt.Printf("%s is an imposter!\n", user.DisplayName)
		}
	}

	users, err = usersFromSlack()

	if err != nil {
		return errors.Wrap(err, "fetching users from slack")
	}

	new := []User{}
	for _, user := range users {
		if !usersLookup[user.ID] && !user.Deleted && user.DisplayName != "" && !strings.ContainsAny(user.DisplayName, " .") {
			fmt.Printf("+ %s\n", user.DisplayName)
			new = append(new, user)
		}
	}

	return nil
}

func diffUsers(knownUsers, slackUsers []User) error {
	lookup := make(map[string]User)

	for _, knownUser := range knownUsers {
		lookup[knownUser.ID] = knownUser
	}

	for _, slackUser := range slackUsers {
		if user, ok := lookup[slackUser.ID]; ok {
			displayName := slackUser.DisplayName
			if displayName != user.DisplayName {
				err := user.ChangeName(slackUser.DisplayName)

				if err != nil {
					return errors.Wrap(err, "changing a users name")
				}
			}

			// NOTE: This is too spammy
			// if slackUser.Status != user.Status {
			// 	err := user.ChangeStatus(slackUser.Status)

			// 	if err != nil {
			// 		return errors.Wrap(err, "updating a users status")
			// 	}
			// }

			if slackUser.Title != user.Title {
				err := user.ChangeTitle(slackUser.Title)

				if err != nil {
					return errors.Wrap(err, "updating a users status")
				}
			}

			if slackUser.Deleted != user.Deleted {
				if slackUser.Deleted {
					err := user.Bury()

					if err != nil {
						return errors.Wrap(err, "burying a user")
					}
				} else {
					err := user.Necromance()

					if err != nil {
						return errors.Wrap(err, "raising a user from the dead")
					}
				}
			}
		} else {
			err := registerAndAnnounceBaby(slackUser)

			if err != nil {
				return errors.Wrapf(err, "delivering a new baby user %s", slackUser.DisplayName)
			}
		}
	}

	return nil
}

func diffEmojis(old, new []Emoji) error {
	oldLookup := make(map[string]Emoji)
	newLookup := make(map[string]Emoji)

	for _, o := range old {
		oldLookup[o.Name] = o
	}

	for _, n := range new {
		newLookup[n.Name] = n
	}

	for _, emoji := range new {
		if _, ok := oldLookup[emoji.Name]; !ok {
			message(fmt.Sprintf(":%s:", emoji.Name), ":heavy_plus_sign:")

			err := createEmoji(&emoji)

			if err != nil {
				return errors.Wrapf(err, "creating emoji %s", emoji.Name)
			}
		}
	}

	for _, emoji := range old {
		if _, ok := newLookup[emoji.Name]; !ok {
			message(fmt.Sprintf(":%s:", emoji.Name), ":heavy_minus_sign:")

			err := deleteEmoji(&emoji)

			if err != nil {
				return errors.Wrapf(err, "deleting emoji %s", emoji.Name)
			}
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

	return errors.Wrap(err, "sending a slack message")
}
