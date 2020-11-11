package users

import (
	"context"
	"errors"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/tsaron/anansi/postgres"
	"golang.org/x/crypto/bcrypt"
)

const (
	RoleMember = "member"
	RoleAdmin  = "admin"
	RoleOwner  = "owner"
)

var ErrExistingPhoneNumber = errors.New("This phone number is already in use")

type ErrEmail string

func (e ErrEmail) Error() string {
	return string(e) + " has already been registered"
}

type Registration struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phone_number"`
}

type User struct {
	ID           uint      `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	FirstName    string    `json:"first_name,omitempty"`
	LastName     string    `json:"last_name,omitempty"`
	Role         string    `json:"role"`
	Password     []byte    `json:"-"`
	EmailAddress string    `json:"email_address"`
	PhoneNumber  string    `json:"phone_number,omitempty"`
	Workspace    uint      `json:"workspace"`
}

type UserRequest struct {
	EmailAddress string
	Role         string
}

type Repo struct {
	db *pg.DB
}

func NewRepo(db *pg.DB) *Repo {
	return &Repo{db}
}

func (r *Repo) Create(ctx context.Context, workspace uint, req UserRequest) (*User, error) {
	user := &User{
		EmailAddress: req.EmailAddress,
		Role:         req.Role,
		Workspace:    workspace,
	}

	_, err := r.db.
		ModelContext(ctx, user).
		Returning("*").
		Insert(user)

	if err != nil && postgres.ErrDuplicate.MatchString(err.Error()) {
		return nil, ErrEmail(req.EmailAddress)
	}

	return user, err
}

func (r *Repo) CreateMany(ctx context.Context, workspace uint, reqs []UserRequest) ([]User, error) {
	var users []User

	for _, req := range reqs {
		users = append(users, User{
			EmailAddress: req.EmailAddress,
			Role:         req.Role,
			Workspace:    workspace,
		})
	}

	_, err := r.db.
		ModelContext(ctx, &users).
		Returning("*").
		Insert(&users)

	if err != nil && postgres.ErrDuplicate.MatchString(err.Error()) {
		return nil, ErrEmail("One of the users")
	}

	return users, err
}

func (r *Repo) Register(ctx context.Context, email string, reg Registration) (*User, error) {
	pwdBytes, err := bcrypt.GenerateFromPassword([]byte(reg.Password), 10)
	if err != nil {
		return nil, err
	}

	user := &User{
		Password:    pwdBytes,
		FirstName:   reg.FirstName,
		LastName:    reg.LastName,
		PhoneNumber: reg.PhoneNumber,
	}

	_, err = r.db.
		ModelContext(ctx, user).
		Where("email_address = ?", email).
		Column("first_name", "last_name", "phone_number", "password").
		Returning("*").
		Update(user)

	if err != nil && postgres.ErrDuplicate.MatchString(err.Error()) {
		return nil, ErrExistingPhoneNumber
	}

	return user, err
}

func (r *Repo) ChangePassword(ctx context.Context, wkID, id uint, password string) (*User, error) {
	pwdBytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return nil, err
	}

	user := &User{Password: pwdBytes}
	_, err = r.db.
		ModelContext(ctx, user).
		Where("id = ?", id).
		Where("workspace = ?", wkID).
		Column("password").
		Returning("*").
		Update(user)

	if err == pg.ErrNoRows {
		return nil, nil
	}

	return user, err
}
