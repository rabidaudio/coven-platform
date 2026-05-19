package users

import (
	"errors"
	"net/http"

	"github.com/atlantacoven/coven-platform/member-site/api"
	"github.com/go-chi/chi/v5"
)

type PutSessionRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type PutSessionResponseBody struct {
	Id    int    `json:"id"`
	Token []byte `json:"token,omitempty"`
}

func Router(r chi.Router) {
	r.Post("/session", postSession)
}

func postSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body := PutSessionRequestBody{}
	if ok := api.UnmarshalBody(w, r, &body); !ok {
		return
	}

	u, err := AuthenticatePassword(ctx, body.Email, body.Password)
	if errors.Is(err, ErrInvalidPassword) || errors.Is(err, ErrNotFound) {
		api.RespondBadFormat(w, err)
		return
	} else if err != nil {
		api.RespondError(w, err)
		return
	}

	token, err := u.GenerateToken()
	if err != nil {
		api.RespondError(w, err)
		return
	}

	res := PutSessionResponseBody{
		Id:    u.Id,
		Token: token,
	}
	api.Respond(w, &res, "OK")
}
