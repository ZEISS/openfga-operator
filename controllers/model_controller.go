package controllers

import (
	"context"

	openfgav1alpha1 "github.com/zeiss/openfga-operator/api/v1alpha1"
	"github.com/zeiss/pkg/cast"
	"github.com/zeiss/pkg/k8s"
	"github.com/zeiss/pkg/k8s/finalizers"
	"github.com/zeiss/pkg/utilx"

	fga "github.com/zeiss/openfga-operator/pkg/client"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	ModelAnnotationPrefix  = "openfga.zeiss.com/model."
	ModelUpdatedAnnotation = ModelAnnotationPrefix + "updated-at"
)

const (
	EventReasonModelCreated EventReason = "ModelCreated"
	EventReasonModelUpdated EventReason = "ModelUpdated"
	EventReasonModelDeleted EventReason = "ModelDeleted"
	EventReasonModelFailed  EventReason = "ModelFailed"
)

// ModelReconciler ...
type ModelReconciler struct {
	client.Client
	Clock
	FGA      *fga.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// NewModelReconciler ...
func NewModelReconciler(fga *fga.Client, mgr ctrl.Manager) *ModelReconciler {
	return &ModelReconciler{
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
func (r *ModelReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

	if !model.ObjectMeta.DeletionTimestamp.IsZero() {
		if finalizers.HasFinalizer(model, openfgav1alpha1.FinalizerName) {
			err := r.reconcileDelete(ctx, model)
			if err != nil {
				return ctrl.Result{}, err
			}

			// Delete
			return reconcile.Result{}, nil
		}
	}

	if err := r.reconcileResources(ctx, model); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ModelReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openfgav1alpha1.Model{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(r)
}

func (r *ModelReconciler) reconcileResources(ctx context.Context, model *openfgav1alpha1.Model) error {
	log := log.FromContext(ctx)

	err := r.reconcileStatus(ctx, model)
	if err != nil {
		log.Error(err, "failed to reconcile status", "name", model.Name, "namespace", model.Namespace)
		return err
	}

	err = r.reconcileModel(ctx, model)
	if err != nil {
		log.Error(err, "failed to reconcile model", "name", model.Name, "namespace", model.Namespace)
		return err
	}

	return nil
}

func (r *ModelReconciler) reconcileModel(ctx context.Context, model *openfgav1alpha1.Model) error {
	log := log.FromContext(ctx)

	log.Info("reconcile model", "name", model.Name, "namespace", model.Namespace)

	store := &openfgav1alpha1.Store{}
	err := k8s.FetchObject(ctx, r.Client, model.Namespace, model.Spec.StoreRef.Name, store)
	if err != nil {
		return err
	}

	log.Info("update model in store", "name", store.Name, "namespace", store.Namespace)

	m, err := r.FGA.UpdateModel(ctx, store.Status.StoreID, model.Spec.Model)
	if err != nil {
		log.Error(err, "failed to update model", "name", model.Name, "namespace", model.Namespace)

		model.Status.Phase = openfgav1alpha1.ModelPhaseFailed
		return r.Status().Update(ctx, model)
	}

	err = controllerutil.SetOwnerReference(store, model, r.Scheme)
	if err != nil {
		return err
	}

	model.Finalizers = finalizers.AddFinalizer(model, openfgav1alpha1.FinalizerName)
	err = r.Update(ctx, model)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	model.Status.InstanceID = m.ID
	model.Status.Phase = openfgav1alpha1.ModelPhaseSynchronized
	err = r.Status().Update(ctx, model)
	if err != nil {
		return err
	}

	r.Recorder.Event(model, corev1.EventTypeNormal, cast.String(EventReasonModelUpdated), "model updated")

	return nil
}

func (r *ModelReconciler) reconcileStatus(ctx context.Context, model *openfgav1alpha1.Model) error {
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

func (r *ModelReconciler) reconcileDelete(ctx context.Context, model *openfgav1alpha1.Model) error {
	log := log.FromContext(ctx)

	log.Info("delete model", "name", model.Name, "namespace", model.Namespace)

	model.SetFinalizers(finalizers.RemoveFinalizer(model, openfgav1alpha1.FinalizerName))
	err := r.Update(ctx, model)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	return nil
}
