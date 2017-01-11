package ocbasicauthmiddleware

import (
	"github.com/clawio/clawiod/root"
	"net/http"
)

type middleware struct {
	cm          root.ContextManager
	userDriver  root.UserDriver
	tokenDriver root.TokenDriver
}

func New(cm root.ContextManager, userDriver root.UserDriver, tokenDriver root.TokenDriver) root.OwnCloudBasicAuthMiddleware {
	return &middleware{cm: cm, userDriver: userDriver, tokenDriver: tokenDriver}
}

func (m *middleware) HandlerFunc(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := m.cm.MustGetLog(r.Context())

		// try to get token from cookie
		authCookie, err := r.Cookie("OC_SessionPassphrase")
		if err == nil {
			user, err := m.tokenDriver.UserFromToken(authCookie.Value)
			if err == nil {
				r = r.WithContext(m.cm.SetUser(r.Context(), user))
				logger.Info().Log("user", user.Username())
				handler(w, r)
				return
			}
			logger.Warn().Log("err", "cookie token is invalid or not longer valid")

		} else {
			logger.Info().Log("msg", "cookie oc_sessionpassphrase not set")
		}

		// try to get credentials using basic auth
		username, password, ok := r.BasicAuth()
		if !ok {
			logger.Info().Log("msg", "basic auth not provided")
			w.Header().Set("WWW-Authenticate", "Basic Realm='clawio credentials'")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// try to authenticate user with username and password
		user, err := m.userDriver.GetByCredentials(username, password)
		if err != nil {
			logger.Error().Log("error", err)
			w.Header().Set("WWW-Authenticate", "Basic Realm='clawio credentials'")
			w.WriteHeader(http.StatusUnauthorized)
			return

		}

		token, err := m.tokenDriver.CreateToken(user)
		if err != nil {
			logger.Error().Log("error", err)
			w.Header().Set("WWW-Authenticate", "Basic Realm='clawio credentials'")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// save token into cookie for further requests
		cookie := &http.Cookie{}
		cookie.Name = "OC_SessionPassphrase"
		cookie.Value = token
		http.SetCookie(w, cookie)

		r = r.WithContext(m.cm.SetUser(r.Context(), user))
		logger.Info().Log("user", user.Username(), "msg", "request is authenticated")
		handler(w, r)
		return
	}
}
