package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/tsaron/anansi/tokens"
	"golang.org/x/crypto/bcrypt"
	"tsaron.com/godview-starter/pkg/notification"
)

var (
	resetTokenDuration = time.Hour * 12

	ErrInvalidPassword   = errors.New("password is incorrect")
	ErrIncompleteProfile = errors.New("password has not been set")
)

type ResetToken struct {
	User      uint      `json:"user"`
	Workspace uint      `json:"workspace"`
	Key       string    `json:"-"`
	Expires   time.Time `json:"-"`
}

func ValidatePassword(password string, hash []byte) error {
	if len(hash) == 0 {
		return ErrIncompleteProfile
	}

	if err := bcrypt.CompareHashAndPassword(hash, []byte(password)); err != nil {
		return ErrInvalidPassword
	}

	return nil
}

func NewResetToken(ctx context.Context, tStore *tokens.Store, user *User) (ResetToken, error) {
	rToken := ResetToken{User: user.ID, Workspace: user.Workspace}

	var err error
	rToken.Key, err = tStore.Commission(ctx, resetTokenDuration, user.EmailAddress, rToken)
	if err != nil {
		return rToken, err
	}

	rToken.Expires = time.Now().Add(resetTokenDuration)

	return rToken, nil
}

func SendResetToken(mailer notification.Mailer, route string, token ResetToken, user *User) error {
	var day string
	if token.Expires.Day() == time.Now().Day() {
		day = "today"
	} else {
		day = "tomorrow"
	}

	data := struct {
		Route     string
		Token     string
		Expires   string
		FirstName string
	}{
		route,
		token.Key,
		fmt.Sprintf("%s %s", token.Expires.Format("3:04 pm"), day),
		user.FirstName,
	}
	return mailer.Send(notification.TemplateMail{
		Sender:        notification.SenderPostmaster,
		Subject:       "Reset your password",
		ReceiverName:  fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		ReceiverEmail: user.EmailAddress,
		Template:      "password-reset",
		TemplateData:  data,
	})
}
