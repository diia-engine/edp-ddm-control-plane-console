package codebase

import (
	"context"
	"ddm-admin-console/app/registry"
	"ddm-admin-console/config"
	"ddm-admin-console/controller"
	"ddm-admin-console/service"
	codebaseService "ddm-admin-console/service/codebase"
	gerritService "ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/git"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"time"

	"github.com/patrickmn/go-cache"
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
)

const (
	registryRemoteName = "registry"
	keySecretIndex     = "id_rsa"
)

const DefaultRetryTimeout = time.Second * 15

type Controller struct {
	logger        controller.Logger
	mgr           ctrl.Manager
	k8sClient     client.Client
	cnf           *config.Settings
	appCache      *cache.Cache
	versionFilter *registry.VersionFilter
	gerrit        gerritService.ServiceInterface
	codebase      codebaseService.ServiceInterface
}

type AdminSyncer interface {
	SyncAdmins(ctx context.Context, registryName string, admins []registry.Admin) error
}

func Make(mgr ctrl.Manager, logger controller.Logger, cnf *config.Settings, _c *cache.Cache,
	gerrit gerritService.ServiceInterface, cbService codebaseService.ServiceInterface) error {
	c := Controller{
		mgr:       mgr,
		logger:    logger,
		k8sClient: mgr.GetClient(),
		cnf:       cnf,
		appCache:  _c,
		gerrit:    gerrit,
		codebase:  cbService,
	}

	vf, err := registry.MakeVersionFilter(cnf.RegistryVersionFilter)
	if err != nil {
		return fmt.Errorf("unable to init version filter, %w", err)
	}
	c.versionFilter = vf

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&codebaseService.Codebase{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: isSpecUpdated})).
		Complete(&c); err != nil {
		return fmt.Errorf("unable to create controller, %w", err)
	}

	return nil
}

func ProcessRegistryVersion(
	ctx context.Context,
	versionFilter *registry.VersionFilter,
	cb *codebaseService.Codebase,
	gr gerritService.ServiceInterface,
) (
	bool,
	error,
) {
	cbs := []codebaseService.Codebase{*cb}

	if err := registry.LoadRegistryVersions(ctx, gr, cbs); err != nil {
		return false, fmt.Errorf("unable to load registry version, %w", err)
	}

	registryVersionCodebase := cbs[0]

	if registryVersionCodebase.Version.Original() == "0" {
		clusterProject, err := gr.GetProject(ctx, cb.Name)
		if err != nil {
			return false, fmt.Errorf("unable to get cluster gerrit project, %w", err)
		}

		registryVersionCodebase.Version = registry.LowestVersion(registry.UpdateBranches(clusterProject.Status.Branches))
	}

	return versionFilter.CheckCodebase(&registryVersionCodebase), nil
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

func (c *Controller) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	c.logger.Infow("reconciling codebase", "Request.Namespace", request.Namespace,
		"Request.Name", request.Name)

	if err := c.reconcile(ctx, request); err != nil {
		c.logger.Errorw(err.Error(), "Request.Namespace", request.Namespace, "Request.Name", request.Name)

		if IsErrPostpone(err) {
			return reconcile.Result{RequeueAfter: errors.Unwrap(err).(ErrPostpone).D()}, nil
		}

		return reconcile.Result{RequeueAfter: DefaultRetryTimeout}, nil
	}

	c.logger.Infow("reconciling done", "Request.Namespace", request.Namespace,
		"Request.Name", request.Name)

	return reconcile.Result{}, nil
}

func (c *Controller) reconcile(ctx context.Context, request reconcile.Request) error {
	var instance codebaseService.Codebase
	if err := c.k8sClient.Get(ctx, request.NamespacedName, &instance); err != nil {
		if k8sErrors.IsNotFound(err) {
			c.logger.Infow("instance not found", "Request.Namespace", request.Namespace, "Request.Name", request.Name)
			return nil
		}

		return fmt.Errorf("unable to get codebase from k8s, %w", err)
	}

	if instance.Spec.Type != codebaseService.RegistryCodebaseType {
		return nil
	}

	processRequest, err := ProcessRegistryVersion(ctx, c.versionFilter, &instance, c.gerrit)
	if err != nil {
		return fmt.Errorf("unable to p, err: %w", err)
	}

	if !processRequest {
		c.logger.Infow("reconciling codebase skipped, wrong registry version",
			"Request.Namespace", request.Namespace, "Request.Name", request.Name)
		return nil
	}

	if err := c.updateImportRepo(ctx, &instance); err != nil {
		return fmt.Errorf("unable to update import repo, %w", err)
	}

	if err := c.checkBranchesStatus(ctx, &instance); err != nil {
		return fmt.Errorf("unable to check branches statuses, %w", err)
	}

	return nil
}

func (c *Controller) checkBranchesStatus(ctx context.Context, instance *codebaseService.Codebase) error {
	branches, err := c.codebase.GetBranchesByCodebase(ctx, instance.Name)
	if err != nil {
		return fmt.Errorf("unable to get codebase branches, %w", err)
	}

	inactiveBranches := false

	for _, b := range branches {
		if b.Status.Value != codebaseService.BranchStatusActive {
			inactiveBranches = true
			break
		}
	}

	if inactiveBranches {
		instance.Annotations[codebaseService.StatusAnnotation] = codebaseService.StatusAnnotationInactiveBranches
		if err := c.codebase.Update(ctx, instance); err != nil {
			return fmt.Errorf("unable to update instance, %w", err)
		}

		return ErrPostpone(DefaultRetryTimeout)
	}

	if annotationStatus, ok := instance.Annotations[codebaseService.StatusAnnotation]; ok &&
		annotationStatus == codebaseService.StatusAnnotationInactiveBranches {
		delete(instance.Annotations, codebaseService.StatusAnnotation)

		if err := c.codebase.Update(ctx, instance); err != nil {
			return fmt.Errorf("unable to update instance, %w", err)
		}
	}

	return nil
}

func (c *Controller) updateImportRepo(ctx context.Context, instance *codebaseService.Codebase) error {
	if instance.Spec.GitUrlPath == nil || *instance.Spec.GitUrlPath != codebaseService.RepoNotReady {
		return nil
	}

	prj, err := c.getGerritProject(ctx, instance.Name)
	if service.IsErrNotFound(err) {
		return ErrPostpone(DefaultRetryTimeout)
	} else if err != nil {
		return fmt.Errorf("unknown error, %w", err)
	}

	if prj.Status.Value != "OK" {
		return ErrPostpone(DefaultRetryTimeout)
	}

	if err := c.pushRegistryTemplate(ctx, instance); err != nil {
		return fmt.Errorf("unable to push registry template, %w", err)
	}

	gitUrlPath := fmt.Sprintf("/%s", instance.Name)
	instance.Spec.GitUrlPath = &gitUrlPath

	if err := c.k8sClient.Update(ctx, instance); err != nil {
		return fmt.Errorf("unable to update codebase, %w", err)
	}

	return nil
}

func (c *Controller) pushRegistryTemplate(ctx context.Context, instance *codebaseService.Codebase) error {
	reposPath, err := PrepareControllerTempFolder(c.cnf.TempFolder, "repos")
	if err != nil {
		return fmt.Errorf("unable to create repos folder, %w", err)
	}

	privateKey, err := GetGerritPrivateKey(ctx, c.k8sClient, c.cnf)
	if err != nil {
		return fmt.Errorf("unable to get gerrit private key, %w", err)
	}

	gitService := git.Make(path.Join(reposPath, instance.Name), c.cnf.GitUsername, privateKey)
	defer func() {
		if err := gitService.Clean(); err != nil {
			c.logger.Error(err)
		}
	}()

	if err := c.initCodebaseRepo(instance, gitService); err != nil {
		return fmt.Errorf("unable to init codebase repo, %w", err)
	}

	_, err = gitService.Pull(registryRemoteName)
	if git.IsErrReferenceNotFound(err) {
		if err := c.replaceDefaultBranch(instance, gitService); err != nil {
			return fmt.Errorf("unable to replace default branch, %w", err)
		}
	} else if git.IsErrNonFastForwardUpdate(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("unable to pull, %w", err)
	}

	cachedToCommit, err := SetCachedFiles(instance.Name, c.appCache, gitService)
	if err != nil {
		return fmt.Errorf("unable to set cached files, %w", err)
	}

	valuesToCommit, err := updateRegistryValues(instance, gitService)
	if err != nil {
		return fmt.Errorf("unable to update registry values, %w", err)
	}

	if cachedToCommit || valuesToCommit {
		if err := gitService.RawCommit(&git.User{Name: instance.Annotations[registry.AnnotationCreatorUsername],
			Email: instance.Annotations[registry.AnnotationCreatorEmail]}, "set initial values.yaml from admin console"); err != nil {
			return fmt.Errorf("unable to commit values, %w", err)
		}
	}

	if err := gitService.Push(registryRemoteName, "--all"); err != nil {
		return fmt.Errorf("unable to push [all] changes to repo registry, %w", err)
	}

	if err := gitService.Push(registryRemoteName, "--tags"); err != nil {
		return fmt.Errorf("unable to push tags to repo registry, %w", err)
	}

	return nil
}

func (c *Controller) replaceDefaultBranch(instance *codebaseService.Codebase, gitService *git.Service) error {
	if instance.Spec.BranchToCopyInDefaultBranch == "" {
		return gitService.Checkout(instance.Spec.DefaultBranch, false)
	}

	if err := gitService.Checkout(instance.Spec.BranchToCopyInDefaultBranch, false); err != nil {
		return fmt.Errorf("unable to checkout, %w", err)
	}

	if err := gitService.RemoveBranch(instance.Spec.DefaultBranch); err != nil {
		return fmt.Errorf("unable to remove default branch, %w", err)
	}

	if err := gitService.Checkout(instance.Spec.DefaultBranch, true); err != nil {
		return fmt.Errorf("unable to copy to default branch, %w", err)
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
		return fmt.Errorf("unable to clone repo, %w", err)
	}

	if err := gitService.AddRemote(registryRemoteName, fmt.Sprintf("%s/%s", gerritSSHURL, instance.Name)); err != nil {
		return fmt.Errorf("unable to add registry remote, %w", err)
	}

	return nil
}

func PrepareControllerTempFolder(tempDir, controllerFolder string) (string, error) {
	controllerFolderPath := path.Join(tempDir, fmt.Sprintf("%s-%d", controllerFolder, time.Now().UnixNano()))
	if _, err := os.Stat(controllerFolderPath); err == nil {
		if err := os.RemoveAll(controllerFolderPath); err != nil {
			return "", fmt.Errorf("unable to clear repos folder, %w", err)
		}
	}

	if err := os.MkdirAll(controllerFolderPath, 0777); err != nil {
		return "", fmt.Errorf("unable to create repo folder, %w", err)
	}

	return controllerFolderPath, nil
}

func GetGerritPrivateKey(ctx context.Context, k8sClient client.Client, cnf *config.Settings) (string, error) {
	var gerritSecret v1.Secret
	if err := k8sClient.Get(ctx, types.NamespacedName{Namespace: cnf.Namespace, Name: cnf.GitKeySecretName},
		&gerritSecret); err != nil {
		return "", fmt.Errorf("unable to get gerrit project creator secret, %w", err)
	}

	key, ok := gerritSecret.Data[keySecretIndex]
	if !ok {
		return "", fmt.Errorf("no data by key %s in gerrit secret", keySecretIndex)
	}

	return string(key), nil
}

func SetCachedFiles(projectName string, appCache *cache.Cache, gitService *git.Service) (bool, error) {
	//TODO: remove cached files
	key := registry.CachedFilesIndex(projectName)
	files, ok := appCache.Get(key)
	if !ok {
		return false, nil
	}

	cachedFiles, ok := files.([]registry.CachedFile)
	if !ok {
		appCache.Delete(key)
		return false, fmt.Errorf("wrong files type, %+v", files)
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
	currentValuesValuesStr, err := gitService.GetFileContents(registry.ValuesLocation)
	if err != nil {
		return false, fmt.Errorf("unable to get values from repo, %w", err)
	}

	var currentValues map[string]interface{}
	if err := yaml.Unmarshal([]byte(currentValuesValuesStr), &currentValues); err != nil {
		return false, fmt.Errorf("unable to decode values, %w", err)
	}
	if currentValues == nil {
		currentValues = make(map[string]interface{})
	}

	var instanceValues map[string]interface{}
	if err := json.Unmarshal([]byte(instance.Annotations[registry.AnnotationValues]), &instanceValues); err != nil {
		return false, fmt.Errorf("unable to decode codebase values, %w", err)
	}

	mergedValues := MergeMaps(currentValues, instanceValues)

	bts, err := yaml.Marshal(mergedValues)
	if err != nil {
		return false, fmt.Errorf("unable to encode values yaml, %w", err)
	}

	newContents := string(bts)

	if newContents != currentValuesValuesStr {
		if err := gitService.SetFileContents(registry.ValuesLocation, newContents); err != nil {
			return false, fmt.Errorf("unable to set values yaml file contents, %w", err)
		}

		if err := gitService.Add(registry.ValuesLocation); err != nil {
			return false, fmt.Errorf("unable to add values file to git, %w", err)
		}

		return true, nil
	}

	return false, nil
}

func MergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = MergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

func (c *Controller) getGerritProject(ctx context.Context, name string) (*gerritService.GerritProject, error) {
	var projectList gerritService.GerritProjectList
	if err := c.k8sClient.List(ctx, &projectList); err != nil {
		return nil, fmt.Errorf("unable to list gerrit projects, %w", err)
	}

	for _, prj := range projectList.Items {
		if prj.Spec.Name == name {
			return &prj, nil
		}
	}

	return nil, service.ErrNotFound("unable to find gerrit project")
}
