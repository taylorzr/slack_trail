package main

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

type Emoji struct {
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

func emojisFromDatabase() ([]Emoji, error) {
	emojis := []Emoji{}

	err := db.Select(&emojis, "SELECT * FROM emojis")

	if err != nil {
		return nil, errors.Wrap(err, "selecting all emojis")
	}

	return emojis, nil
}

func getEmoji(name string) (*Emoji, error) {
	emoji := Emoji{}

	err := db.Get(&emoji, "SELECT * from emojis where name = $1", name)

	return &emoji, errors.Wrapf(err, "get emoji %s failed", name)
}

func createEmoji(emoji *Emoji) error {
	emoji.CreatedAt = time.Now()

	_, err := db.NamedExec(`
		INSERT INTO emojis
		(name, created_at)
		VALUES
		(:name, :created_at)
		`, emoji)

	return errors.Wrapf(err, "inserting emoji %#v", emoji)
}

func deleteEmoji(emoji *Emoji) error {
	_, err := db.NamedExec(`DELETE FROM emojis WHERE name = :name`, emoji)

	return err
}

func emojisFromSlack() ([]Emoji, error) {
	slackEmojis, err := slackClient.GetEmoji()

	if err != nil {
		return nil, errors.Wrap(err, "getting slack emoji")
	}

	emojis := []Emoji{}

	for name, _ := range slackEmojis {
		emojis = append(emojis, Emoji{Name: name})
	}

	return emojis, nil
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
			sendMessage(fmt.Sprintf(":%s:", emoji.Name), ":heavy_plus_sign:")

			err := createEmoji(&emoji)

			if err != nil {
				return errors.Wrapf(err, "creating emoji %s", emoji.Name)
			}
		}
	}

	for _, emoji := range old {
		if _, ok := newLookup[emoji.Name]; !ok {
			sendMessage(fmt.Sprintf(":%s:", emoji.Name), ":heavy_minus_sign:")

			err := deleteEmoji(&emoji)

			if err != nil {
				return errors.Wrapf(err, "deleting emoji %s", emoji.Name)
			}
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
