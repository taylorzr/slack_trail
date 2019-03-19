package main

import (
	"fmt"
	"time"

	"github.com/lib/pq"
	"github.com/nlopes/slack"
)

type User struct {
	ID        string      `db:"id"`
	Deleted   bool        `db:"deleted"`
	RealName  string      `db:"real_name"`
	Name      string      `db:"name"`
	Avatar    string      `db:"avatar"`
	CreatedAt time.Time   `db:"created_at"`
	DeletedAt pq.NullTime `db:"deleted_at"`
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

	user.DeletedAt = pq.NullTime{Time: time.Now(), Valid: true}

	_, err = db.NamedExec(`
		UPDATE users SET
		  deleted = true,
			deleted_at = :deleted_at
		WHERE id = :id`, user)

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

func createUser(slacker slack.User) error {
	user := User{
		ID:        slacker.ID,
		Name:      slacker.Name,
		RealName:  slacker.RealName,
		Deleted:   slacker.Deleted,
		Avatar:    slacker.Profile.ImageOriginal,
		CreatedAt: time.Now(),
	}

	_, err := db.NamedExec(`
		INSERT INTO users
		(id, name, real_name, deleted, avatar, created_at)
		VALUES
		(:id, :name, :real_name, :deleted, :avatar, :created_at)
		`, user)

	return err
}
