package mocks

import (
	"ddm-admin-console/app/cluster"
	"ddm-admin-console/app/registry"
	"ddm-admin-console/config"
	mockCodebase "ddm-admin-console/mocks/codebase"
	mockDashboard "ddm-admin-console/mocks/dashboard"
	mockEdpComponent "ddm-admin-console/mocks/edp_component"
	mockGerrit "ddm-admin-console/mocks/gerrit"
	mockJenkins "ddm-admin-console/mocks/jenkins"
	mockK8S "ddm-admin-console/mocks/k8s"
	mockKeycloak "ddm-admin-console/mocks/keycloak"
	mockOpenshift "ddm-admin-console/mocks/openshift"
	mockPermissions "ddm-admin-console/mocks/permissions"
	mockVault "ddm-admin-console/mocks/vault"
	"ddm-admin-console/service/codebase"
	edpcomponent "ddm-admin-console/service/edp_component"
	"ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/openshift"
	"net/url"
	"time"

	"github.com/hashicorp/vault/api"

	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func InitServices(cnf *config.Settings) *config.Services {
	edpComponent := mockEdpComponent.ServiceInterface{}
	edpComponent.On("GetAll", mock.Anything, mock.Anything).Return([]edpcomponent.EDPComponent{}, nil)
	edpComponent.On("Get", mock.Anything, mock.Anything).Return(&edpcomponent.EDPComponent{
		Spec: edpcomponent.EDPComponentSpec{Url: "https://example.com/foo/bar"},
	}, nil)
	edpComponent.On("GetAllNamespace", mock.Anything, mock.Anything, true).Return([]edpcomponent.EDPComponent{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "mock"},
		},
	}, nil)

	openShift := mockOpenshift.ServiceInterface{}
	openShift.On("GetMe", mock.Anything).Return(&openshift.User{
		Metadata: openshift.Metadata{
			Name: "mock@example.com",
		},
		FullName: "mock",
	}, nil)

	pms := mockPermissions.ServiceInterface{}
	pms.On("DeleteTokenContext", mock.Anything).Return(nil)
	pms.On("LoadUserRegistries", mock.Anything).Return(nil)
	pms.On("FilterCodebases", mock.Anything, mock.Anything, mock.Anything).Return([]codebase.WithPermissions{
		{
			Codebase:  &codebase.Codebase{ObjectMeta: metav1.ObjectMeta{Name: "mock"}},
			CanUpdate: true,
			CanDelete: true,
		},
	}, nil)

	svc := config.Services{
		Codebase:     initCodebaseService(cnf),
		Vault:        initVault(),
		K8S:          initK8SService(cnf),
		Jenkins:      &mockJenkins.ServiceInterface{},
		Gerrit:       initMockGerrit(cnf),
		EDPComponent: &edpComponent,
		Keycloak:     &mockKeycloak.ServiceInterface{},
		OpenShift:    &openShift,
		PermService:  &pms,
	}

	return &svc
}

func initVault() *mockVault.ServiceInterface {
	v := mockVault.ServiceInterface{}

	v.On("Write", mock.Anything, mock.Anything).Return(&api.Secret{}, nil)

	return &v
}

func initCodebaseService(cnf *config.Settings) *mockCodebase.ServiceInterface {
	cbService := mockCodebase.ServiceInterface{}
	cbService.On("GetAllByType", mock.Anything).Return([]codebase.Codebase{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "mock"},
		},
	}, nil)
	cbService.On("CheckPermissions", mock.Anything, mock.Anything).Return([]codebase.WithPermissions{}, nil)
	cbService.On("CheckIsAllowedToCreate", mock.Anything).Return(true, nil)
	cbService.On("ServiceForContext", mock.Anything).Return(&cbService, nil)

	clusterDescription := "cluster description"
	cbService.On("Get", cnf.ClusterCodebaseName).Return(&codebase.Codebase{
		ObjectMeta: metav1.ObjectMeta{Name: cnf.ClusterCodebaseName},
		Spec:       codebase.CodebaseSpec{Description: &clusterDescription},
	}, nil)
	cbService.On("GetBranchesByCodebase", cnf.ClusterCodebaseName).Return([]codebase.CodebaseBranch{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "master"},
		},
	}, nil)

	return &cbService
}

func initK8SService(cnf *config.Settings) *mockK8S.ServiceInterface {
	k8sService := mockK8S.ServiceInterface{}
	k8sService.On("ServiceForContext", mock.Anything).Return(&k8sService, nil)
	k8sService.On("CanI", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

	return &k8sService
}

func initMockGerrit(cnf *config.Settings) *mockGerrit.ServiceInterface {
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

	mockRegistryValues := registry.Values{
		Global: registry.Global{},
		Administrators: []registry.Admin{
			{Email: "foo@bar.com"},
		},
		Keycloak: registry.Keycloak{
			CustomHost: "foo.bar.com",
		},
	}
	bts, err := yaml.Marshal(mockRegistryValues)
	if err != nil {
		panic(err)
	}

	grService.On("GetBranchContent", cnf.ClusterCodebaseName, "master",
		url.PathEscape(registry.ValuesLocation)).Return(string(bts), nil)
	grService.On("GetBranchContent", "mock", "master", url.PathEscape(registry.ValuesLocation)).Return(string(bts), nil)
	grService.On("GetMergeRequestByProject", mock.Anything, cnf.ClusterCodebaseName).Return([]gerrit.GerritMergeRequest{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "mock-mr"},
		},
	}, nil)
	grService.On("GetProject", mock.Anything, cnf.ClusterCodebaseName).Return(&gerrit.GerritProject{
		ObjectMeta: metav1.ObjectMeta{Name: "mock-project"},
	}, nil)

	mockClusterValues := cluster.Values{
		Keycloak: cluster.Keycloak{CustomHosts: []cluster.CustomHost{
			{
				Host:            "foo.bar.com",
				CertificatePath: "/foo/bar/com",
			},
			{
				Host:            "foo2.bar.com",
				CertificatePath: "/foo2/bar/com",
			},
		}},
	}
	bts, err = yaml.Marshal(mockClusterValues)
	if err != nil {
		panic(err)
	}
	grService.On("GetFileContents", mock.Anything, cnf.ClusterCodebaseName, registry.MasterBranch,
		registry.ValuesLocation).Return(string(bts), nil)

	return &grService
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
