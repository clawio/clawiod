package remotedatawebservice

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
		"/data/upload": {
			"POST": s.uploadEndpoint(),
		},
		"/data/download": {
			"POST": s.downloadEndpoint(),
		},
	}
}

func (s *service) uploadEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Log("msg", "upload request forwarded", "remote", s.url)
		s.proxy.ServeHTTP(w, r)
		return
	}
}

func (s *service) downloadEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Log("msg", "download request forwarded", "remote", s.url)
		s.proxy.ServeHTTP(w, r)
		return
	}
}

