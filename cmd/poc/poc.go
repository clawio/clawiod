package main

import (
	"flag"
	"fmt"
	"github.com/clawio/clawiod/root"
	"github.com/clawio/clawiod/root/authenticationmiddleware"
	"github.com/clawio/clawiod/root/authenticationwebservice"
	"github.com/clawio/clawiod/root/basicauthmiddleware"
	"github.com/clawio/clawiod/root/contextmanager"
	"github.com/clawio/clawiod/root/datawebservice"
	"github.com/clawio/clawiod/root/fileconfigurationsource"
	"github.com/clawio/clawiod/root/fsdatadriver"
	"github.com/clawio/clawiod/root/fsmdatadriver"
	"github.com/clawio/clawiod/root/jwttokendriver"
	"github.com/clawio/clawiod/root/loggermiddleware"
	"github.com/clawio/clawiod/root/memuserdriver"
	"github.com/clawio/clawiod/root/metadatawebservice"
	"github.com/clawio/clawiod/root/mimeguesser"
	"github.com/clawio/clawiod/root/ocfsdatadriver"
	"github.com/clawio/clawiod/root/ocfsmdatadriver"
	"github.com/clawio/clawiod/root/ocwebservice"
	"github.com/clawio/clawiod/root/remotedatadriver"
	"github.com/clawio/clawiod/root/remotemdatadriver"
	"github.com/clawio/clawiod/root/weberrorconverter"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/levels"
	"github.com/gorilla/mux"
	"github.com/iris-contrib/errors"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

var flagConfigurationSource string

func init() {
	flag.StringVar(&flagConfigurationSource, "configuraton", "file:clawio.conf", "Configuration source where to obtain the configuration")
	flag.Parse()
}

func main() {
	configurationSource, err := getConfigurationSource(flagConfigurationSource)
	if err != nil {
		fmt.Println(err)
		fmt.Println("can not instantiate configuration source")
		os.Exit(1)
	}
	config, err := configurationSource.LoadConfiguration()
	if err != nil {
		fmt.Println(err)
		fmt.Println("can not load configuration")
		os.Exit(1)
	}

	logger, err := getLogger(config)
	if err != nil {
		fmt.Println(err)
		fmt.Println("can not instantiate logger")
		os.Exit(1)
	}

	mainLogger := logger.With("pkg", "main")

	loggerMiddleware, err := getLoggerMiddleware(config)
	if err != nil {
		mainLogger.Error().Log("error", err)
		os.Exit(1)
	}

	authenticationWebService, err := getAuthenticationWebService(config)
	if err != nil {
		mainLogger.Error().Log("error", err)
		os.Exit(1)
	}

	dataWebService, err := getDataWebService(config)
	if err != nil {
		mainLogger.Error().Log("error", err)
		os.Exit(1)
	}

	metaDataWebService, err := getMetaDataWebService(config)
	if err != nil {
		mainLogger.Error().Log("error", err)
		os.Exit(1)
	}

	ownCloudWebService, err := getOCWebService(config)
	if err != nil {
		mainLogger.Error().Log("error", err)
		os.Exit(1)
	}

	webServices := []root.WebService{
		authenticationWebService,
		dataWebService,
		metaDataWebService,
		ownCloudWebService}

	router := mux.NewRouter()

	for _, service := range webServices {
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

	hostname, err := os.Hostname()
	if err != nil {
		logger.Error().Log("error", err)
		os.Exit(1)
	}

	mainLogger.Info().Log("msg", "server is listening", "port", config.GetPort(), "url", fmt.Sprintf("http://%s:%d", hostname, config.GetPort()))
	mainLogger.Error().Log("error", http.ListenAndServe(fmt.Sprintf(":%d",config.GetPort()), router))
}

func getUserDriver(config root.Configuration) (root.UserDriver, error) {
	switch config.GetUserDriver() {
	case "memuserdriver":
		return memuserdriver.New(config.GetMemUserDriverUsers()), nil
	default:
		return nil, errors.New("configured user driver does not exist")
	}
}

func getTokenDriver(config root.Configuration) (root.TokenDriver, error) {
	switch config.GetTokenDriver() {
	case "jwttokendriver":
		cm, err := getContextManager(config)
		if err != nil {
			return nil, err
		}
		logger, err := getLogger(config)
		if err != nil {
			return nil, err
		}
		return jwttokendriver.New(config.GetJWTTokenDriverKey(), cm, logger), nil
	default:
		return nil, errors.New("configured token driver does not exist")
	}

}

func getDataDriver(config root.Configuration) (root.DataDriver, error) {
	switch config.GetDataDriver() {
	case "fsdatadriver":
		logger, err := getLogger(config)
		if err != nil {
			return nil, err
		}
		return fsdatadriver.New(
			logger,
			config.GetFSDataDriverDataFolder(),
			config.GetFSDataDriverTemporaryFolder(),
			config.GetFSDataDriverChecksum(),
			config.GetFSDataDriverVerifyClientChecksum())
	case "ocfsdatadriver":
		logger, err := getLogger(config)
		if err != nil {
			return nil, err
		}
		metaDataDriver, err := getMetaDataDriver(config)
		if err != nil {
			return nil, err
		}
		return ocfsdatadriver.New(logger,
			config.GetOCFSDataDriverDataFolder(),
			config.GetOCFSDataDriverDataFolder(),
			config.GetOCFSDataDriverChecksum(),
			config.GetOCFSDataDriverVerifyClientChecksum(),
			metaDataDriver)
	case "remote":
		logger, err := getLogger(config)
		if err != nil {
			return nil, err
		}
		cm, err := getContextManager(config)
		if err != nil {
			return nil, err
		}
		return remotedatadriver.New(logger, cm, config.GetRemoteDataDriverURL()), nil
	default:
		return nil, errors.New("configured datadriver does not exist")

	}

}

func getMetaDataDriver(config root.Configuration) (root.MetaDataDriver, error) {
	switch config.GetMetaDataDriver() {
	case "fsmdatadriver":
		logger, err := getLogger(config)
		if err != nil {
			return nil, err
		}
		return fsmdatadriver.New(
			logger,
			config.GetFSMDataDriverDataFolder(),
			config.GetFSMDataDriverTemporaryFolder())
	case "ocfsmdatadriver":
		logger, err := getLogger(config)
		if err != nil {
			return nil, err
		}
		return ocfsmdatadriver.New(logger,
			nil,
			config.GetOCFSMDataDriverMaxSQLIddle(),
			config.GetOCFSMDataDriverMaxSQLConcurrent(),
			config.GetOCFSMDataDriverDataFolder(),
			config.GetOCFSMDataDriverTemporaryFolder(),
			config.GetOCFSMDataDriverDSN())
	case "remotemdatadriver":
		logger, err := getLogger(config)
		if err != nil {
			return nil, err
		}
		cm, err := getContextManager(config)
		if err != nil {
			return nil, err
		}
		return remotemdatadriver.New(logger, cm, config.GetRemoteMDataDriverURL()), nil
	default:
		return nil, errors.New("configured metadata driver does not exist")
	}
}

func getContextManager(config root.Configuration) (root.ContextManager, error) {
	// only one
	return contextmanager.New(), nil
}

func getMimeGuesser(config root.Configuration) (root.MimeGuesser, error) {
	return mimeguesser.New(), nil
}

func getAuthenticationMiddleware(config root.Configuration) (root.AuthenticationMiddleware, error) {
	cm, err := getContextManager(config)
	if err != nil {
		return nil, err
	}
	tokenDriver, err := getTokenDriver(config)
	if err != nil {
		return nil, err
	}
	return authenticationmiddleware.New(cm, tokenDriver), nil
}

func getBasicAuthMiddleware(config root.Configuration) (root.BasicAuthMiddleware, error) {
	cm, err := getContextManager(config)
	if err != nil {
		return nil, err
	}
	tokenDriver, err := getTokenDriver(config)
	if err != nil {
		return nil, err
	}
	userDriver, err := getUserDriver(config)
	if err != nil {
		return nil, err
	}
	return basicauthmiddleware.New(cm, userDriver, tokenDriver, config.GetBasicAuthMiddlewareCookieName()), nil
}

func getLogger(config root.Configuration) (levels.Levels, error) {
	var out io.Writer
	switch config.GetAppLoggerOut() {
	case "1":
		out = os.Stdout
	case "2":
		out = os.Stderr
	case "":
		out = ioutil.Discard
	default:
		out = &lumberjack.Logger{Filename: config.GetAppLoggerOut()}
	}
	l := log.NewLogfmtLogger(log.NewSyncWriter(out))
	l = log.NewContext(l).With("ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)
	return levels.New(l), nil
}

func getLoggerMiddleware(config root.Configuration) (root.LoggerMiddleware, error) {
	logger, err := getLogger(config)
	if err != nil {
		return nil, err
	}
	cm, err := getContextManager(config)
	if err != nil {
		return nil, err
	}
	return loggermiddleware.New(cm, logger), nil
}

func getAuthenticationWebService(config root.Configuration) (root.WebService, error) {
	logger, err := getLogger(config)
	if err != nil {
		return nil, err
	}
	cm, err := getContextManager(config)
	if err != nil {
		return nil, err
	}
	userDriver, err := getUserDriver(config)
	if err != nil {
		return nil, err
	}
	tokenDriver, err := getTokenDriver(config)
	if err != nil {
		return nil, err
	}
	authenticationMiddleware, err := getAuthenticationMiddleware(config)
	if err != nil {
		return nil, err
	}
	webErrorConverter, err := getWebErrorConverter(config)
	if err != nil {
		return nil, err
	}
	return authenticationwebservice.New(cm,
		logger,
		userDriver,
		tokenDriver,
		authenticationMiddleware,
		webErrorConverter), nil
}

func getDataWebService(config root.Configuration) (root.WebService, error) {
	logger, err := getLogger(config)
	if err != nil {
		return nil, err
	}
	cm, err := getContextManager(config)
	if err != nil {
		return nil, err
	}
	dataDriver, err := getDataDriver(config)
	if err != nil {
		return nil, err
	}
	authenticationMiddleware, err := getAuthenticationMiddleware(config)
	if err != nil {
		return nil, err
	}
	webErrorConverter, err := getWebErrorConverter(config)
	if err != nil {
		return nil, err
	}
	return datawebservice.New(cm,
		logger,
		dataDriver,
		authenticationMiddleware,
		webErrorConverter,
		config.GetDataWebServiceMaxUploadFileSize()), nil
}

func getMetaDataWebService(config root.Configuration) (root.WebService, error) {
	logger, err := getLogger(config)
	if err != nil {
		return nil, err
	}
	cm, err := getContextManager(config)
	if err != nil {
		return nil, err
	}
	metaDataDriver, err := getMetaDataDriver(config)
	if err != nil {
		return nil, err
	}
	authenticationMiddleware, err := getAuthenticationMiddleware(config)
	if err != nil {
		return nil, err
	}
	webErrorConverter, err := getWebErrorConverter(config)
	if err != nil {
		return nil, err
	}
	return metadatawebservice.New(
		cm,
		logger,
		metaDataDriver,
		authenticationMiddleware,
		webErrorConverter,
	), nil
}

func getOCWebService(config root.Configuration) (root.WebService, error) {
	logger, err := getLogger(config)
	if err != nil {
		return nil, err
	}
	cm, err := getContextManager(config)
	if err != nil {
		return nil, err
	}
	dataDriver, err := getDataDriver(config)
	if err != nil {
		return nil, err
	}
	metaDataDriver, err := getMetaDataDriver(config)
	if err != nil {
		return nil, err
	}
	webErrorConverter, err := getWebErrorConverter(config)
	if err != nil {
		return nil, err
	}
	mimeGuesser, err := getMimeGuesser(config)
	if err != nil {
		return nil, err
	}
	basicAuthMiddleware, err := getBasicAuthMiddleware(config)
	if err != nil {
		return nil, err
	}
	return ocwebservice.New(cm,
		logger,
		dataDriver,
		metaDataDriver,
		basicAuthMiddleware,
		webErrorConverter,
		mimeGuesser,
		config.GetOCWebServiceMaxUploadFileSize(),
		config.GetOCWebServiceChunksFolder()), nil
}

func getConfigurationSource(source string) (root.ConfigurationSource, error) {
	var protocol string
	var specific string
	parts := strings.Split(source, ":")
	if len(parts) == 0 {
		return nil, errors.New("configuration source is empty")
	}
	if len(parts) >= 2 {
		protocol = parts[0]
		specific = parts[1]
	} else {
		// default to file
		protocol = "file"
		specific = parts[0]

	}
	switch protocol {
	case "file":
		return fileconfigurationsource.New(specific)
	default:
		return nil, errors.New("configuration protocol does not exist")

	}

}

func getWebErrorConverter(config root.Configuration) (root.WebErrorConverter, error) {
	return weberrorconverter.New(), nil
}
