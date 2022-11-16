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

	"gopkg.in/yaml.v3"

	"github.com/pkg/errors"
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
}

func Make(mgr ctrl.Manager, logger controller.Logger, cnf *config.Settings) error {
	c := Controller{
		mgr:       mgr,
		logger:    logger,
		k8sClient: mgr.GetClient(),
		cnf:       cnf,
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&gerritService.GerritMergeRequest{}, builder.WithPredicates(predicate.Funcs{
			UpdateFunc: isSpecUpdated})).
		Complete(&c); err != nil {
		return errors.Wrap(err, "unable to create controller")
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

	return
}

// actions
// 1. clone repo
// 2. checkout target branch
// 3. backup values.yaml if it exists
// 4. checkout source branch
// 5. backup source branch
// 6. rebase from target branch with automatic conflict resolution
// 7. restore source branch
// 8. manually merge values.yaml from backup with new version from source branch if it backup`ed
// 9. create new change to source branch if it not clean
// 10. apply and submit change
// 11. set merge request cr spec source branch to pass it to gerrit operator

func (c *Controller) prepareMergeRequest(ctx context.Context, instance *gerritService.GerritMergeRequest) error {
	if instance.Labels[registry.MRLabelAction] != registry.MRLabelActionBranchMerge || instance.Spec.SourceBranch != "" {
		return nil
	}

	controllerFolderPath, err := codebase.PrepareControllerTempFolder(c.cnf.TempFolder, "merge-requests")
	if err != nil {
		return errors.Wrap(err, "unable to create merge-requests folder")
	}

	backupFolder, err := codebase.PrepareControllerTempFolder(c.cnf.TempFolder, "mr-backup")
	if err != nil {
		return errors.Wrap(err, "unable to create backup folder")
	}

	privateKey, err := codebase.GetGerritPrivateKey(ctx, c.k8sClient, c.cnf)
	if err != nil {
		return errors.Wrap(err, "unable to get gerrit private key")
	}

	gitService := git.Make(path.Join(controllerFolderPath, instance.Spec.ProjectName), c.cnf.GitUsername, privateKey)
	defer func() {
		if err := gitService.Clean(); err != nil {
			c.logger.Error(err)
		}
	}()

	targetBranch, ok := instance.Labels[registry.MRLabelTargetBranch]
	if !ok {
		return errors.New("target branch is not specified")
	}

	sourceBranch, ok := instance.Labels[registry.MRLabelSourceBranch]
	if !ok {
		return errors.New("source branch is not specified")
	}

	gerritSSHURL := codebase.GerritSSHURL(c.cnf)
	if err := gitService.Clone(fmt.Sprintf("%s/%s", gerritSSHURL, instance.Spec.ProjectName)); err != nil {
		return errors.Wrap(err, "unable to clone repo")
	}

	if err := gitService.Checkout(targetBranch, false); err != nil {
		return errors.Wrap(err, "unable to checkout branch")
	}

	projectPath := path.Join(controllerFolderPath, instance.Spec.ProjectName)

	//backup values.yaml
	valuesBackupPath := path.Join(backupFolder, "backup-values.yaml")
	projectValuesPath := path.Join(projectPath, registry.ValuesLocation)
	if err := CopyFile(projectValuesPath, valuesBackupPath); err != nil {
		return errors.Wrap(err, "unable to backup values yaml")
	}

	if err := gitService.Checkout(sourceBranch, false); err != nil {
		return errors.Wrap(err, "unable to checkout")
	}

	//backup source branch
	projectBackupPath := path.Join(backupFolder, fmt.Sprintf("backup-%s", instance.Spec.ProjectName))
	if err := CopyFolder(projectPath, projectBackupPath); err != nil {
		return errors.Wrap(err, "unable to backup source branch")
	}

	if msg, err := gitService.Rebase(targetBranch, "-X", "ours"); err != nil {
		return errors.Wrapf(err, "unable to rebase from target branch: %s", msg)
	}

	if err := CopyFolder(projectBackupPath, projectPath); err != nil {
		return errors.Wrap(err, "unable to restore source branch")
	}

	if err := MergeValuesFiles(valuesBackupPath, projectValuesPath); err != nil {
		return errors.Wrap(err, "unable to merge values")
	}

	if err := gitService.Add("."); err != nil {
		return errors.Wrap(err, "unable to add all files")
	}

	//check for change id!
	if err := gitService.Commit(fmt.Sprintf("update branch values.yaml from [%s] branch", targetBranch),
		[]string{}, &git.User{Name: instance.Spec.AuthorName, Email: instance.Spec.AuthorEmail}); err != nil {
		return errors.Wrap(err, "unable to commit changes")
	}

	//restore source branch

	return nil
}

func MergeValuesFiles(src, dst string) error {
	srcFp, err := os.Open(src)
	if err != nil {
		return errors.Wrap(err, "unable to open src file")
	}

	dstFp, err := os.Open(dst)
	if err != nil {
		return errors.Wrap(err, "unable to open dst file")
	}

	var (
		srcData map[string]interface{}
		dstData map[string]interface{}
	)

	if err := yaml.NewDecoder(srcFp).Decode(&srcData); err != nil {
		return errors.Wrap(err, "unable to decode src values")
	}

	if err := yaml.NewDecoder(dstFp).Decode(&dstData); err != nil {
		return errors.Wrap(err, "unable to decode dst values")
	}

	if err := srcFp.Close(); err != nil {
		return errors.Wrap(err, "unable to close src")
	}

	if err := dstFp.Close(); err != nil {
		return errors.Wrap(err, "unable to close dst")
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
		return errors.Wrap(err, "unable to recreate dst")
	}

	if err := yaml.NewEncoder(dstFp).Encode(dstData); err != nil {
		return errors.Wrap(err, "unable to encode dst data")
	}

	if err := dstFp.Close(); err != nil {
		return errors.Wrap(err, "unable to close dst")
	}

	return nil
}

func CopyFolder(src, dst string) error {
	cmd := exec.Command("cp", "-r", src, dst)
	bts, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrap(err, "unable to copy folder")
	}

	if len(bts) > 0 {
		return errors.New(string(bts))
	}

	return nil
}

func CopyFile(src, dst string) error {
	if _, err := os.Stat(dst); err == nil {
		if err := os.Remove(dst); err != nil {
			return errors.Wrap(err, "unable to remove file")
		}
	}

	srcFp, err := os.Open(src)
	if err != nil {
		return errors.Wrap(err, "unable to open file")
	}

	dstFp, err := os.Create(dst)
	if err != nil {
		return errors.Wrap(err, "unable to create file")
	}

	if _, err := io.Copy(dstFp, srcFp); err != nil {
		return errors.Wrap(err, "unable to copy files")
	}

	if err := srcFp.Close(); err != nil {
		return errors.Wrap(err, "unable to close file")
	}

	if err := dstFp.Close(); err != nil {
		return errors.Wrap(err, "unable to close file")
	}

	return nil
}
