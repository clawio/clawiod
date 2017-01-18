package proxiedmetadatawebservice

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
		"/meta/examine": {
			"POST": s.examineEndpoint(),
		},
		"/meta/list": {
			"POST": s.listFolderEndpoint(),
		},
		"/meta/move": {
			"POST": s.moveEndpoint(),
		},
		"/meta/delete": {
			"POST": s.deleteEndpoint(),
		},
		"/meta/makefolder": {
			"POST": s.makeFolderEndpoint(),
		},
	}
}

func (s *service) examineEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Log("msg", "examine request forwarded", "remote", s.url)
		s.proxy.ServeHTTP(w, r)
		return
	}
}

func (s *service) listFolderEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Log("msg", "listfolder request forwarded", "remote", s.url)
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

func (s *service) deleteEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Log("msg", "delete request forwarded", "remote", s.url)
		s.proxy.ServeHTTP(w, r)
		return
	}
}

func (s *service) makeFolderEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info().Log("msg", "makefolder request forwarded", "remote", s.url)
		s.proxy.ServeHTTP(w, r)
		return
	}
}
