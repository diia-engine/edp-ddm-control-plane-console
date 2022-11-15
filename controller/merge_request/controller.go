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
	"path"
	"reflect"

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
// 3. backup values.yaml
// 4. checkout source branch
// 5. backup source branch
// 6. rebase from target branch with automatic conflict resolution
// 7. restore source branch
// 8. manually merge values.yaml from backup with new version from source branch
// 9. create new change to source branch
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

	//backup values yaml
	//os.
	return nil
}
