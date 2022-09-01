package codebase

import (
	"context"
	"fmt"
	"time"

	"ddm-admin-console/service"
	"ddm-admin-console/service/k8s"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	pkgScheme "sigs.k8s.io/controller-runtime/pkg/scheme"
)

const (
	defaultBranch         = "master"
	gerritCreatorUsername = "user"
	gerritCreatorPassword = "password"
	RegistryCodebaseType  = "registry"
)

type Service struct {
	service.UserConfig
	k8sClient  client.Client
	scheme     *runtime.Scheme
	namespace  string
	restConfig *rest.Config
}

type ErrAlreadyExists string

func (e ErrAlreadyExists) Error() string {
	return string(e)
}

func IsErrAlreadyExists(err error) bool {
	switch errors.Cause(err).(type) {
	case ErrAlreadyExists:
		return true
	}

	return false
}

func Make(sch *runtime.Scheme, k8sConfig *rest.Config, namespace string) (*Service, error) {
	builder := pkgScheme.Builder{GroupVersion: schema.GroupVersion{Group: "v2.edp.epam.com", Version: "v1alpha1"}}
	builder.Register(&Codebase{}, &CodebaseBranch{}, &CodebaseBranchList{}, &CodebaseList{}, &GitServer{},
		&GitServerList{})

	if err := builder.AddToScheme(sch); err != nil {
		return nil, errors.Wrap(err, "error during builder add to scheme")
	}

	cl, err := client.New(k8sConfig, client.Options{
		Scheme: sch,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to init k8s jenkins client")
	}

	return &Service{
		k8sClient: cl,
		scheme:    sch,
		namespace: namespace,
		UserConfig: service.UserConfig{
			RestConfig: k8sConfig,
		},
		restConfig: k8sConfig,
	}, nil
}

func (s *Service) GetAll() ([]Codebase, error) {
	var lst CodebaseList
	if err := s.k8sClient.List(context.Background(), &lst, &client.ListOptions{Namespace: s.namespace}); err != nil {
		return nil, errors.Wrap(err, "unable to get codebases")
	}

	return lst.Items, nil
}

func (s *Service) GetAllByType(tp string) ([]Codebase, error) {
	all, err := s.GetAll()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get all codebases")
	}

	result := make([]Codebase, 0, len(all))
	for _, v := range all {
		if v.Spec.Type == tp {
			result = append(result, v)
		}
	}

	return result, nil
}

func (s *Service) Get(name string) (*Codebase, error) {
	var cb Codebase
	if err := s.k8sClient.Get(context.Background(), types.NamespacedName{Namespace: s.namespace, Name: name}, &cb); err != nil {
		return nil, errors.Wrapf(err, "unable to get codebase: %s", name)
	}

	return &cb, nil
}

func (s *Service) Create(cb *Codebase) error {
	cb.Namespace = s.namespace

	if err := s.k8sClient.Create(context.Background(), cb); err != nil {
		return errors.Wrapf(err, "unable to create codebase: %+v", cb)
	}

	return nil
}

func (s *Service) CreateBranch(branch *CodebaseBranch) error {
	branch.Namespace = s.namespace

	if err := s.k8sClient.Create(context.Background(), branch); err != nil {
		return errors.Wrap(err, "unable to create codebase branch")
	}

	return nil
}

func (s *Service) Update(cb *Codebase) error {
	if err := s.k8sClient.Update(context.Background(), cb); err != nil {
		return errors.Wrapf(err, "unable to update codebase: %+v", cb)
	}

	return nil
}

func (s *Service) Delete(name string) error {
	cb, err := s.Get(name)
	if err != nil {
		return errors.Wrapf(err, "unable to get codebase: %s", name)
	}

	fg := metav1.DeletePropagationForeground

	if err := s.k8sClient.Delete(context.Background(), cb, &client.DeleteOptions{
		PropagationPolicy: &fg,
	}); err != nil {
		return errors.Wrapf(err, "unable to delete codebase: %+v", cb)
	}

	return nil
}

func (s *Service) GetAllBranches() ([]CodebaseBranch, error) {
	var lst CodebaseBranchList
	if err := s.k8sClient.List(context.Background(), &lst, &client.ListOptions{Namespace: s.namespace}); err != nil {
		return nil, errors.Wrap(err, "unable to get all codebase branches")
	}

	return lst.Items, nil
}

func (s *Service) GetBranchesByCodebase(codebaseName string) ([]CodebaseBranch, error) {
	branches, err := s.GetAllBranches()
	if err != nil {
		return nil, errors.Wrap(err, "unable to get all branches")
	}

	filteredBranches := make([]CodebaseBranch, 0, len(branches))
	for _, br := range branches {
		if br.Spec.CodebaseName == codebaseName {
			filteredBranches = append(filteredBranches, br)
		}
	}

	return filteredBranches, nil
}

func (s *Service) CreateDefaultBranch(cb *Codebase) error {
	blockOwnerDel := true
	buildNo := "0"

	branch := CodebaseBranch{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v2.edp.epam.com/v1alpha1",
			Kind:       "CodebaseBranch",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", cb.Name, defaultBranch),
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "v2.edp.epam.com/v1alpha1",
					Kind:               "Codebase",
					Name:               cb.Name,
					UID:                cb.UID,
					BlockOwnerDeletion: &blockOwnerDel,
				},
			},
		},
		Spec: CodebaseBranchSpec{
			BranchName:   defaultBranch,
			Version:      cb.Spec.Versioning.StartFrom,
			Release:      false,
			CodebaseName: cb.Name,
		},
		Status: CodebaseBranchStatus{
			Build:               &buildNo,
			Status:              "initialized",
			LastTimeUpdated:     time.Now(),
			LastSuccessfulBuild: nil,
			Username:            "",
			Action:              "codebase_branch_registration",
			Result:              "success",
			Value:               "inactive",
		},
	}

	if err := s.CreateBranch(&branch); err != nil {
		return errors.Wrap(err, "unable to create branch")
	}

	return nil
}

func (s *Service) ServiceForContext(ctx context.Context) (ServiceInterface, error) {
	userConfig, changed := s.UserConfig.CreateConfig(ctx)
	if !changed {
		return s, nil
	}

	svc, err := Make(s.scheme, userConfig, s.namespace)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create service for context")
	}

	return svc, nil
}

func (s *Service) CreateTempSecrets(cb *Codebase, k8sService k8s.ServiceInterface, gerritCreatorSecretName string) error {
	secret, err := k8sService.GetSecret(gerritCreatorSecretName)
	if err != nil {
		return errors.Wrap(err, "unable to get secret")
	}

	username, ok := secret.Data[gerritCreatorUsername]
	if !ok {
		return errors.Wrap(err, "gerrit creator secret does not have username")
	}

	pwd, ok := secret.Data[gerritCreatorPassword]
	if !ok {
		return errors.Wrap(err, "gerrit creator secret does not have password")
	}

	repoSecretName := fmt.Sprintf("repository-codebase-%s-temp", cb.Name)
	if err := k8sService.RecreateSecret(repoSecretName, map[string][]byte{
		"username":            username,
		gerritCreatorPassword: pwd,
	}); err != nil {
		return errors.Wrapf(err, "unable to create secret: %s", repoSecretName)
	}

	return nil
}
