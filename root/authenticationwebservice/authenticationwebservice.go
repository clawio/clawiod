package authenticationwebservice

import (
	"encoding/json"
	"net/http"

	"github.com/clawio/clawiod/root"
	"github.com/go-kit/kit/log/levels"
)

type service struct {
	cm          root.ContextManager
	logger      *levels.Levels
	userDriver  root.UserDriver
	tokenDriver root.TokenDriver
	am          root.AuthenticationMiddleware
	wec         root.WebErrorConverter
}

func New(
	cm root.ContextManager,
	logger *levels.Levels,
	userDriver root.UserDriver,
	tokenDriver root.TokenDriver,
	am root.AuthenticationMiddleware,
	wec root.WebErrorConverter) root.WebService {
	return &service{
		cm:          cm,
		logger:      logger,
		userDriver:  userDriver,
		tokenDriver: tokenDriver,
		am:          am,
		wec:         wec,
	}
}

func (s *service) Endpoints() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/auth/token": {
			"POST": s.tokenEndpoint,
		},
		"/auth/ping": {
			"GET": s.am.HandlerFunc(s.pingEndpoint),
		},
	}
}

func (s *service) tokenEndpoint(w http.ResponseWriter, r *http.Request) {
	logger := s.cm.MustGetLog(r.Context())

	if r.Body == nil {
		logger.Error().Log("err", "body is <nil>")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	req := &tokenRequest{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		logger.Error().Log("error", err)
		err = badRequestError("invalid json")
		jsonError, err := s.wec.ErrorToJSON(err)
		if err != nil {
			logger.Error().Log("error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonError)
		return
	}

	user, err := s.userDriver.GetByCredentials(req.Username, req.Password)
	if err != nil {
		s.handleTokenEndpointError(err, w, r)
		return
	}

	token, err := s.tokenDriver.CreateToken(user)
	if err != nil {
		s.handleTokenEndpointError(err, w, r)
		return
	}

	logger.Info().Log("msg", "token generated for user", "user", user.Username())
	res := &tokenResponse{AccessToken: token}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		logger.Error().Log("error", err)
	}

}

func (s *service) handleTokenEndpointError(err error, w http.ResponseWriter, r *http.Request) {
	logger := s.cm.MustGetLog(r.Context())
	logger.Error().Log("error", err)
	err = badRequestError("invalid authentication credentials")
	logger.Error().Log("error", err.Error())
	jsonErr, err := s.wec.ErrorToJSON(err)
	if err != nil {
		logger.Error().Log("error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	w.Write(jsonErr)
	return
}

func (s *service) pingEndpoint(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

type (
	tokenRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
)

type badRequestError string

func (e badRequestError) Error() string {
	return string(e)
}
func (e badRequestError) Code() root.Code {
	return root.Code(root.CodeBadInputData)
}
func (e badRequestError) Message() string {
	return string(e)
}
