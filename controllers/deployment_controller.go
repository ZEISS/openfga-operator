package controllers

import (
	"context"
	"strings"
	"time"

	openfgav1alpha1 "github.com/zeiss/openfga-operator/api/v1alpha1"
	fga "github.com/zeiss/openfga-operator/pkg/client"
	"github.com/zeiss/pkg/cast"
	"github.com/zeiss/pkg/mapx"
	"github.com/zeiss/pkg/slices"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	EventReasonDeploymentEnvUpdated EventReason = "DeploymentEnvUpdated"
)

// PodReconciler ...
type PodReconciler struct {
	client.Client
	Clock
	FGA      *fga.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// NewPodReconciler ...
func NewPodReconciler(fga *fga.Client, mgr ctrl.Manager) *PodReconciler {
	return &PodReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor(EventRecorderLabel),
		FGA:      fga,
	}
}

//+kubebuilder:rbac:groups=openfga.zeiss.com,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openfga.zeiss.com,resources=deployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openfga.zeiss.com,resources=deployments/finalizers,verbs=update

// Reconcile ...
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("reconcile model", "name", req.Name, "namespace", req.Namespace)

	deployment := &appsv1.Deployment{}
	if err := r.Get(ctx, req.NamespacedName, deployment); err != nil {
		log.Error(err, "deployment not found", "store", req.NamespacedName)
		// Request object not found, could have been deleted after reconcile request.
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	if !deployment.ObjectMeta.DeletionTimestamp.IsZero() {
		// Delete
		return reconcile.Result{}, nil
	}

	if err := r.reconcileResources(ctx, deployment); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{})).
		Complete(r)
}

func (r *PodReconciler) reconcileResources(ctx context.Context, deployment *appsv1.Deployment) error {
	log := log.FromContext(ctx)

	log.Info("reconcile openfga deployment", "name", deployment.Name, "namespace", deployment.Namespace)

	err := r.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
	if err != nil {
		return err
	}

	annotations := deployment.GetAnnotations()
	for k, v := range deployment.Annotations {
		if strings.HasPrefix(k, ModelAnnotationPrefix) {
			annotations[strings.TrimPrefix(k, ModelAnnotationPrefix)] = v
		}
	}

	if !mapx.Exists(annotations, "ref") {
		return nil
	}

	model := &openfgav1alpha1.Model{
		ObjectMeta: metav1.ObjectMeta{
			Name:      annotations["ref"],
			Namespace: deployment.Namespace,
		},
	}
	if err := r.Get(ctx, client.ObjectKeyFromObject(model), model); err != nil {
		return client.IgnoreNotFound(err)
	}

	store := &openfgav1alpha1.Store{
		ObjectMeta: metav1.ObjectMeta{
			Name:      model.Spec.StoreRef.Name,
			Namespace: deployment.Namespace,
		},
	}
	if err := r.Get(ctx, client.ObjectKeyFromObject(store), store); err != nil {
		return client.IgnoreNotFound(err)
	}

	env := []corev1.EnvVar{
		{
			Name:  "OPENFGA_MODEL_INSTANCE_ID",
			Value: model.Status.InstanceID,
		},
		{
			Name:  "OPENFGA_MODEL_STORE_ID",
			Value: store.Status.StoreID,
		},
	}

	for i, container := range deployment.Spec.Template.Spec.Containers {
		deployment.Spec.Template.Spec.Containers[i].Env = slices.Unique(func(v corev1.EnvVar) string { return v.Name }, slices.Append(env, container.Env...)...)
	}

	if mapx.Exists(annotations, ModelUpdatedAnnotation) {
		return nil
	}

	annotations[ModelUpdatedAnnotation] = time.Now().Format(time.RFC3339)
	deployment.SetAnnotations(annotations)

	if err := r.Update(ctx, deployment); err != nil {
		return err
	}

	r.Recorder.Event(deployment, corev1.EventTypeNormal, cast.String(EventReasonDeploymentEnvUpdated), "OpenFGA model instance added to the environment")

	return nil
}
