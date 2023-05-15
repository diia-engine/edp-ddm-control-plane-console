package main

import (
	"context"
	"ddm-admin-console/app/cluster"
	"ddm-admin-console/app/dashboard"
	"ddm-admin-console/app/registry"
	oauth "ddm-admin-console/auth"
	"ddm-admin-console/config"
	codebaseController "ddm-admin-console/controller/codebase"
	mergeRequestController "ddm-admin-console/controller/merge_request"
	"ddm-admin-console/mocks"
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	edpComponent "ddm-admin-console/service/edp_component"
	"ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/jenkins"
	"ddm-admin-console/service/k8s"
	"ddm-admin-console/service/keycloak"
	"ddm-admin-console/service/openshift"
	"ddm-admin-console/service/permissions"
	"ddm-admin-console/service/vault"
	"encoding/gob"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/leonelquinteros/gotext"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/oauth2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	configPath string
	cachePath  string
)

func main() {
	flag.StringVar(&configPath, "c", "default.env", "config file path")
	flag.StringVar(&cachePath, "ch", "cache.db", "cache file path")
	flag.Parse()

	cnf, err := loadConfig(configPath)
	if err != nil {
		panic(err)
	}

	logger, err := getLogger(cnf.LogLevel, cnf.LogEncoding)
	if err != nil {
		panic(err)
	}

	buildInfo := config.BuildInfoGet()
	logger.Sugar().Infow("starting the console",
		"version", buildInfo.Version,
		"git-commit", buildInfo.GitCommit,
		"git-tag", buildInfo.GitTag,
		"build-date", buildInfo.BuildDate,
		"go-version", buildInfo.Go,
		"platform", buildInfo.Platform,
	)

	logger.Info("init gin router")
	gin.SetMode(cnf.GinMode)
	r := gin.New()
	r.SetFuncMap(template.FuncMap{"i18n": i18n})
	r.LoadHTMLGlob("templates/**/*")
	r.Static("/static", "./static")
	r.Static("/assets", "./frontend/dist/assets")
	store := cookie.NewStore([]byte(cnf.SessionSecret))
	r.Use(sessions.Sessions("cookie-session", store))

	logger.Info("init apps")
	if err := initApps(logger, cnf, r, buildInfo.Date()); err != nil {
		panic(fmt.Sprintf("%+v", err))
	}

	logger.Info("init i18n")
	gotext.Configure("locale", "uk_UA", "default")

	logger.Info("run router on port", zap.String("port", cnf.HTTPPort))
	if err := r.Run(fmt.Sprintf(":%s", cnf.HTTPPort)); err != nil {
		panic(err)
	}
}

func exitWait(sigs chan os.Signal, appCache *cache.Cache) {
	<-sigs

	if err := saveCache(appCache); err != nil {
		panic(err)
	}

	os.Exit(0)
}

func saveCache(appCache *cache.Cache) error {
	if err := appCache.SaveFile(cachePath); err != nil {
		return fmt.Errorf("unable to save cache to file")
	}

	return nil
}

func loadConfig(path string) (*config.Settings, error) {
	if err := godotenv.Load(path); err != nil {
		return nil, fmt.Errorf("unable to load config file, %w", err)
	}

	var cnf config.Settings
	if err := envconfig.Process("", &cnf); err != nil {
		return nil, fmt.Errorf("unable to parse env variables, %w", err)
	}

	return &cnf, nil
}

func getLogger(level, encoding string) (*zap.Logger, error) {
	levels := map[string]zapcore.Level{
		"DEBUG":   zap.DebugLevel,
		"INFO":    zap.InfoLevel,
		"WARNING": zap.WarnLevel,
		"ERROR":   zap.ErrorLevel,
		"DPANIC":  zap.DPanicLevel,
		"PANIC":   zap.PanicLevel,
		"FATAL":   zap.FatalLevel,
	}

	logLevel, ok := levels[level]
	if !ok {
		logLevel = zap.InfoLevel
	}

	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(logLevel),
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         encoding,
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to build logger, %w", err)
	}

	return logger, nil
}

func initServices(sch *runtime.Scheme, restConf *rest.Config, appConf *config.Settings, logger *zap.Logger) (*config.Services, error) {
	if appConf.Mock != "" {
		return mocks.InitServices(appConf), nil
	}

	var err error
	serviceItems := config.Services{}

	serviceItems.EDPComponent, err = edpComponent.Make(sch, restConf, appConf.Namespace)
	if err != nil {
		return nil, fmt.Errorf("unable to init edp component service, %w", err)
	}

	serviceItems.Codebase, err = codebase.Make(sch, restConf, appConf.Namespace)
	if err != nil {
		return nil, fmt.Errorf("unable to init codebase service, %w", err)
	}

	serviceItems.K8S, err = k8s.Make(restConf, appConf.Namespace)
	if err != nil {
		return nil, fmt.Errorf("unable to init k8s service, %w", err)
	}

	serviceItems.Jenkins, err = jenkins.Make(sch, restConf, serviceItems.K8S,
		jenkins.Config{Namespace: appConf.Namespace, APIUrl: appConf.JenkinsAPIURL,
			AdminSecretName: appConf.JenkinsAdminSecretName})
	if err != nil {
		return nil, fmt.Errorf("unable to init jenkins service, %w", err)
	}

	serviceItems.OpenShift, err = openshift.Make(restConf, serviceItems.K8S)
	if err != nil {
		return nil, fmt.Errorf("unable to init open shift service, %w", err)
	}

	serviceItems.Gerrit, err = gerrit.Make(sch, restConf,
		gerrit.Config{Namespace: appConf.Namespace, GerritAPIUrlTemplate: appConf.GerritAPIUrlTemplate,
			RootGerritName: appConf.RootGerritName})
	if err != nil {
		return nil, fmt.Errorf("unable to create gerrit service, %w", err)
	}

	serviceItems.Keycloak, err = keycloak.Make(sch, restConf, appConf.UsersNamespace)
	if err != nil {
		return nil, fmt.Errorf("unable to create keycloak service, %w", err)
	}

	serviceItems.Vault, err = vault.Make(appConf.VaultConfig(), serviceItems.K8S)
	if err != nil {
		return nil, fmt.Errorf("unable to init vault service, %w", err)
	}

	serviceItems.Cache = cache.New(time.Hour, time.Minute)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go exitWait(sigs, serviceItems.Cache)

	gob.Register([]registry.CachedFile{})
	if err := serviceItems.Cache.LoadFile(cachePath); err != nil {
		logger.Warn("unable to load cache")
	}

	serviceItems.PermService = permissions.Make(serviceItems.Codebase, serviceItems.K8S)

	return &serviceItems, nil
}

func initControllers(sch *runtime.Scheme, namespace string, logger *zap.Logger, cnf *config.Settings,
	services *config.Services) error {
	if cnf.Mock != "" {
		return nil
	}

	cfg := ctrl.GetConfigOrDie()

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             sch,
		Namespace:          namespace,
		MetricsBindAddress: "0",
	})

	if err != nil {
		return fmt.Errorf("unable to ini manager, %w", err)
	}

	l := logger.Sugar()

	if err := codebaseController.Make(mgr, l, cnf, services.Cache, services.Gerrit, services.Codebase); err != nil {
		return fmt.Errorf("unable to init codebase controller, %w", err)
	}

	if err := mergeRequestController.Make(mgr, l, cnf, services.Gerrit, services.Codebase, services.Jenkins,
		services.Cache); err != nil {
		return fmt.Errorf("unable to init merge request controller, %w", err)
	}

	go func() {
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			logger.Sugar().Error(err.Error(), "unable to start manager")
		}
	}()

	return nil
}

func initApps(logger *zap.Logger, cnf *config.Settings, r *gin.Engine, buildTime time.Time) error {
	restConf, err := initKubeConfig()
	if err != nil {
		return fmt.Errorf("unable to init kube config, %w", err)
	}

	appRouter := router.Make(r, logger, buildTime)

	sch := runtime.NewScheme()
	if err := v1.AddToScheme(sch); err != nil {
		return fmt.Errorf("unable to add core api to scheme, %w", err)
	}

	serviceItems, err := initServices(sch, restConf, cnf, logger)
	if err != nil {
		return fmt.Errorf("unable to init services, %w", err)
	}

	if err := initControllers(sch, cnf.Namespace, logger, cnf, serviceItems); err != nil {
		return fmt.Errorf("unable to init controllers, %w", err)
	}

	oa, err := initOauth(restConf, cnf, r, serviceItems.K8S)
	if err != nil {
		return fmt.Errorf("unable to init oauth, %w", err)
	}

	_, err = dashboard.Make(appRouter, oa, serviceItems, cnf.ClusterCodebaseName)
	if err != nil {
		return fmt.Errorf("unable to make dashboard app, %w", err)
	}

	_, err = registry.Make(appRouter, serviceItems.RegistryServices(), cnf.RegistryConfig())
	if err != nil {
		return fmt.Errorf("unable to make registry app, %w", err)
	}

	_, err = cluster.Make(appRouter, serviceItems.ClusterServices(), cnf.ClusterConfig(), serviceItems.Cache)
	if err != nil {
		return fmt.Errorf("unable to init cluster app, %w", err)
	}

	return nil
}

func initKubeConfig() (*rest.Config, error) {
	k8sConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	restConfig, err := k8sConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to get k8s client config, %w", err)
	}

	return restConfig, nil
}

func initOauth(k8sConfig *rest.Config, cfg *config.Settings, r *gin.Engine, k8sService k8s.ServiceInterface) (dashboard.OAuth, error) {
	var oAuth dashboard.OAuth
	if cfg.Mock != "" {
		oAuth = mocks.OAuth()
	} else {

		transport, err := rest.TransportFor(k8sConfig)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create transport for k8s config")
		}

		oa, err := oauth.InitOauth2(
			cfg.OCClientID,
			cfg.OCClientSecret,
			k8sConfig.Host,
			cfg.Host+"/auth/callback",
			&http.Client{Transport: transport})
		if err != nil {
			return nil, errors.Wrap(err, "unable to init oauth2 client")
		}

		if !cfg.OAuthUseExternalTokenURL {
			if err := oa.UseInternalTokenService(context.Background(), cfg.OAuthInternalTokenHost, k8sService); err != nil {
				return nil, errors.Wrap(err, "unable to load internal oauth host")
			}
		}

		oAuth = oa
	}

	gob.Register(&oauth2.Token{})
	r.Use(oauth.MakeGinMiddleware(oAuth, router.AuthTokenSessionKey, router.AuthTokenValidSessionKey, "/admin/"))
	r.Use(router.UserDataMiddleware)

	return oAuth, nil
}

func i18n(word ...string) string {
	message := strings.TrimSpace(strings.Join(word, " "))
	return gotext.Get(message)
}
