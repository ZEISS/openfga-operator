package controllers

import (
	"context"
	"time"

	openfgav1alpha1 "github.com/zeiss/openfga-operator/api/v1alpha1"
	"github.com/zeiss/pkg/k8s/finalizers"
	"github.com/zeiss/pkg/utilx"

	fga "github.com/zeiss/openfga-operator/pkg/client"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	EventRecorderLabel = "openfga-store-controller"
)

type EventReason string

const (
	EventReasonStoreFetchFailed  EventReason = "StoreFetchFailed"
	EventReasonStoreCreateFailed EventReason = "StoreCreateFailed"
	EventReasonStoreUpdateFailed EventReason = "StoreUpdateFailed"
)

// StoreReconciler ...
type StoreReconciler struct {
	client.Client
	Clock
	FGA      *fga.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// NewStoreReconciler ...
func NewStoreReconciler(fga *fga.Client, mgr ctrl.Manager) *StoreReconciler {
	return &StoreReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(EventRecorderLabel),
		FGA:      fga,
	}
}

type Clock interface {
	Now() time.Time
}

//+kubebuilder:rbac:groups=openfga.zeiss.com,resources=stores,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openfga.zeiss.com,resources=stores/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openfga.zeiss.com,resources=stores/finalizers,verbs=update
//+kubebuilder:rbac:groups=,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile ...
func (r *StoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("reconcile store", "name", req.Name, "namespace", req.Namespace)

	store := &openfgav1alpha1.Store{}
	if err := r.Get(ctx, req.NamespacedName, store); err != nil {
		log.Error(err, "store not found", "store", req.NamespacedName)
		// Request object not found, could have been deleted after reconcile request.
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	if !store.ObjectMeta.DeletionTimestamp.IsZero() {
		if finalizers.HasFinalizer(store, openfgav1alpha1.FinalizerName) {
			err := r.reconcileDelete(ctx, store)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		// Delete
		return reconcile.Result{}, nil
	}

	// get the lastest version of the store instance before reconciling
	if err := r.Get(ctx, req.NamespacedName, store); err != nil {
		log.Error(err, "store not found", "store", req.NamespacedName)

		return reconcile.Result{}, err
	}

	if err := r.reconcileResources(ctx, store); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openfgav1alpha1.Store{}).
		Complete(r)
}

func (r *StoreReconciler) reconcileResources(ctx context.Context, s *openfgav1alpha1.Store) error {
	log := log.FromContext(ctx)

	err := r.reconcileStatus(ctx, s)
	if err != nil {
		log.Error(err, "failed to reconcile status", "name", s.Name, "namespace", s.Namespace)
		return err
	}

	err = r.reconcileStore(ctx, s)
	if err != nil {
		log.Error(err, "failed to reconcile store", "name", s.Name, "namespace", s.Namespace)
		return err
	}

	return nil
}

func (r *StoreReconciler) reconcileStore(ctx context.Context, store *openfgav1alpha1.Store) error {
	log := log.FromContext(ctx)

	log.Info("reconcile resource", "name", store.Name, "namespace", store.Namespace)

	if utilx.NotEmpty(store.Status.StoreID) {
		return nil
	}

	s, err := r.FGA.CreateStore(ctx, store.Name)
	if err != nil {
		return err
	}

	store.Finalizers = finalizers.AddFinalizer(store, openfgav1alpha1.FinalizerName)
	err = r.Update(ctx, store)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	store.Status.StoreID = s.ID
	err = r.Status().Update(ctx, store)
	if err != nil {
		return err
	}

	return nil
}

func (r *StoreReconciler) reconcileStatus(ctx context.Context, store *openfgav1alpha1.Store) error {
	log := log.FromContext(ctx)
	log.Info("reconcile status", "name", store.Name, "namespace", store.Namespace)

	phase := openfgav1alpha1.StorePhaseNone

	if utilx.Empty(store.Status.StoreID) {
		phase = openfgav1alpha1.StorePhaseCreating
	}

	if utilx.NotEmpty(store.Status.StoreID) {
		phase = openfgav1alpha1.StorePhaseSynchronized
	}

	if store.Status.Phase != phase {
		store.Status.Phase = phase

		return r.Status().Update(ctx, store)
	}

	return nil
}

func (r *StoreReconciler) reconcileDelete(ctx context.Context, s *openfgav1alpha1.Store) error {
	log := log.FromContext(ctx)

	log.Info("reconcile delete store", "name", s.Name, "namespace", s.Namespace)

	err := r.FGA.DeleteStore(ctx, s.Status.StoreID)
	if err != nil {
		return err
	}

	s.SetFinalizers(finalizers.RemoveFinalizer(s, openfgav1alpha1.FinalizerName))
	err = r.Update(ctx, s)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	return nil
}
