package controllers

import (
	"context"

	fga "github.com/zeiss/openfga-operator/pkg/client"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	openfgav1alpha1 "github.com/zeiss/openfga-operator/api/v1alpha1"
)

// OpenFGAStoreReconciler ...
type OpenFGAStoreReconciler struct {
	client.Client
	FGA    *fga.Client
	Scheme *runtime.Scheme
}

// NewOpenFGAStoreReconciler ...
func NewOpenFGAStoreReconciler(fga *fga.Client, mgr ctrl.Manager) *OpenFGAStoreReconciler {
	return &OpenFGAStoreReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		FGA:    fga,
	}
}

//+kubebuilder:rbac:groups=openfga.zeiss.com,resources=stores,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openfga.zeiss.com,resources=stores/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openfga.zeiss.com,resources=stores/finalizers,verbs=update
//+kubebuilder:rbac:groups=,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile ...
func (r *OpenFGAStoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("reconcile store")

	store := &openfgav1alpha1.Store{}

	err := r.Get(ctx, req.NamespacedName, store)
	if err != nil && errors.IsNotFound(err) {
		// Request object not found, could have been deleted after reconcile request.
		return reconcile.Result{}, nil
	}

	if err != nil {
		return reconcile.Result{}, err
	}

	if store.Status.Phase == openfgav1alpha1.StorePhaseSynchronized {
		return reconcile.Result{}, nil
	}

	// get the latest version of octopinger instance before reconciling
	err = r.Get(ctx, req.NamespacedName, store)
	if err != nil {
		return reconcile.Result{}, err
	}

	store.Status.Phase = openfgav1alpha1.StorePhaseCreating
	err = r.Status().Update(ctx, store)
	if err != nil {
		return reconcile.Result{}, err
	}

	// get the latest version of octopinger instance before reconciling
	err = r.Get(ctx, req.NamespacedName, store)
	if err != nil {
		return reconcile.Result{}, err
	}

	createdStore, err := r.FGA.CreateStore(ctx, store.Name)
	if err != nil {
		return reconcile.Result{}, err
	}

	store.Status.Phase = openfgav1alpha1.StorePhaseSynchronized
	store.Status.StoreID = createdStore.ID
	err = r.Status().Update(ctx, store)
	if err != nil {
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OpenFGAStoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openfgav1alpha1.Store{}).
		Complete(r)
}
