package sessions

import (
	"context"
	"fmt"
	"time"

	"github.com/tsaron/anansi/tokens"
	"tsaron.com/godview-starter/pkg/users"
	"tsaron.com/godview-starter/pkg/workspaces"
)

type Session struct {
	Workspace   uint   `json:"workspace"`
	User        uint   `json:"user"`
	Role        string `json:"role"`
	CompanyName string `json:"company_name"`
	SessionKey  string `json:"session_key"`
	FullName    string `json:"full_name"`
}

type Store struct {
	tStore *tokens.Store
	wRepo  *workspaces.Repo
}

func NewStore(tStore *tokens.Store, wRepo *workspaces.Repo) *Store {
	return &Store{tStore, wRepo}
}

func (s *Store) Create(ctx context.Context, u *users.User) (Session, error) {
	workspace, err := s.wRepo.Get(ctx, u.Workspace)
	if err != nil {
		return Session{}, err
	}

	session := Session{
		Workspace:   u.Workspace,
		User:        u.ID,
		Role:        u.Role,
		CompanyName: workspace.CompanyName,
		FullName:    fmt.Sprintf("%s %s", u.FirstName, u.LastName),
	}

	token, err := s.tStore.Commission(ctx, time.Hour*24, u.EmailAddress, session)
	if err != nil {
		return Session{}, err
	}

	session.SessionKey = token
	return session, nil
}
