package codebase

//
//import (
//	"context"
//	"reflect"
//
//	"github.com/pkg/errors"
//	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
//	ctrl "sigs.k8s.io/controller-runtime"
//	"sigs.k8s.io/controller-runtime/pkg/builder"
//	"sigs.k8s.io/controller-runtime/pkg/event"
//	"sigs.k8s.io/controller-runtime/pkg/predicate"
//	"sigs.k8s.io/controller-runtime/pkg/reconcile"
//)
//
//func isSpecUpdated(e event.UpdateEvent) bool {
//	oo := e.ObjectOld.(*Codebase)
//	no := e.ObjectNew.(*Codebase)
//
//	return !reflect.DeepEqual(oo.Spec, no.Spec) ||
//		(oo.GetDeletionTimestamp().IsZero() && !no.GetDeletionTimestamp().IsZero())
//}
//
//func (s *Service) RunController() error {
//	//TODO: move manager from service to main if there is any other controllers in app
//	cfg := ctrl.GetConfigOrDie()
//
//	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
//		Scheme:    s.scheme,
//		Namespace: s.namespace,
//	})
//
//	if err != nil {
//		return errors.Wrap(err, "unable to init manager")
//	}
//
//	if err := ctrl.NewControllerManagedBy(mgr).
//		For(&Codebase{}, builder.WithPredicates(predicate.Funcs{
//			UpdateFunc: isSpecUpdated})).
//		Complete(s); err != nil {
//		return errors.Wrap(err, "unable to create controller")
//	}
//
//	go func() {
//		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
//			s.logger.Error(err, "unable to start manager")
//		}
//	}()
//
//	return nil
//}
//
//func (s *Service) Reconcile(ctx context.Context, request reconcile.Request) (result reconcile.Result,
//	resultErr error) {
//	s.logger.Infow("reconciling codebase", "Request.Namespace", request.Namespace,
//		"Request.Name", request.Name)
//
//	var instance Codebase
//	if err := s.k8sClient.Get(ctx, request.NamespacedName, &instance); err != nil {
//		if k8sErrors.IsNotFound(err) {
//			s.logger.Infow("instance not found", "Request.Namespace", request.Namespace, "Request.Name", request.Name)
//			return
//		}
//
//		resultErr = errors.Wrap(err, "unable to get codebase from k8s")
//		return
//	}
//
//	if err := s.reconcile(ctx, &instance); err != nil {
//		resultErr = errors.Wrap(err, "unable to reconcile codebase")
//		return
//	}
//
//	s.logger.Infow("reconciling done", "Request.Namespace", request.Namespace,
//		"Request.Name", request.Name)
//
//	return
//}
//
//func (s *Service) reconcile(ctx context.Context, instance *Codebase) error {
//	s.logger.Info(instance.Name)
//
//	//TODO: remove in future release
//	if err := s.migrateAdminAnnotations(ctx, instance); err != nil {
//		return errors.Wrap(err, "unable to migrate admin annotations")
//	}
//
//	return nil
//}
//
////deprecated
//func (s *Service) migrateAdminAnnotations(ctx context.Context, instance *Codebase) error {
//	s.logger.Info(instance.Annotations)
//	return nil
//}
