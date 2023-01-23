package dashboard

import (
	"context"
	"ddm-admin-console/service/openshift"
	"net/http"

	"ddm-admin-console/config"
	"ddm-admin-console/router"
	"ddm-admin-console/service/codebase"
	edpComponent "ddm-admin-console/service/edp_component"
	"ddm-admin-console/service/k8s"

	"golang.org/x/oauth2"

	"go.uber.org/zap"
)

type Logger interface {
	Error(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
}

type OAuth interface {
	GetTokenClient(ctx context.Context, code string) (token *oauth2.Token, oauthClient *http.Client, err error)
}

type App struct {
	router              router.Interface
	logger              Logger
	edpComponentService edpComponent.ServiceInterface
	oauth               OAuth
	k8sService          k8s.ServiceInterface
	codebaseService     codebase.ServiceInterface
	openShiftService    openshift.ServiceInterface

	clusterCodebaseName string
}

func Make(router router.Interface, oauth OAuth, services *config.Services, clusterCodebaseName string) (*App, error) {
	app := App{
		router:              router,
		edpComponentService: services.EDPComponent,
		oauth:               oauth,
		k8sService:          services.K8S,
		openShiftService:    services.OpenShift,
		clusterCodebaseName: clusterCodebaseName,
		codebaseService:     services.Codebase,
	}

	app.createRoutes()

	return &app, nil
}
