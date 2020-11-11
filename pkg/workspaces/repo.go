package workspaces

import (
	"context"
	"time"

	"github.com/go-pg/pg/v10"
)

type Workspace struct {
	ID           uint      `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	CompanyName  string    `json:"company_name"`
	EmailAddress string    `json:"email_address"`
}

type Repo struct {
	db *pg.DB
}

func NewRepo(db *pg.DB) *Repo {
	return &Repo{db}
}

// Create a workspace
func (r *Repo) Create(ctx context.Context, name, email string) (*Workspace, error) {
	workspace := &Workspace{CompanyName: name, EmailAddress: email}

	_, err := r.db.
		ModelContext(ctx, workspace).
		Returning("*").
		Insert(workspace)

	return workspace, err
}

// Get returns the workspace with the given ID. Returns nil if the workspace doesn't exist
func (r *Repo) Get(ctx context.Context, id uint) (*Workspace, error) {
	workspace := &Workspace{ID: id}
	err := r.db.ModelContext(ctx, workspace).WherePK().Select()

	if err == pg.ErrNoRows {
		return nil, nil
	}

	return workspace, err
}

// ChangeName updates the name of a workspace.
func (r *Repo) ChangeName(ctx context.Context, id uint, name string) (*Workspace, error) {
	workspace := &Workspace{
		ID:          id,
		CompanyName: name,
	}

	_, err := r.db.
		ModelContext(ctx, workspace).
		WherePK().
		Column("company_name").
		Returning("*").
		Update(workspace)

	return workspace, err
}
