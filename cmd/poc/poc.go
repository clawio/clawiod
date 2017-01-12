package main

import (
	"github.com/clawio/clawiod/root"
	"github.com/clawio/clawiod/root/authenticationmiddleware"
	"github.com/clawio/clawiod/root/authenticationwebservice"
	"github.com/clawio/clawiod/root/contextmanager"
	"github.com/clawio/clawiod/root/datawebservice"
	"github.com/clawio/clawiod/root/jwttokendriver"
	"github.com/clawio/clawiod/root/loggermiddleware"
	"github.com/clawio/clawiod/root/memuserdriver"
	"github.com/clawio/clawiod/root/mimeguesser"
	"github.com/clawio/clawiod/root/ocbasicauthmiddleware"
	"github.com/clawio/clawiod/root/ocfsdatadriver"
	"github.com/clawio/clawiod/root/ocfsmdatadriver"
	"github.com/clawio/clawiod/root/ocwebservice"
	"github.com/clawio/clawiod/root/weberrorconverter"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/levels"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"os"
	"github.com/clawio/clawiod/root/metadatawebservice"
)

func main() {
	l := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	l = log.NewContext(l).With("ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	logger := levels.New(l)
	mainLogger := logger.With("pkg", "main")

	cm := contextmanager.New()
	userDriver := memuserdriver.New("demo:demo:demo@example.org:Super Demo")
	tokenDriver := jwttokendriver.New("shhhh", cm, logger.With("pkg", "jwttokendriver"))
	//dataDriver, err := fsdatadriver.New(logger.With("pkg", "fsdatadriver"), "/tmp/clawiod/", "/tmp/clawiod", "", false)
	//if err != nil {
	//	mainLogger.Error().Log("error", err)
	//	os.Exit(1)
	//}
	//metaDataDriver, err := fsmdatadriver.New(logger.With("pkg", "fsmdatadriver"), "/tmp/clawiod", "/tmp/clawiod")
	//if err != nil {
	//	mainLogger.Error().Log("error", err)
	//	os.Exit(1)
	//}

	ocMetaDataDriver, err := ocfsmdatadriver.New(logger, nil, 1064, 100, "/tmp/clawiod", "/tmp/clawiod", "root:passwd@tcp(192.168.99.100:32768)/owncloud")
	if err != nil {
		mainLogger.Error().Log("error", err)
		os.Exit(1)
	}
	ocDataDriver, err := ocfsdatadriver.New(logger, "/tmp/clawiod/", "/tmp/clawiod/", "", false, ocMetaDataDriver)
	if err != nil {
		mainLogger.Error().Log("error", err)
		os.Exit(1)
	}

	authMiddleware := authenticationmiddleware.New(cm, tokenDriver)
	loggerMiddleware := loggermiddleware.New(cm, logger.With("pkg", "loggermiddleware"))
	ocBasicAuthMiddleware := ocbasicauthmiddleware.New(cm, userDriver, tokenDriver)
	wec := weberrorconverter.New()
	mg := mimeguesser.New()

	authService := authenticationwebservice.New(cm, logger.With("pkg", "authenticationwebservice"), userDriver, tokenDriver, authMiddleware, wec)
	//dataService := datawebservice.New(cm, logger.With("pkg", "datawebservice"), dataDriver, authMiddleware, wec, 1000000000000)
	//metaDataService := metadatawebservice.New(cm, logger.With("pkg", "metadatawebservice"), metaDataDriver, authMiddleware, wec)
	dataService := datawebservice.New(cm, logger.With("pkg", "datawebservice"),ocDataDriver, authMiddleware, wec, 1000000000000)
	metaDataService := metadatawebservice.New(cm, logger.With("pkg", "metadatawebservice"),ocMetaDataDriver, authMiddleware, wec)
	ocService := ocwebservice.New(cm, logger.With("pkg", "ocwebservice"), ocDataDriver, ocMetaDataDriver, ocBasicAuthMiddleware, wec, mg, 1000000000000, "/tmp/chunks/")

	services := []root.WebService{authService, dataService, metaDataService, ocService}

	router := mux.NewRouter()

	for _, service := range services {
		for path, methods := range service.Endpoints() {
			for method, handlerFunc := range methods {
				handlerFunc = loggerMiddleware.HandlerFunc(handlerFunc)
				handlerFunc := http.HandlerFunc(handlerFunc)
				var handler http.Handler
				handler = handlerFunc
				router.Handle(path, handler).Methods(method)
				prometheus.InstrumentHandler(path, handler)
				mainLogger.Info().Log("method", method, "endpoint", path, "msg", "registered new endpoint")
			}
		}
	}

	mainLogger.Error().Log("error", http.ListenAndServe(":1515", router))
}
