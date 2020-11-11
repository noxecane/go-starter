package invitations

import (
	"context"
	"time"

	"github.com/tsaron/anansi/tokens"
)

var ErrExpired = tokens.ErrTokenNotFound

type Invitation struct {
	Workspace    uint   `json:"workspace"`
	CompanyName  string `json:"company_name"`
	EmailAddress string `json:"email_address"`
	Token        string `json:"token"`
}

type Store struct {
	tStore *tokens.Store
}

func NewStore(tStore *tokens.Store) *Store {
	return &Store{tStore}
}

func (s *Store) Create(ctx context.Context, wkpID uint, wkpName string, email string) (Invitation, error) {
	iv := Invitation{
		Workspace:    wkpID,
		CompanyName:  wkpName,
		EmailAddress: email,
	}

	var err error
	iv.Token, err = s.tStore.Commission(ctx, time.Hour*48, email, iv)
	if err != nil {
		return Invitation{}, err
	}

	return iv, err
}

func (s *Store) Extend(ctx context.Context, token string) (Invitation, error) {
	var iv Invitation
	err := s.tStore.Extend(ctx, token, time.Hour, &iv)

	return iv, err
}

func (s *Store) View(ctx context.Context, token string) (Invitation, error) {
	var iv Invitation
	err := s.tStore.Peek(ctx, token, &iv)

	return iv, err
}

func (s *Store) Revoke(ctx context.Context, key string) error {
	return s.tStore.Revoke(ctx, key)
}
