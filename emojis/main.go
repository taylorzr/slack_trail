package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	raven "github.com/getsentry/raven-go"
	"github.com/nlopes/slack"
	"github.com/pkg/errors"
)

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

func initializeApplication() error {
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
	slackEmojis, err := emojisFromSlack()

	if err != nil {
		return errors.Wrap(err, "fetching emojis from slack")
	}

	knownEmojis, err := emojisFromDatabase()

	if err != nil {
		return errors.Wrap(err, "fetching emojis from the database")
	}

	err = diff(knownEmojis, slackEmojis)

	return errors.Wrap(err, "diffing emojis")
}

func diff(old, new []Emoji) error {
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
		slack.MsgOptionUsername("emoji"),
		slack.MsgOptionIconEmoji(emoji),
		slack.MsgOptionText(text, false),
		slack.MsgOptionAttachments(attachments...),
	)

	return errors.Wrap(err, "sending a slack message")
}
