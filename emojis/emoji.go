package main

import (
	"time"

	"github.com/pkg/errors"
)

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

func emojisFromDatabase() ([]Emoji, error) {
	emojis := []Emoji{}

	err := db.Select(&emojis, "SELECT * FROM emojis")

	if err != nil {
		return nil, errors.Wrap(err, "selecting all emojis")
	}

	return emojis, nil
}

type Emoji struct {
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
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
