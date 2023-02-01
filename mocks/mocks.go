package mocks

import (
	"ddm-admin-console/config"
	mockCodebase "ddm-admin-console/mocks/codebase"
	mockDashboard "ddm-admin-console/mocks/dashboard"
	mockEdpComponent "ddm-admin-console/mocks/edp_component"
	mockGerrit "ddm-admin-console/mocks/gerrit"
	mockJenkins "ddm-admin-console/mocks/jenkins"
	mockK8S "ddm-admin-console/mocks/k8s"
	mockKeycloak "ddm-admin-console/mocks/keycloak"
	mockOpenshift "ddm-admin-console/mocks/openshift"
	mockVault "ddm-admin-console/mocks/vault"
	"ddm-admin-console/service/codebase"
	edpcomponent "ddm-admin-console/service/edp_component"
	"ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/openshift"
	"time"

	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func InitServices(cnf *config.Settings) *config.Services {
	edpComponent := mockEdpComponent.ServiceInterface{}
	edpComponent.On("GetAll", mock.Anything).Return([]edpcomponent.EDPComponent{}, nil)
	edpComponent.On("Get", mock.Anything, cnf.DDMManualEDPComponent).Return(&edpcomponent.EDPComponent{
		Spec: edpcomponent.EDPComponentSpec{Url: "http://example.com/foo/bar"},
	}, nil)

	openShift := mockOpenshift.ServiceInterface{}
	openShift.On("GetMe", mock.Anything).Return(&openshift.User{
		Metadata: openshift.Metadata{
			Name: "mock@example.com",
		},
		FullName: "mock",
	}, nil)

	k8sService := mockK8S.ServiceInterface{}
	k8sService.On("ServiceForContext", mock.Anything).Return(&k8sService, nil)
	k8sService.On("CanI", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

	cbService := mockCodebase.ServiceInterface{}
	cbService.On("GetAllByType", mock.Anything).Return([]codebase.Codebase{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "mock"},
		},
	}, nil)
	cbService.On("CheckPermissions", mock.Anything, mock.Anything).Return([]codebase.WithPermissions{}, nil)
	cbService.On("CheckIsAllowedToCreate", mock.Anything).Return(true, nil)

	grService := mockGerrit.ServiceInterface{}
	grService.On("GetMergeRequests", mock.Anything).Return([]gerrit.GerritMergeRequest{}, nil)
	grService.On("GetProjects", mock.Anything).Return([]gerrit.GerritProject{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "mock-project"},
			Status: gerrit.GerritProjectStatus{
				Branches: []string{"mock-branch"},
			},
		},
	}, nil)

	svc := config.Services{
		Codebase:     &cbService,
		Vault:        &mockVault.ServiceInterface{},
		K8S:          &k8sService,
		Jenkins:      &mockJenkins.ServiceInterface{},
		Gerrit:       &grService,
		EDPComponent: &edpComponent,
		Keycloak:     &mockKeycloak.ServiceInterface{},
		OpenShift:    &openShift,
	}

	return &svc
}

func OAuth() *mockDashboard.OAuth {
	oa := mockDashboard.OAuth{}
	oa.On("AuthCodeURL").Return("/auth/callback?code=mock")
	oa.On("GetTokenClient", mock.Anything, "mock").Return(&oauth2.Token{
		AccessToken: "mock-token",
		Expiry:      time.Now().Add(time.Hour * 24 * 365),
	}, nil, nil)
	return &oa
}
