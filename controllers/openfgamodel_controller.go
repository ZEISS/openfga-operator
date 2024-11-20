package controllers

import (
	"context"

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

// OpenFGAModelReconciler ...
type OpenFGAModelReconciler struct {
	client.Client
	Clock
	FGA      *fga.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// NewOpenFGAModelReconciler ...
func NewOpenFGAModelReconciler(fga *fga.Client, mgr ctrl.Manager) *OpenFGAModelReconciler {
	return &OpenFGAModelReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(EventRecorderLabel),
		FGA:      fga,
	}
}

//+kubebuilder:rbac:groups=openfga.zeiss.com,resources=models,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openfga.zeiss.com,resources=models/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openfga.zeiss.com,resources=models/finalizers,verbs=update
//+kubebuilder:rbac:groups=,resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile ...
func (r *OpenFGAModelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("reconcile model", "name", req.Name, "namespace", req.Namespace)

	model := &openfgav1alpha1.Model{}
	if err := r.Get(ctx, req.NamespacedName, model); err != nil {
		log.Error(err, "model not found", "model", req.NamespacedName)
		// Request object not found, could have been deleted after reconcile request.
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// get the lastest version of the model instance before reconciling
	if err := r.Get(ctx, req.NamespacedName, model); err != nil {
		log.Error(err, "model not found", "model", req.NamespacedName)

		return reconcile.Result{}, err
	}

	if err := r.reconcileResources(ctx, model); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OpenFGAModelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openfgav1alpha1.Model{}).
		Complete(r)
}

func (r *OpenFGAModelReconciler) reconcileResources(ctx context.Context, model *openfgav1alpha1.Model) error {
	return nil
}

func (r *OpenFGAModelReconciler) reconcileStore(ctx context.Context, model *openfgav1alpha1.Model) error {
	return nil
}

func (r *OpenFGAModelReconciler) reconcileStatus(ctx context.Context, model *openfgav1alpha1.Model) error {
	log := log.FromContext(ctx)
	log.Info("change status", "name", model.Name, "namespace", model.Namespace)

	phase := openfgav1alpha1.ModelPhaseNone

	if utilx.Empty(model.Status.InstanceID) {
		phase = openfgav1alpha1.ModelPhaseCreating
	}

	if utilx.NotEmpty(model.Status.InstanceID) {
		phase = openfgav1alpha1.ModelPhaseSynchronized
	}

	if model.Status.Phase != phase {
		model.Status.Phase = phase

		return r.Status().Update(ctx, model)
	}

	return nil
}

func (r *OpenFGAModelReconciler) reconcileDelete(ctx context.Context, s *openfgav1alpha1.Store) error {
	log := log.FromContext(ctx)

	log.Info("delete store", "name", s.Name, "namespace", s.Namespace)

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
