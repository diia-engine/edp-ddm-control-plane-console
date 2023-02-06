package codebase

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"reflect"
	"time"

	"github.com/patrickmn/go-cache"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"ddm-admin-console/app/registry"
	"ddm-admin-console/config"
	"ddm-admin-console/controller"
	"ddm-admin-console/service"
	codebaseService "ddm-admin-console/service/codebase"
	gerritService "ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/git"
)

const (
	registryRemoteName = "registry"
	keySecretIndex     = "id_rsa"
)

type Controller struct {
	logger    controller.Logger
	mgr       ctrl.Manager
	k8sClient client.Client
	cnf       *config.Settings
	appCache  *cache.Cache
}

type AdminSyncer interface {
	SyncAdmins(ctx context.Context, registryName string, admins []registry.Admin) error
}

func Make(mgr ctrl.Manager, logger controller.Logger, cnf *config.Settings, _c *cache.Cache) error {
	c := Controller{
		mgr:       mgr,
		logger:    logger,
		k8sClient: mgr.GetClient(),
		cnf:       cnf,
		appCache:  _c,
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&codebaseService.Codebase{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: isSpecUpdated})).
		Complete(&c); err != nil {
		return errors.Wrap(err, "unable to create controller")
	}

	return nil
}

func GerritSSHURL(cnf *config.Settings) string {
	return fmt.Sprintf("ssh://%s@%s:%s", cnf.GitUsername, cnf.GitHost, cnf.GitPort)
}

func isSpecUpdated(e event.UpdateEvent) bool {
	oo := e.ObjectOld.(*codebaseService.Codebase)
	no := e.ObjectNew.(*codebaseService.Codebase)

	return !reflect.DeepEqual(oo.Spec, no.Spec) ||
		(oo.GetDeletionTimestamp().IsZero() && !no.GetDeletionTimestamp().IsZero())
}

func (c *Controller) Reconcile(ctx context.Context, request reconcile.Request) (result reconcile.Result,
	resultErr error) {
	c.logger.Infow("reconciling codebase", "Request.Namespace", request.Namespace,
		"Request.Name", request.Name)

	var instance codebaseService.Codebase
	if err := c.k8sClient.Get(ctx, request.NamespacedName, &instance); err != nil {
		if k8sErrors.IsNotFound(err) {
			c.logger.Infow("instance not found", "Request.Namespace", request.Namespace, "Request.Name", request.Name)
			return
		}

		resultErr = errors.Wrap(err, "unable to get codebase from k8s")
		return
	}

	if instance.Spec.Type == codebaseService.RegistryCodebaseType {
		if err := c.reconcile(ctx, &instance); err != nil {
			c.logger.Error(err)
			result = reconcile.Result{RequeueAfter: time.Second * 10}
			return
		}
	}

	c.logger.Infow("reconciling done", "Request.Namespace", request.Namespace,
		"Request.Name", request.Name)

	return
}

func (c *Controller) reconcile(ctx context.Context, instance *codebaseService.Codebase) error {
	if err := c.updateImportRepo(ctx, instance); err != nil {
		return errors.Wrap(err, "unable to update import repo")
	}

	return nil
}

func (c *Controller) updateImportRepo(ctx context.Context, instance *codebaseService.Codebase) error {
	if instance.Spec.GitUrlPath == nil || *instance.Spec.GitUrlPath != codebaseService.RepoNotReady {
		return nil
	}

	prj, err := c.getGerritProject(ctx, instance.Name)
	if service.IsErrNotFound(err) {
		return ErrPostpone(time.Second * 5)
	} else if err != nil {
		return errors.Wrap(err, "unknown error")
	}

	if prj.Status.Value != "OK" {
		return ErrPostpone(time.Second * 5)
	}

	if err := c.pushRegistryTemplate(ctx, instance); err != nil {
		return errors.Wrap(err, "unable to push registry template")
	}

	gitUrlPath := fmt.Sprintf("/%s", instance.Name)
	instance.Spec.GitUrlPath = &gitUrlPath

	if err := c.k8sClient.Update(ctx, instance); err != nil {
		return errors.Wrap(err, "unable to update codebase")
	}

	return nil
}

func (c *Controller) pushRegistryTemplate(ctx context.Context, instance *codebaseService.Codebase) error {
	reposPath, err := PrepareControllerTempFolder(c.cnf.TempFolder, "repos")
	if err != nil {
		return errors.Wrap(err, "unable to create repos folder")
	}

	privateKey, err := GetGerritPrivateKey(ctx, c.k8sClient, c.cnf)
	if err != nil {
		return errors.Wrap(err, "unable to get gerrit private key")
	}

	gitService := git.Make(path.Join(reposPath, instance.Name), c.cnf.GitUsername, privateKey)
	defer func() {
		if err := gitService.Clean(); err != nil {
			c.logger.Error(err)
		}
	}()

	if err := c.initCodebaseRepo(instance, gitService); err != nil {
		return errors.Wrap(err, "unable to init codebase repo")
	}

	_, err = gitService.Pull(registryRemoteName)
	if git.IsErrReferenceNotFound(err) {
		if err := c.replaceDefaultBranch(instance, gitService); err != nil {
			return errors.Wrap(err, "unable to replace default branch")
		}
	} else if err != nil {
		return errors.Wrap(err, "unable to pull")
	}

	cachedToCommit, err := c.setCachedFiles(instance, gitService)
	if err != nil {
		return fmt.Errorf("unable to set cached files, %w", err)
	}

	valuesToCommit, err := updateRegistryValues(instance, gitService)
	if err != nil {
		return errors.Wrap(err, "unable to update registry values")
	}

	if cachedToCommit || valuesToCommit {
		if err := gitService.SetAuthor(&git.User{Name: instance.Annotations[registry.AnnotationCreatorUsername],
			Email: instance.Annotations[registry.AnnotationCreatorEmail]}); err != nil {
			return errors.Wrap(err, "unable to set author")
		}

		if err := gitService.RawCommit("set initial values.yaml from admin console"); err != nil {
			return fmt.Errorf("unable to commit values, %w", err)
		}
	}

	if err := gitService.Push(registryRemoteName, "--all"); err != nil {
		return errors.Wrap(err, "unable to push [all] changes to repo registry")
	}

	if err := gitService.Push(registryRemoteName, "--tags"); err != nil {
		return errors.Wrap(err, "unable to push tags to repo registry")
	}

	return nil
}

func (c *Controller) replaceDefaultBranch(instance *codebaseService.Codebase, gitService *git.Service) error {
	if instance.Spec.BranchToCopyInDefaultBranch == "" {
		return gitService.Checkout(instance.Spec.DefaultBranch, false)
	}

	if err := gitService.Checkout(instance.Spec.BranchToCopyInDefaultBranch, false); err != nil {
		return errors.Wrap(err, "unable to checkout")
	}

	if err := gitService.RemoveBranch(instance.Spec.DefaultBranch); err != nil {
		return errors.Wrap(err, "unable to remove default branch")
	}

	if err := gitService.Checkout(instance.Spec.DefaultBranch, true); err != nil {
		return errors.Wrap(err, "unable to copy to default branch")
	}

	return nil
}

func (c *Controller) initCodebaseRepo(instance *codebaseService.Codebase, gitService *git.Service) error {
	gerritSSHURL := GerritSSHURL(c.cnf)

	tpl, ok := instance.Annotations[registry.AnnotationTemplateName]
	if !ok {
		return errors.New("template annotation not found")
	}
	if err := gitService.Clone(fmt.Sprintf("%s/%s", gerritSSHURL, tpl)); err != nil {
		return errors.Wrap(err, "unable to clone repo")
	}

	if err := gitService.AddRemote(registryRemoteName, fmt.Sprintf("%s/%s", gerritSSHURL, instance.Name)); err != nil {
		return errors.Wrap(err, "unable to add registry remote")
	}

	return nil
}

func PrepareControllerTempFolder(tempDir, controllerFolder string) (string, error) {
	controllerFolderPath := path.Join(tempDir, fmt.Sprintf("%s-%d", controllerFolder, time.Now().UnixNano()))
	if _, err := os.Stat(controllerFolderPath); err == nil {
		if err := os.RemoveAll(controllerFolderPath); err != nil {
			return "", errors.Wrap(err, "unable to clear repos folder")
		}
	}

	if err := os.MkdirAll(controllerFolderPath, 0777); err != nil {
		return "", errors.Wrap(err, "unable to create repo folder")
	}

	return controllerFolderPath, nil
}

func GetGerritPrivateKey(ctx context.Context, k8sClient client.Client, cnf *config.Settings) (string, error) {
	var gerritSecret v1.Secret
	if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: cnf.Namespace, Name: cnf.GitKeySecretName},
		&gerritSecret); err != nil {
		return "", errors.Wrap(err, "unable to get gerrit project creator secret")
	}

	key, ok := gerritSecret.Data[keySecretIndex]
	if !ok {
		return "", errors.Errorf("no data by key %s in gerrit secret", keySecretIndex)
	}

	return string(key), nil
}

func (c *Controller) setCachedFiles(instance *codebaseService.Codebase, gitService *git.Service) (bool, error) {
	files, ok := c.appCache.Get(registry.CachedFilesIndex(instance.Name))
	if !ok {
		return false, nil
	}

	cachedFiles, ok := files.([]registry.CachedFile)
	if !ok {
		return false, errors.New("wrong files type")
	}

	updatedFiles := 0

	for _, f := range cachedFiles {
		bts, err := os.ReadFile(f.TempPath)
		if err != nil {
			return false, fmt.Errorf("unable to read file, %w", err)
		}

		repoContents, err := gitService.GetFileContents(f.RepoPath)
		if err == nil && repoContents == string(bts) {
			continue
		}

		if err := gitService.SetFileContents(f.RepoPath, string(bts)); err != nil {
			return false, fmt.Errorf("unable to set file contents, %w", err)
		}

		if err := gitService.Add(f.RepoPath); err != nil {
			return false, fmt.Errorf("unable to add file to git, %w", err)
		}

		updatedFiles += 1
	}

	return updatedFiles > 0, nil
}

func updateRegistryValues(instance *codebaseService.Codebase, gitService *git.Service) (bool, error) {
	valuesStr, err := gitService.GetFileContents(registry.ValuesLocation)
	if err != nil {
		return false, errors.Wrap(err, "unable to get values from repo")
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal([]byte(valuesStr), &raw); err != nil {
		return false, errors.Wrap(err, "unable to decode values")
	}
	if raw == nil {
		raw = make(map[string]interface{})
	}

	var values map[string]interface{}
	if err := json.Unmarshal([]byte(instance.Annotations[registry.AnnotationValues]), &values); err != nil {
		return false, errors.Wrap(err, "unable to decode codebase values")
	}

	for k, v := range values {
		raw[k] = v
	}

	bts, err := yaml.Marshal(raw)
	if err != nil {
		return false, errors.Wrap(err, "unable to encode values yaml")
	}

	newContents := string(bts)

	if newContents != valuesStr {
		if err := gitService.SetFileContents(registry.ValuesLocation, newContents); err != nil {
			return false, errors.Wrap(err, "unable to set values yaml file contents")
		}

		if err := gitService.Add(registry.ValuesLocation); err != nil {
			return false, fmt.Errorf("unable to add values file to git, %w", err)
		}

		return true, nil

		//if err := gitService.Commit("set initial values.yaml from admin console",
		//	[]string{registry.ValuesLocation}, &git.User{
		//		Name:  instance.Annotations[registry.AnnotationCreatorUsername],
		//		Email: instance.Annotations[registry.AnnotationCreatorEmail],
		//	}); err != nil {
		//	return errors.Wrap(err, "unable to commit values yaml")
		//}
	}

	return false, nil
}

func (c *Controller) getGerritProject(ctx context.Context, name string) (*gerritService.GerritProject, error) {
	var projectList gerritService.GerritProjectList
	if err := c.k8sClient.List(ctx, &projectList); err != nil {
		return nil, errors.Wrap(err, "unable to list gerrit projects")
	}

	for _, prj := range projectList.Items {
		if prj.Spec.Name == name {
			return &prj, nil
		}
	}

	return nil, service.ErrNotFound("unable to find gerrit project")
}
