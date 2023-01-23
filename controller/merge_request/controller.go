package merge_request

import (
	"context"
	"ddm-admin-console/app/registry"
	"ddm-admin-console/config"
	"ddm-admin-console/controller"
	"ddm-admin-console/controller/codebase"
	gerritService "ddm-admin-console/service/gerrit"
	"ddm-admin-console/service/git"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"reflect"

	"k8s.io/apimachinery/pkg/types"

	k8sController "sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Controller struct {
	logger    controller.Logger
	mgr       ctrl.Manager
	k8sClient client.Client
	cnf       *config.Settings
	gerrit    gerritService.ServiceInterface
}

func Make(mgr ctrl.Manager, logger controller.Logger, cnf *config.Settings, gerrit gerritService.ServiceInterface) error {
	c := Controller{
		mgr:       mgr,
		logger:    logger,
		k8sClient: mgr.GetClient(),
		cnf:       cnf,
		gerrit:    gerrit,
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&gerritService.GerritMergeRequest{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: isSpecUpdated})).WithOptions(k8sController.Options{
		MaxConcurrentReconciles: 1,
	}).
		Complete(&c); err != nil {
		return fmt.Errorf("unable to create controller, err: %w", err)
	}

	return nil
}

func isSpecUpdated(e event.UpdateEvent) bool {
	oo := e.ObjectOld.(*gerritService.GerritMergeRequest)
	no := e.ObjectNew.(*gerritService.GerritMergeRequest)

	return !reflect.DeepEqual(oo.Spec, no.Spec) ||
		(oo.GetDeletionTimestamp().IsZero() && !no.GetDeletionTimestamp().IsZero())
}

func (c *Controller) Reconcile(ctx context.Context, request reconcile.Request) (result reconcile.Result,
	resultErr error) {
	defer func() {
		if resultErr != nil {
			c.logger.Errorw("reconciling merge request", "Request.Namespace", request.Namespace,
				"Request.Name", request.Name, "error", fmt.Sprintf("%+v", resultErr))
		}
	}()

	c.logger.Infow("reconciling merge request", "Request.Namespace", request.Namespace,
		"Request.Name", request.Name)

	var instance gerritService.GerritMergeRequest
	if err := c.k8sClient.Get(ctx, request.NamespacedName, &instance); err != nil {
		if k8sErrors.IsNotFound(err) {
			c.logger.Infow("instance not found", "Request.Namespace", request.Namespace, "Request.Name", request.Name)
			return
		}

		resultErr = fmt.Errorf("unable to get merge request from k8s, err: %w", err)
		return
	}

	if err := c.prepareMergeRequest(ctx, &instance); err != nil {
		resultErr = fmt.Errorf("unable to prepare merge request, err: %w", err)
		return
	}

	if err := c.autoApproveMergeRequest(&instance); err != nil {
		resultErr = fmt.Errorf("unable to approve MR, err: %w", err)
		return
	}

	c.logger.Infow("reconciling merge request done", "Request.Namespace", request.Namespace,
		"Request.Name", request.Name)

	return
}

func (c *Controller) autoApproveMergeRequest(instance *gerritService.GerritMergeRequest) error {
	if instance.Status.ChangeID == "" || instance.Status.Value != gerritService.StatusNew {
		return nil
	}

	label, ok := instance.Labels[registry.MRLabelApprove]
	if !ok || label != registry.MRLabelApproveAuto {
		return nil
	}

	if err := c.gerrit.ApproveAndSubmitChange(instance.Status.ChangeID, instance.Spec.AuthorName,
		instance.Spec.AuthorEmail); err != nil {
		return fmt.Errorf("unable to approve and submit change, err: %w", err)
	}

	return nil
}

// actions
// 1. clone repo
// 2. checkout target branch
// 3. backup values.yaml if it exists
// 4. checkout source branch
// 5. backup source branch
// 5.1 checkout to target branch
// 6. delete source branch
// 6.1 checkout -b source branch from target branch
// 7. restore source branch
// 8. manually merge values.yaml from backup with new version from source branch if it backup`ed
// 9. create new change to source branch with new commit
// 10. apply and submit change
// 11. set merge request cr spec source branch to pass it to gerrit operator

func (c *Controller) prepareMergeRequest(ctx context.Context, instance *gerritService.GerritMergeRequest) error {
	if instance.Labels[registry.MRLabelAction] != registry.MRLabelActionBranchMerge ||
		instance.Spec.SourceBranch != "" || instance.Status.ChangeID != "" {
		c.logger.Infow("nothing need to be done", "Request.Namespace", instance.Namespace,
			"Request.Name", instance.Name)
		return nil
	}

	_, backupFolderPath, projectPath, err := prepareControllerFolders(c.cnf.TempFolder, instance.Spec.ProjectName)
	if err != nil {
		return fmt.Errorf("unable to prepare controller tmp folders, err: %w", err)
	}

	gitService, err := c.initGitService(ctx, projectPath)
	if err != nil {
		return fmt.Errorf("unable to init git service, err: %w", err)
	}

	targetBranch, sourceBranch, err := getBranchesFromLabels(instance.Labels)
	if err != nil {
		return fmt.Errorf("unable to get branches from instance labels, err: %w", err)
	}

	if err := gitService.Clone(fmt.Sprintf("%s/%s", codebase.GerritSSHURL(c.cnf), instance.Spec.ProjectName)); err != nil {
		return fmt.Errorf("unable to clone repo, err: %w", err)
	}

	if err := gitService.RawCheckout(targetBranch, false); err != nil {
		return fmt.Errorf("unable to checkout branch, err: %w", err)
	}
	//backup values.yaml
	valuesBackupPath := path.Join(backupFolderPath, "backup-values.yaml")
	projectValuesPath := path.Join(projectPath, registry.ValuesLocation)
	if err := CopyFile(projectValuesPath, valuesBackupPath); err != nil {
		return fmt.Errorf("unable to backup values yaml, err: %w", err)
	}

	if err := gitService.RawCheckout(sourceBranch, false); err != nil {
		return fmt.Errorf("unable to checkout, err: %w", err)
	}
	//backup source branch
	projectBackupPath, err := backupProject(backupFolderPath, projectPath, instance.Spec.ProjectName)
	if err != nil {
		return fmt.Errorf("unable to backup source branch, err: %w", err)
	}
	//checkout to target branch
	if err := gitService.RawCheckout(targetBranch, false); err != nil {
		return fmt.Errorf("unable to checkout, err: %w", err)
	}
	//delete source branch
	if err := gitService.DeleteBranch(sourceBranch); err != nil {
		return fmt.Errorf("unable to delete source branch, err: %w", err)
	}
	//recreate source branch
	if err := gitService.RawCheckout(sourceBranch, true); err != nil {
		return fmt.Errorf("unable to checkout to source branch, err: %w", err)
	}
	//restore source branch
	if err := CopyFolder(fmt.Sprintf("%s/.", projectBackupPath), fmt.Sprintf("%s/", projectPath)); err != nil {
		return fmt.Errorf("unable to restore source branch, err: %w", err)
	}
	//merge values from target branch
	if err := MergeValuesFiles(valuesBackupPath, projectValuesPath); err != nil {
		return fmt.Errorf("unable to merge values, err: %w", err)
	}
	//add all changes
	if err := gitService.Add("."); err != nil {
		return fmt.Errorf("unable to add all files, err: %w", err)
	}

	changeID, err := gitService.GenerateChangeID()
	if err != nil {
		return fmt.Errorf("unable to generate change id, err: %w", err)
	}

	if err := gitService.SetAuthor(&git.User{Name: instance.Spec.AuthorName, Email: instance.Spec.AuthorEmail}); err != nil {
		return fmt.Errorf("unable to set author, err: %w", err)
	}

	if err := gitService.RawCommit(
		git.CommitMessageWithChangeID(
			fmt.Sprintf("Add new branch %s\n\nupdate branch values.yaml from [%s] branch", sourceBranch,
				targetBranch), changeID)); err != nil {
		return fmt.Errorf("unable to commit changes, err: %w", err)
	}

	if err := gitService.Push("origin", fmt.Sprintf("refs/heads/%s:%s", sourceBranch, sourceBranch), "--force"); err != nil {
		return fmt.Errorf("unable to push repo, err: %w", err)
	}

	var reloadInstance gerritService.GerritMergeRequest
	if err := c.k8sClient.Get(ctx, types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name}, &reloadInstance); err != nil {
		return fmt.Errorf("unable to reload instance, err: %w", err)
	}

	reloadInstance.Spec.SourceBranch = sourceBranch
	if err := c.k8sClient.Update(ctx, &reloadInstance); err != nil {
		return fmt.Errorf("unable to update merge request, err: %w", err)
	}

	return nil
}

func backupProject(backupFolderPath, projectPath, projectName string) (string, error) {
	projectBackupPath := path.Join(backupFolderPath, fmt.Sprintf("backup-%s", projectName))
	if err := CopyFolder(projectPath, projectBackupPath); err != nil {
		return "", fmt.Errorf("unable to backup source branch, err: %w", err)
	}
	if err := os.RemoveAll(path.Join(projectBackupPath, ".git")); err != nil {
		return "", fmt.Errorf("unable to remove .git folder from backup, err: %w", err)
	}

	return projectBackupPath, nil
}

func (c *Controller) initGitService(ctx context.Context, projectPath string) (*git.Service, error) {
	privateKey, err := codebase.GetGerritPrivateKey(ctx, c.k8sClient, c.cnf)
	if err != nil {
		return nil, fmt.Errorf("unable to get gerrit private key, err: %w", err)
	}

	return git.Make(projectPath, c.cnf.GitUsername, privateKey), nil
}

func getBranchesFromLabels(labels map[string]string) (targetBranch, sourceBranch string, err error) {
	targetBranch, ok := labels[registry.MRLabelTargetBranch]
	if !ok {
		err = errors.New("target branch is not specified")
		return
	}

	sourceBranch, ok = labels[registry.MRLabelSourceBranch]
	if !ok {
		err = errors.New("source branch is not specified")
		return
	}

	return
}

func prepareControllerFolders(tempFolder, projectName string) (controllerFolderPath, backupFolderPath, projectPath string, retErr error) {
	controllerFolderPath, retErr = codebase.PrepareControllerTempFolder(tempFolder, "merge-requests")
	if retErr != nil {
		return
	}

	backupFolderPath, retErr = codebase.PrepareControllerTempFolder(tempFolder, "mr-backup")
	if retErr != nil {
		return
	}

	projectPath = path.Join(controllerFolderPath, projectName)

	return
}

func MergeValuesFiles(src, dst string) error {
	srcFp, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("unable to open src file, err: %w", err)
	}

	dstFp, err := os.Open(dst)
	if err != nil {
		return fmt.Errorf("unable to open dst file, err: %w", err)
	}

	var (
		srcData map[string]interface{}
		dstData map[string]interface{}
	)

	if err := yaml.NewDecoder(srcFp).Decode(&srcData); err != nil {
		return fmt.Errorf("unable to decode src values, err: %w", err)
	}

	if err := yaml.NewDecoder(dstFp).Decode(&dstData); err != nil {
		return fmt.Errorf("unable to decode dst values, err: %w", err)
	}

	if err := srcFp.Close(); err != nil {
		return fmt.Errorf("unable to close src, err: %w", err)
	}

	if err := dstFp.Close(); err != nil {
		return fmt.Errorf("unable to close dst, err: %w", err)
	}

	//merge global
	dstGlobal, ok := dstData["global"]
	if !ok {
		dstGlobal = make(map[string]interface{})
	}
	dstGlobalDict := dstGlobal.(map[string]interface{})

	srcGlobal, ok := srcData["global"]
	if !ok {
		srcGlobal = make(map[string]interface{})
	}
	srcGlobalDict := srcGlobal.(map[string]interface{})

	for k, v := range srcGlobalDict {
		dstGlobalDict[k] = v
	}

	//merge else
	for k, v := range srcData {
		dstData[k] = v
	}
	//rewrite global
	dstData["global"] = dstGlobalDict

	dstFp, err = os.Create(dst)
	if err != nil {
		return fmt.Errorf("unable to recreate dst, err: %w", err)
	}

	if err := yaml.NewEncoder(dstFp).Encode(dstData); err != nil {
		return fmt.Errorf("unable to encode dst data, err: %w", err)
	}

	if err := dstFp.Close(); err != nil {
		return fmt.Errorf("unable to close dst, err: %w", err)
	}

	return nil
}

func CopyFolder(src, dst string) error {
	cmd := exec.Command("cp", "-r", src, dst)
	var msg string
	bts, err := cmd.CombinedOutput()
	if len(bts) > 0 {
		msg = string(bts)
	}

	if err != nil {
		return fmt.Errorf("unable to copy folder %s, err: %w", msg, err)
	}

	return nil
}

func CopyFile(src, dst string) error {
	if _, err := os.Stat(dst); err == nil {
		if err := os.Remove(dst); err != nil {
			return fmt.Errorf("unable to remove file, err: %w", err)
		}
	}

	srcFp, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("unable to open file, err: %w", err)
	}

	dstFp, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("unable to create file, err: %w", err)
	}

	if _, err := io.Copy(dstFp, srcFp); err != nil {
		return fmt.Errorf("unable to copy files, err: %w", err)
	}

	if err := srcFp.Close(); err != nil {
		return fmt.Errorf("unable to close file, err: %w", err)
	}

	if err := dstFp.Close(); err != nil {
		return fmt.Errorf("unable to close file, err: %w", err)
	}

	return nil
}
