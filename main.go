package main

import (
	"context"
	"ddm-admin-console/service/openshift"
	"ddm-admin-console/service/vault"
	"encoding/gob"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/leonelquinteros/gotext"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/oauth2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"

	"ddm-admin-console/app/cluster"
	"ddm-admin-console/app/dashboard"
	"ddm-admin-console/app/registry"
	"ddm-admin-console/auth"
	oauth "ddm-admin-console/auth"
	"ddm-admin-console/config"
	codebaseController "ddm-admin-console/controller/codebase"
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	edpComponent "ddm-admin-console/service/edp_component"
	"ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/jenkins"
	"ddm-admin-console/service/k8s"
	"ddm-admin-console/service/keycloak"
)

func main() {
	configPath := flag.String("c", "default.env", "config file path")
	flag.Parse()

	cnf, err := loadConfig(*configPath)
	if err != nil {
		panic(err)
	}

	logger, err := getLogger(cnf.LogLevel, cnf.LogEncoding)
	if err != nil {
		panic(err)
	}

	logger.Info("init gin router")
	gin.SetMode(cnf.GinMode)
	r := gin.New()
	r.SetFuncMap(template.FuncMap{"i18n": i18n})
	r.LoadHTMLGlob("templates/**/*")
	r.Static("/static", "./static")
	store := cookie.NewStore([]byte(cnf.SessionSecret))
	r.Use(sessions.Sessions("cookie-session", store))

	logger.Info("init apps")
	if err := initApps(logger, cnf, r); err != nil {
		panic(fmt.Sprintf("%+v", err))
	}

	logger.Info("init i18n")
	gotext.Configure("locale", "uk_UA", "default")

	logger.Info("run router on port", zap.String("port", cnf.HTTPPort))
	if err := r.Run(fmt.Sprintf(":%s", cnf.HTTPPort)); err != nil {
		panic(err)
	}
}

func loadConfig(path string) (*config.Settings, error) {
	if err := godotenv.Load(path); err != nil {
		return nil, errors.Wrap(err, "unable to load config file")
	}

	var cnf config.Settings
	if err := envconfig.Process("", &cnf); err != nil {
		return nil, errors.Wrap(err, "unable to parse env variables")
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
		return nil, errors.Wrap(err, "unable to build logger")
	}

	return logger, nil
}

func initServices(sch *runtime.Scheme, restConf *rest.Config, appConf *config.Settings) (*config.Services, error) {
	var err error
	serviceItems := config.Services{}

	serviceItems.EDPComponent, err = edpComponent.Make(sch, restConf, appConf.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init edp component service")
	}

	serviceItems.Codebase, err = codebase.Make(sch, restConf, appConf.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init codebase service")
	}

	serviceItems.K8S, err = k8s.Make(restConf, appConf.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init k8s service")
	}

	serviceItems.Jenkins, err = jenkins.Make(sch, restConf, serviceItems.K8S,
		jenkins.Config{Namespace: appConf.Namespace, APIUrl: appConf.JenkinsAPIURL,
			AdminSecretName: appConf.JenkinsAdminSecretName})
	if err != nil {
		return nil, errors.Wrap(err, "unable to init jenkins service")
	}

	serviceItems.OpenShift, err = openshift.Make(restConf)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init open shift service")
	}

	serviceItems.Gerrit, err = gerrit.Make(sch, restConf,
		gerrit.Config{Namespace: appConf.Namespace, GerritAPIUrlTemplate: appConf.GerritAPIUrlTemplate,
			RootGerritName: appConf.RootGerritName})
	if err != nil {
		return nil, errors.Wrap(err, "unable to create gerrit service")
	}

	serviceItems.Keycloak, err = keycloak.Make(sch, restConf, appConf.UsersNamespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create keycloak service")
	}

	serviceItems.Vault, err = vault.Make(appConf.VaultConfig(), serviceItems.K8S)
	if err != nil {
		return nil, errors.Wrap(err, "unable to init vault service")
	}

	return &serviceItems, nil
}

func initControllers(sch *runtime.Scheme, namespace string, logger *zap.Logger, cnf *config.Settings,
	services *config.Services) error {
	cfg := ctrl.GetConfigOrDie()

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             sch,
		Namespace:          namespace,
		MetricsBindAddress: "0",
	})

	if err != nil {
		return errors.Wrap(err, "unable to ini manager")
	}

	if err := codebaseController.Make(mgr, logger.Sugar(),
		registry.MakeAdmins(services.Keycloak, cnf.UsersRealm, cnf.UsersNamespace), cnf); err != nil {
		return errors.Wrap(err, "unable to init codebase controller")
	}

	go func() {
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			logger.Sugar().Error(err.Error(), "unable to start manager")
		}
	}()

	return nil
}

func initApps(logger *zap.Logger, cnf *config.Settings, r *gin.Engine) error {
	restConf, err := initKubeConfig()
	if err != nil {
		return errors.Wrap(err, "unable to init kube config")
	}

	appRouter := router.Make(r, logger)

	sch := runtime.NewScheme()
	if err := v1.AddToScheme(sch); err != nil {
		return errors.Wrap(err, "unable to add core api to scheme")
	}

	serviceItems, err := initServices(sch, restConf, cnf)
	if err != nil {
		return errors.Wrap(err, "unable to init services")
	}

	if err := initControllers(sch, cnf.Namespace, logger, cnf, serviceItems); err != nil {
		return errors.Wrap(err, "unable to init controllers")
	}

	oa, err := initOauth(restConf, cnf, r, serviceItems.K8S)
	if err != nil {
		return errors.Wrap(err, "unable to init oauth")
	}

	_, err = dashboard.Make(appRouter, oa, serviceItems, cnf.ClusterCodebaseName)
	if err != nil {
		return errors.Wrap(err, "unable to make dashboard app")
	}

	_, err = registry.Make(appRouter, serviceItems.RegistryServices(), cnf.RegistryConfig())
	if err != nil {
		return errors.Wrap(err, "unable to make registry app")
	}

	cluster.Make(appRouter, serviceItems.ClusterServices(), cnf.ClusterConfig())

	return nil
}

func initKubeConfig() (*rest.Config, error) {
	k8sConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)

	restConfig, err := k8sConfig.ClientConfig()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get k8s client config")
	}

	return restConfig, nil
}

func initOauth(k8sConfig *rest.Config, cfg *config.Settings, r *gin.Engine, k8sService k8s.ServiceInterface) (*auth.OAuth2, error) {
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

	gob.Register(&oauth2.Token{})
	r.Use(oauth.MakeGinMiddleware(oa, router.AuthTokenSessionKey, router.AuthTokenValidSessionKey, "/admin/"))
	r.Use(router.UserDataMiddleware)

	return oa, nil
}

func i18n(word ...string) string {
	message := strings.TrimSpace(strings.Join(word, " "))
	return gotext.Get(message)
}
