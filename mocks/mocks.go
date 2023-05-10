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
	"errors"
	"net/url"
	"time"

	"github.com/hashicorp/vault/api"

	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func InitServices(cnf *config.Settings) *config.Services {
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
		{
			Codebase: &codebase.Codebase{
				ObjectMeta: metav1.ObjectMeta{Name: "mock-branch-inactive", Annotations: map[string]string{
					codebase.StatusAnnotation: codebase.StatusAnnotationInactiveBranches,
				}},
			},
			CanUpdate: true,
			CanDelete: true,
		},
	}, nil)

	svc := config.Services{
		Codebase:     initCodebaseService(cnf),
		Vault:        initVault(),
		K8S:          initK8SService(cnf),
		Jenkins:      initJenkinsService(),
		Gerrit:       initMockGerrit(cnf),
		EDPComponent: initEDPComponent(),
		Keycloak:     &mockKeycloak.ServiceInterface{},
		OpenShift:    &openShift,
		PermService:  &pms,
	}

	return &svc
}

func initEDPComponent() *mockEdpComponent.ServiceInterface {
	edpComponent := mockEdpComponent.ServiceInterface{}
	edpComponent.On("GetAll", mock.Anything, mock.Anything).Return([]edpcomponent.EDPComponent{}, nil)
	edpComponent.On("Get", mock.Anything, mock.Anything).Return(&edpcomponent.EDPComponent{
		Spec: edpcomponent.EDPComponentSpec{Url: "https://example.com/foo/bar"},
	}, nil)
	edpComponent.On("GetAllNamespace", mock.Anything, mock.Anything, true).Return([]edpcomponent.EDPComponent{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "mock"},
			Spec:       edpcomponent.EDPComponentSpec{Url: "http://google.com"},
		},
	}, nil)
	edpComponent.On("GetAllCategory", mock.Anything, mock.Anything).Return(map[string][]edpcomponent.EDPComponentItem{
		"foo": {
			{},
		},
	}, nil)

	return &edpComponent
}

func initJenkinsService() *mockJenkins.ServiceInterface {
	svc := mockJenkins.ServiceInterface{}
	svc.On("GetJobStatus", mock.Anything, "mock/view/MOCK-BRANCH/job/MOCK-BRANCH-Build-mock").
		Return("SUCCESS", int64(11), nil)

	svc.On("GetJobStatus", mock.Anything, "mock/view/MASTER/job/MASTER-Build-mock").
		Return("FAILURE", int64(11), nil)

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
		{
			ObjectMeta: metav1.ObjectMeta{Name: "mock-branch-inactive", Annotations: map[string]string{
				codebase.StatusAnnotation: codebase.StatusAnnotationInactiveBranches,
			}},
		},
	}, nil)
	cbService.On("CheckPermissions", mock.Anything, mock.Anything).Return([]codebase.WithPermissions{}, nil)
	cbService.On("CheckIsAllowedToCreate", mock.Anything).Return(true, nil)
	cbService.On("ServiceForContext", mock.Anything).Return(&cbService, nil)

	clusterDescription := "cluster description"
	cbService.On("Get", cnf.ClusterCodebaseName).Return(&codebase.Codebase{
		ObjectMeta: metav1.ObjectMeta{Name: cnf.ClusterCodebaseName},
		Spec:       codebase.CodebaseSpec{Description: &clusterDescription, Repository: &codebase.Repository{}},
	}, nil)
	cbService.On("GetBranchesByCodebase", mock.Anything, cnf.ClusterCodebaseName).Return([]codebase.CodebaseBranch{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "master"},
			Spec: codebase.CodebaseBranchSpec{
				BranchName:   "mock-branch",
				CodebaseName: "mock",
			},
		},
	}, nil)
	cbService.On("CheckIsAllowedToUpdate", mock.Anything, mock.Anything).Return(true, nil)

	mockDescription := "mock description"
	cbService.On("Get", "mock").Return(&codebase.Codebase{
		ObjectMeta: metav1.ObjectMeta{Name: "mock"},
		Spec:       codebase.CodebaseSpec{Description: &mockDescription, Repository: &codebase.Repository{}},
	}, nil)
	cbService.On("GetBranchesByCodebase", mock.Anything, "mock").Return([]codebase.CodebaseBranch{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "master"},
			Spec: codebase.CodebaseBranchSpec{
				BranchName:   "mock-branch",
				CodebaseName: "mock",
			},
		},
	}, nil)

	cbService.On("Get", "mock2").Return(nil, errors.New("not found"))

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
		{
			ObjectMeta: metav1.ObjectMeta{Name: "registry-tenant-template-tpl1"},
			Status: gerrit.GerritProjectStatus{
				Branches: []string{"refs/heads/1.0"},
			},
			Spec: gerrit.GerritProjectSpec{Name: "registry-tenant-template-tpl1"},
		},
	}, nil)

	mockClusterValuesFromRegistry := registry.ClusterValues{Keycloak: registry.ClusterKeycloak{CustomHosts: []registry.CustomHost{
		{Host: "foo.bar.com"},
		{Host: "zozo.com.ua"},
	}}}
	clusterValuesBts, err := yaml.Marshal(mockClusterValuesFromRegistry)
	if err != nil {
		panic(err)
	}

	grService.On("GetBranchContent", cnf.ClusterCodebaseName, "master",
		url.PathEscape(registry.ValuesLocation)).Return(string(clusterValuesBts), nil)

	mockRegistryValues := registry.Values{
		Global: registry.Global{
			DeploymentMode: registry.DeploymentModeDevelopment,
			WhiteListIP: registry.WhiteListIP{
				CitizenPortal: "192.168.1.1 18.1.1.0/32",
				AdminRoutes:   "10.0.0.1 8.8.8.8",
				OfficerPortal: "10.5.1.2/32 2.5.2.1",
			},
			CrunchyPostgres: registry.CrunchyPostgres{
				CrunchyPostgresPostgresql: registry.CrunchyPostgresPostgresql{
					CrunchyPostgresPostgresqlParameters: registry.CrunchyPostgresPostgresqlParameters{
						MaxConnections: 150},
				},
				StorageSize: "10Gi",
			},
		},
		Administrators: []registry.Admin{
			{Email: "foo@bar.com"},
		},
		Keycloak: registry.Keycloak{
			CustomHost: "foo.bar.com",
		},
		Trembita: registry.Trembita{
			IPList: []string{"8.8.8.8", "9.9.9.9", "10.0.0.1", "10.0.0.2", "10.0.0.3", "10.0.0.4", "10.0.0.5",
				"10.0.0.6", "10.0.0.7"},
			Registries: map[string]registry.TrembitaRegistry{
				"test": {
					Mock: true,
					URL:  "http://wiremock/",
					Type: "registry",
				},
			},
		},
	}
	registryValuesBts, err := yaml.Marshal(mockRegistryValues)
	if err != nil {
		panic(err)
	}

	grService.On("GetBranchContent", "mock", "master", url.PathEscape(registry.ValuesLocation)).Return(string(registryValuesBts), nil)
	grService.On("GetBranchContent", "registry-tenant-template-tpl1", "1.0",
		url.PathEscape(registry.ValuesLocation)).Return(string(registryValuesBts), nil)
	grService.On("GetMergeRequestByProject", mock.Anything, cnf.ClusterCodebaseName).Return([]gerrit.GerritMergeRequest{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "mock-mr", Labels: map[string]string{}},
		},
	}, nil)
	grService.On("GetMergeRequestByProject", mock.Anything, "mock").Return([]gerrit.GerritMergeRequest{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "mock-mr", Labels: map[string]string{}},
		},
	}, nil)
	grService.On("GetProject", mock.Anything, cnf.ClusterCodebaseName).Return(&gerrit.GerritProject{
		ObjectMeta: metav1.ObjectMeta{Name: "mock-project"},
	}, nil)

	grService.On("GetProject", mock.Anything, "mock").Return(&gerrit.GerritProject{
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
	bts, err := yaml.Marshal(mockClusterValues)
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
