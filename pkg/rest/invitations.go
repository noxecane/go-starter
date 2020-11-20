package rest

import (
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-chi/chi"
	ozzo "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/tsaron/anansi"
	"tsaron.com/godview-starter/pkg/config"
	"tsaron.com/godview-starter/pkg/invitations"
	"tsaron.com/godview-starter/pkg/notification"
	"tsaron.com/godview-starter/pkg/sessions"
	"tsaron.com/godview-starter/pkg/users"
)

var (
	isPhone        = regexp.MustCompile("0[789][01][0-9]{8,8}")
	errPhone       = ozzo.NewError("validation_is_phone", "must be a valid phone number(080xxxxxxxx)")
	phoneValidator = ozzo.NewStringRuleWithError(
		func(p string) bool {
			return isPhone.MatchString(p)
		},
		errPhone,
	)
)

type InvitationDTO struct {
	EmailAddress string `json:"email_address"`
	Role         string `json:"role"`
}

func (t *InvitationDTO) Validate() error {
	return ozzo.ValidateStruct(t,
		ozzo.Field(&t.EmailAddress, ozzo.Required, is.Email),
		ozzo.Field(&t.Role, ozzo.Required, ozzo.In("member", "admin")),
	)
}

type RegistrationDTO struct {
	CompanyName string `json:"company_name"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phone_number"`
}

func (t *RegistrationDTO) Validate() error {
	return ozzo.ValidateStruct(t,
		ozzo.Field(&t.CompanyName),
		ozzo.Field(&t.FirstName, ozzo.Required),
		ozzo.Field(&t.LastName, ozzo.Required),
		ozzo.Field(&t.Password, ozzo.Required, ozzo.Length(8, 64)),
		ozzo.Field(&t.PhoneNumber, ozzo.Required, phoneValidator),
	)
}

func Invitations(r *chi.Mux, app *config.App, sStore *sessions.Store, mailer notification.Mailer) {
	ivStore := invitations.NewStore(app.Tokens)
	uRepo := users.NewRepo(app.DB)

	r.Route("/invitations", func(r chi.Router) {
		r.Post("/", inviteUsers(app.Auth, uRepo, ivStore, app.Env, mailer))
		r.Patch("/{token}/extend", extendInvitation(ivStore))
		r.Patch("/{token}/accept", acceptInvitation(ivStore, uRepo, sStore))
	})
}

func extendInvitation(ivStore *invitations.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := anansi.StringParam(r, "token")

		iv, err := ivStore.Extend(r.Context(), token)
		if err != nil {
			if errors.Is(err, invitations.ErrExpired) {
				panic(anansi.APIError{
					Code:    http.StatusUnauthorized,
					Message: "Your invitation token has expired",
				})
			}
			panic(err)
		}

		anansi.SendSuccess(r, w, iv)
	}
}

func acceptInvitation(ivStore *invitations.Store, uRepo *users.Repo, sStore *sessions.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var dto RegistrationDTO
		anansi.ReadJSON(r, &dto)

		token := anansi.StringParam(r, "token")

		iv, err := ivStore.View(r.Context(), token)
		if err != nil {
			if errors.Is(err, invitations.ErrExpired) {
				panic(anansi.APIError{
					Code:    http.StatusUnauthorized,
					Message: "Your invitation token has expired",
				})
			}
			panic(err)
		}

		user, err := uRepo.Register(r.Context(), iv.EmailAddress, users.Registration{
			FirstName:   dto.FirstName,
			LastName:    dto.LastName,
			PhoneNumber: dto.PhoneNumber,
			Password:    dto.Password,
		})
		if err != nil {
			if errors.Is(err, users.ErrExistingPhoneNumber) {
				panic(anansi.APIError{
					Code:    http.StatusConflict,
					Message: err.Error(),
				})
			} else {
				panic(err)
			}
		}

		if err := ivStore.Revoke(r.Context(), user.EmailAddress); err != nil {
			panic(err)
		}

		session, err := sStore.Create(r.Context(), user)
		if err != nil {
			panic(err)
		}

		anansi.SendSuccess(r, w, session)
	}
}

func inviteUsers(auth *anansi.SessionStore, uRepo *users.Repo, ivStore *invitations.Store, env *config.Env, mailer notification.Mailer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var session sessions.Session
		auth.Load(r, &session)

		if session.Role == "member" {
			panic(anansi.APIError{
				Code:    http.StatusForbidden,
				Message: "You are not allowed to invite other users",
			})
		}

		var dtos []InvitationDTO
		anansi.ReadJSON(r, &dtos)

		// create the invited users
		var reqs []users.UserRequest
		for _, dto := range dtos {
			reqs = append(reqs, users.UserRequest{
				EmailAddress: strings.ToLower(dto.EmailAddress),
				Role:         dto.Role,
			})
		}
		ux, err := uRepo.CreateMany(r.Context(), session.Workspace, reqs)
		if err != nil {
			if errMail, ok := err.(users.ErrEmail); ok {
				panic(anansi.APIError{
					Code:    http.StatusConflict,
					Message: errMail.Error(),
				})
			} else {
				panic(err)
			}
		}

		// send them mail invitations
		var ivs []invitations.Invitation
		for _, u := range ux {
			iv, err := ivStore.Create(r.Context(), session.Workspace, session.CompanyName, u.EmailAddress)
			if err != nil {
				panic(err)
			}

			if err := invitations.SendInvitation(mailer, env.ClientUserPage, iv); err != nil {
				panic(err)
			}

			ivs = append(ivs, iv)
		}

		anansi.SendSuccess(r, w, ivs)
	}
}
