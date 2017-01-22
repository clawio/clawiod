package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/clawio/clawiod/root"
	"github.com/clawio/clawiod/root/authenticationmiddleware"
	"github.com/clawio/clawiod/root/authenticationwebservice"
	"github.com/clawio/clawiod/root/basicauthmiddleware"
	"github.com/clawio/clawiod/root/contextmanager"
	"github.com/clawio/clawiod/root/corsmiddleware"
	"github.com/clawio/clawiod/root/datawebservice"
	"github.com/clawio/clawiod/root/datawebserviceclient"
	"github.com/clawio/clawiod/root/etcdregistrydriver"
	"github.com/clawio/clawiod/root/fileconfigurationsource"
	"github.com/clawio/clawiod/root/fsdatadriver"
	"github.com/clawio/clawiod/root/fsmdatadriver"
	"github.com/clawio/clawiod/root/jwttokendriver"
	"github.com/clawio/clawiod/root/ldapuserdriver"
	"github.com/clawio/clawiod/root/loggermiddleware"
	"github.com/clawio/clawiod/root/memuserdriver"
	"github.com/clawio/clawiod/root/metadatawebservice"
	"github.com/clawio/clawiod/root/metadatawebserviceclient"
	"github.com/clawio/clawiod/root/mimeguesser"
	"github.com/clawio/clawiod/root/ocfsdatadriver"
	"github.com/clawio/clawiod/root/ocfsmdatadriver"
	"github.com/clawio/clawiod/root/ocwebservice"
	"github.com/clawio/clawiod/root/proxiedauthenticationwebservice"
	"github.com/clawio/clawiod/root/proxieddatawebservice"
	"github.com/clawio/clawiod/root/proxiedmetadatawebservice"
	"github.com/clawio/clawiod/root/proxiedocwebservice"
	"github.com/clawio/clawiod/root/remoteocwebservice"
	"github.com/clawio/clawiod/root/weberrorconverter"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/levels"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
)

var flagConfigurationSource string

func init() {
	flag.StringVar(&flagConfigurationSource, "conf", "file:clawiod.conf", "Configuration source where to obtain the configuration")
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

	// Set CPU capacity
	err = setCPU(config.GetCPU())
	if err != nil {
		mainLogger.Crit().Log("msg", "error tweaking cpu", "error", err)
		os.Exit(1)
	}

	server, err := newServer(config)
	if err != nil {
		mainLogger.Error().Log("error", err)
		os.Exit(1)
	}

	hostname, err := os.Hostname()
	if err != nil {
		logger.Error().Log("error", err)
		os.Exit(1)
	}

	addr := fmt.Sprintf(":%d", config.GetPort())
	if config.IsTLSEnabled() {
		mainLogger.Info().Log("msg", "serving secure client requests", "addr", fmt.Sprintf("https://%s:%d", hostname, config.GetPort()))
		mainLogger.Error().Log("error", http.ListenAndServeTLS(
			addr,
			config.GetTLSCertificate(),
			config.GetTLSPrivateKey(),
			server))
	} else {
		mainLogger.Warn().Log("msg", "serving insecure client requests", "addr", fmt.Sprintf("http://%s:%d", hostname, config.GetPort()))
		mainLogger.Error().Log("error", http.ListenAndServe(
			addr,
			server))
	}
}

func getUserDriver(config root.Configuration) (root.UserDriver, error) {
	switch config.GetUserDriver() {
	case "memuserdriver":
		return memuserdriver.New(config.GetMemUserDriverUsers()), nil
	case "ldapuserdriver":
		logger, err := getLogger(config)
		if err != nil {
			return nil, err
		}
		return ldapuserdriver.New(logger,
			config.GetLDAPUserDriverBindUsername(),
			config.GetLDAPUserDriverBindPassword(),
			config.GetLDAPUserDriverHostname(),
			config.GetLDAPUserDriverPort(),
			config.GetLDAPUserDriverBaseDN(),
			config.GetLDAPUserDriverFilter())
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
			config.GetOCFSDataDriverTemporaryFolder(),
			config.GetOCFSDataDriverChunksFolder(),
			config.GetOCFSDataDriverChecksum(),
			config.GetOCFSDataDriverVerifyClientChecksum(),
			metaDataDriver)
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
		out = &lumberjack.Logger{
			Filename:   config.GetAppLoggerOut(),
			MaxSize:    config.GetAppLoggerMaxSize(),
			MaxAge:     config.GetAppLoggerMaxAge(),
			MaxBackups: config.GetAppLoggerMaxBackups()}
	}
	hostname, err := os.Hostname()
	if err != nil {
		return levels.Levels{}, err
	}
	hostname = fmt.Sprintf("%s:%d", hostname, config.GetPort())
	l := log.NewLogfmtLogger(log.NewSyncWriter(out))
	l = log.NewContext(l).With("ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller, "host", hostname)
	return levels.New(l), nil
}

func getHTTPLogger(config root.Configuration) (io.Writer, error) {
	var out io.Writer
	switch config.GetAppLoggerOut() {
	case "1":
		out = os.Stdout
	case "2":
		out = os.Stderr
	case "":
		out = ioutil.Discard
	default:
		out = &lumberjack.Logger{
			Filename:   config.GetHTTPAccessLoggerOut(),
			MaxAge:     config.GetHTTPAccessLoggerMaxAge(),
			MaxBackups: config.GetHTTPAccessLoggerMaxBackups(),
			MaxSize:    config.GetHTTPAccessLoggerMaxSize(),
		}
	}
	return out, nil
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
	switch config.GetAuthenticationWebService() {
	case "local":
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
			webErrorConverter,
			false), nil
	case "proxied":
		logger, err := getLogger(config)
		if err != nil {
			return nil, err
		}
		return proxiedauthenticationwebservice.New(logger, config.GetProxiedAuthenticationWebServiceURL())
	default:
		return nil, errors.New("configured authentication web service does not exist")

	}
}

func getDataWebService(config root.Configuration) (root.WebService, error) {
	switch config.GetDataWebService() {
	case "local":
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
	case "proxied":
		logger, err := getLogger(config)
		if err != nil {
			return nil, err
		}
		return proxieddatawebservice.New(logger, config.GetProxiedDataWebServiceURL())

	default:
		return nil, errors.New("configured data webservice does not exist")

	}
}

func getMetaDataWebService(config root.Configuration) (root.WebService, error) {
	switch config.GetMetaDataWebService() {
	case "local":
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
	case "proxied":
		logger, err := getLogger(config)
		if err != nil {
			return nil, err
		}
		return proxiedmetadatawebservice.New(logger, config.GetProxiedMetaDataWebServiceURL())
	default:
		return nil, errors.New("configured metadata webservice does not exist")
	}
}

func getOCWebService(config root.Configuration) (root.WebService, error) {
	switch config.GetOCWebService() {
	case "local":
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
			config.GetOCWebServiceMaxUploadFileSize()), nil
	case "proxied":
		logger, err := getLogger(config)
		if err != nil {
			return nil, err
		}
		return proxiedocwebservice.New(logger, config.GetProxiedOCWebServiceURL())
	case "remote":
		logger, err := getLogger(config)
		if err != nil {
			return nil, err
		}
		cm, err := getContextManager(config)
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
		registryDriver, err := getRegistryDriver(config)
		if err != nil {
			return nil, err
		}
		dataWebServiceClient := datawebserviceclient.New(logger, cm, registryDriver)
		metaDataWebServiceClient := metadatawebserviceclient.New(logger, cm, registryDriver)
		return remoteocwebservice.New(cm,
			logger,
			dataWebServiceClient,
			metaDataWebServiceClient,
			basicAuthMiddleware,
			webErrorConverter,
			mimeGuesser,
			config.GetRemoteOCWebServiceMaxUploadFileSize()), nil
	default:
		return nil, errors.New("configured oc webservice does not exist")

	}
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

func getRegistryDriver(config root.Configuration) (root.RegistryDriver, error) {
	logger, err := getLogger(config)
	if err != nil {
		return nil, err
	}
	logger = logger.With("pkg", "etcdregistrydriver")

	switch config.GetRegistryDriver() {
	case "etcd":
		return etcdregistrydriver.New(
			logger,
			config.GetETCDRegistryDriverUrls(),
			config.GetETCDRegistryDriverKey(),
			config.GetETCDRegistryDriverUsername(),
			config.GetETCDRegistryDriverPassword())
	default:
		return nil, fmt.Errorf("registry driver does not exist")
	}
}

func getWebErrorConverter(config root.Configuration) (root.WebErrorConverter, error) {
	return weberrorconverter.New(), nil
}

func getCORSMiddleware(config root.Configuration) (root.CorsMiddleware, error) {
	logger, err := getLogger(config)
	if err != nil {
		return nil, err
	}

	return corsmiddleware.New(
		logger.With("pkg", "corsmiddleware"),
		config.GetCORSMiddlewareAccessControlAllowOrigin(),
		config.GetCORSMiddlewareAccessControlAllowMethods(),
		config.GetCORSMiddlewareAccessControlAllowHeaders()), nil
}

func find(needle string, haystack []string) bool {
	for _, v := range haystack {
		if v == needle {
			return true
		}
	}
	return false
}

func getWebServices(config root.Configuration) (map[string]root.WebService, error) {
	enabledWebServices := strings.Split(config.GetEnabledWebServices(), ",")
	webServices := map[string]root.WebService{}
	if find("authentication", enabledWebServices) {
		authenticationWebService, err := getAuthenticationWebService(config)
		if err != nil {
			return nil, err
		}
		webServices["authentication"] = authenticationWebService
	}

	if find("data", enabledWebServices) {
		dataWebService, err := getDataWebService(config)
		if err != nil {
			return nil, err
		}
		webServices["data"] = dataWebService
	}

	if find("metadata", enabledWebServices) {
		metaDataWebService, err := getMetaDataWebService(config)
		if err != nil {
			return nil, err
		}
		webServices["metadata"] = metaDataWebService
	}

	if find("owncloud", enabledWebServices) {
		ownCloudWebService, err := getOCWebService(config)
		if err != nil {
			return nil, err
		}
		webServices["owncloud"] = ownCloudWebService
	}
	return webServices, nil
}

// setCPU parses string cpu and sets GOMAXPROCS
// according to its value. It accepts either
// a number (e.g. 3) or a percent (e.g. 50%).
func setCPU(cpu string) error {
	var numCPU int

	availCPU := runtime.NumCPU()

	if strings.HasSuffix(cpu, "%") {
		// Percent
		var percent float32
		pctStr := cpu[:len(cpu)-1]
		pctInt, err := strconv.Atoi(pctStr)
		if err != nil || pctInt < 1 || pctInt > 100 {
			return errors.New("invalid CPU value: percentage must be between 1-100")
		}
		percent = float32(pctInt) / 100
		numCPU = int(float32(availCPU) * percent)
	} else {
		// Number
		num, err := strconv.Atoi(cpu)
		if err != nil || num < 1 {
			return errors.New("invalid CPU value: provide a number or percent greater than 0")
		}
		numCPU = num
	}

	if numCPU > availCPU {
		numCPU = availCPU
	}

	runtime.GOMAXPROCS(numCPU)
	return nil
}
