package main

import (
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
	"github.com/nlopes/slack"
)

type User struct {
	ID          string      `db:"id"`
	Deleted     bool        `db:"deleted"`
	RealName    string      `db:"real_name"`
	DisplayName string      `db:"display_name"`
	Name        string      `db:"name"`
	Avatar      string      `db:"avatar"`
	Status      string      `db:"status"`
	CreatedAt   time.Time   `db:"created_at"`
	DeletedAt   pq.NullTime `db:"deleted_at"`
}

func (user User) Update() error {
	_, err := db.NamedExec(`
    UPDATE users SET
			name = :name,
			real_name = :real_name,
			display_name = :display_name,
			avatar = :avatar,
			deleted = :deleted,
			deleted_at = :deleted_at,
			status = :status
		WHERE
		  id = :id
	`, user)

	return err
}

func (user User) Bury() error {
	text := fmt.Sprintf("After %s, %s died of %s", user.Age(), user.SomeName(), randomDisease())

	err := message(text, ":rip:", slack.Attachment{
		ImageURL: user.Avatar,
		Title:    "",
	})

	if err != nil {
		return err
	}

	user.Deleted = true
	user.DeletedAt = pq.NullTime{Time: time.Now(), Valid: true}

	err = user.Update()

	return err
}

func (user User) ChangeName(newName string) error {
	text := fmt.Sprintf("%s changed their handle from %s to %s", user.SomeName(), user.DisplayName, newName)

	err := message(text, ":name_badge:")

	if err != nil {
		return err
	}

	user.DisplayName = newName

	err = user.Update()

	return err
}

func ignorableStatus(from, to string) bool {
	ignorable := map[string]bool{":slack_call: On a call": true}

	return ignorable[from] || ignorable[to]
}

func (user User) ChangeStatus(newStatus string) error {
	text := fmt.Sprintf("%s changed their status from %s to %s", user.SomeName(), user.Status, newStatus)

	if ignorableStatus(user.Status, newStatus) {
		log.Println("Status is spam, not sending slack message...")
	} else {
		err := message(text, ":thought_balloon:")

		if err != nil {
			return err
		}
	}

	user.Status = newStatus

	err := user.Update()

	return err
}

func (user User) Necromance() error {
	text := fmt.Sprintf("%s is back from the dead!", user.SomeName())

	err := message(text, ":zombie:", slack.Attachment{
		ImageURL: user.Avatar,
		Title:    "",
	})

	if err != nil {
		return err
	}

	_, err = db.NamedExec(`
		UPDATE users SET
		  deleted = false,
			deleted_at = null
		WHERE id = :id`, user)

	return err
}

func (user User) SomeName() string {
	if user.RealName == "" {
		return user.Name
	} else {
		return user.RealName
	}
}

func (user User) Age() time.Duration {
	t := time.Now()

	if user.DeletedAt.Valid {
		t = user.DeletedAt.Time
	}

	duration := t.Sub(user.CreatedAt)

	return duration
}

func getUser(name string) (User, error) {
	user := User{}

	err := db.Get(&user, "SELECT * from users where name = $1", name)

	return user, err
}

func allUsers() ([]User, error) {
	users := []User{}

	err := db.Select(&users, "SELECT * FROM users")

	if err != nil {
		return nil, err
	}

	return users, nil
}

func createUser(slacker slack.User) (User, error) {
	user := fromSlacker(slacker)
	user.CreatedAt = time.Now()

	if user.Deleted {
		user.DeletedAt = pq.NullTime{Time: time.Now(), Valid: true}
	}

	_, err := db.NamedExec(`
		INSERT INTO users
		(id, name, real_name, display_name, avatar, deleted, deleted_at, created_at, status)
		VALUES
		(:id, :name, :real_name, :display_name, :avatar, :deleted, :deleted_at, :created_at, :status)
		`, user)

	return user, err
}

func fromSlacker(slacker slack.User) User {
	return User{
		ID:          slacker.ID,
		Name:        slacker.Name,
		RealName:    slacker.RealName,
		DisplayName: slacker.Profile.DisplayName,
		Deleted:     slacker.Deleted,
		Avatar:      slacker.Profile.ImageOriginal,
		Status:      fmt.Sprintf("%s %s", slacker.Profile.StatusEmoji, slacker.Profile.StatusText),
	}
}

func registerAndAnnounceBabies(babies []slack.User) error {
	for _, baby := range babies {
		user, err := createUser(baby)

		if err != nil {
			return err
		}

		text := ""
		if baby.Deleted {
			text = "I'm sorry for your loss, %s was stillborn"
		} else {
			text = "Congratulations, you have a beautiful new baby named %s"
		}

		err = message(fmt.Sprintf(text, user.SomeName()), ":baby:")

		if err != nil {
			return err
		}
	}

	return nil
}
