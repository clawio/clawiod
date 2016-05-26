package authentication

import (
	"encoding/json"
	"net/http"

	"github.com/clawio/clawiod/codes"
	"github.com/clawio/clawiod/keys"
)

type (
	// TokenRequest specifies the data received by the Token endpoint.
	TokenRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// TokenResponse specifies the data returned from the Token endpoint.
	TokenResponse struct {
		AccessToken string `json:"access_token"`
	}
)

// Authenticate authenticates an user using an username and a password.
func (s *svc) Token(w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r)
	if r.Body == nil {
		log.Error("body is nil")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	authReq := &TokenRequest{}
	if err := json.NewDecoder(r.Body).Decode(authReq); err != nil {
		e := codes.NewErr(codes.BadInputData, "")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(e)
		return
	}
	token, err := s.authenticationController.Authenticate(authReq.Username, authReq.Password)
	if err != nil {
		s.handleTokenError(err, w, r)
		return
	}
	res := &TokenResponse{AccessToken: token}
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Error(err)
	}
}

func (s *svc) handleTokenError(err error, w http.ResponseWriter, r *http.Request) {
	log := keys.MustGetLog(r)
	e := codes.NewErr(codes.BadInputData, "user or password do not match")
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(e); err != nil {
		log.Error(err)
	}
	return
}
