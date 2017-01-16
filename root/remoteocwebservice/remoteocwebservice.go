package remoteocwebservice

import (
	"github.com/clawio/clawiod/root"
	"github.com/go-kit/kit/log/levels"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type service struct {
	logger levels.Levels
	url    string
	client *http.Client
	proxy  *httputil.ReverseProxy
}

func New(logger levels.Levels, urlString string) (root.WebService, error) {
	u, err := url.Parse(urlString)
	if err != nil {
		return nil, err
	}
	logger.Info().Log("msg", "reverse proxy configured to route requests to url", "url", u.String())
	proxy := httputil.NewSingleHostReverseProxy(u)
	return &service{
		logger: logger,
		url:    urlString,
		proxy:  proxy,
	}, nil
}

func (s *service) Endpoints() map[string]map[string]http.HandlerFunc {
	return map[string]map[string]http.HandlerFunc{
		"/ocwebdav/status.php": {
			"GET": s.statusEndpoint(),
		},
		"/ocwebdav/ocs/v1.php/cloud/capabilities": {
			"GET": s.capabilitiesEndpoint(),
		},
		"/ocwebdav/remote.php/webdav/{path:.*}": {
			"GET":       s.getEndpoint(),
			"PUT":       s.putEndpoint(),
			"OPTIONS":   s.optionsEndpoint(),
			"LOCK":      s.lockEndpoint(),
			"UNLOCK":    s.unlockEndpoint(),
			"HEAD":      s.headEndpoint(),
			"MKCOL":     s.mkcolEndpoint(),
			"PROPPATCH": s.proppatchEndpoint(),
			"PROPFIND":  s.propfindEndpoint(),
			"DELETE":    s.deleteEndpoint(),
			"MOVE":      s.moveEndpoint(),
		},
	}
}

func (s *service) statusEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Log("msg", "status request forwarded", "remote", s.url)
		s.proxy.ServeHTTP(w, r)
		return
	}
}
func (s *service) capabilitiesEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Log("msg", "capabilities request forwarded", "remote", s.url)
		s.proxy.ServeHTTP(w, r)
		return
	}
}
func (s *service) getEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Log("msg", "get request forwarded", "remote", s.url)
		s.proxy.ServeHTTP(w, r)
		return
	}
}

func (s *service) putEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Log("msg", "put request forwarded", "remote", s.url)
		s.proxy.ServeHTTP(w, r)
		return
	}
}

func (s *service) optionsEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Log("msg", "options request forwarded", "remote", s.url)
		s.proxy.ServeHTTP(w, r)
		return
	}
}

func (s *service) lockEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Log("msg", "lock request forwarded", "remote", s.url)
		s.proxy.ServeHTTP(w, r)
		return
	}
}
func (s *service) unlockEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Log("msg", "unlock request forwarded", "remote", s.url)
		s.proxy.ServeHTTP(w, r)
		return
	}
}
func (s *service) headEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Log("msg", "head request forwarded", "remote", s.url)
		s.proxy.ServeHTTP(w, r)
		return
	}
}
func (s *service) mkcolEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Log("msg", "mkcol request forwarded", "remote", s.url)
		s.proxy.ServeHTTP(w, r)
		return
	}
}
func (s *service) proppatchEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Log("msg", "proppatch request forwarded", "remote", s.url)
		s.proxy.ServeHTTP(w, r)
		return
	}
}
func (s *service) propfindEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Log("msg", "propfind request forwarded", "remote", s.url)
		s.proxy.ServeHTTP(w, r)
		return
	}
}
func (s *service) deleteEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Log("msg", "delete request forwarded", "remote", s.url)
		s.proxy.ServeHTTP(w, r)
		return
	}
}
func (s *service) moveEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Log("msg", "move request forwarded", "remote", s.url)
		s.proxy.ServeHTTP(w, r)
		return
	}
}
