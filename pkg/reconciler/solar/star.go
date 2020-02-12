package solar

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/logging/logkey"
	"reflect"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/client-go/tools/cache"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	appsv1listers "k8s.io/client-go/listers/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clientset "my.dev/solar-system/pkg/client/clientset/versioned"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
	"knative.dev/pkg/tracker"
	samplesv1alpha1 "my.dev/solar-system/pkg/apis/solar/v1alpha1"
	listers "my.dev/solar-system/pkg/client/listers/solar/v1alpha1"
	"knative.dev/pkg/controller"
)

const(
	ImagePath = "docker.io/houshengbo/energy-source:latest"
)

// Reconciler implements controller.Reconciler for Star resources.
type Reconciler struct {
	// Tracker builds an index of what resources are watching other resources
	// so that we can immediately react to changes tracked resources.
	Tracker tracker.Interface

	KubeClientSet kubernetes.Interface
	starClient clientset.Interface
	// Listers index properties about resources
	deploymentLister    appsv1listers.DeploymentLister
	starLister listers.StarLister

}

// Check that our Reconciler implements Interface
var _ controller.Reconciler = (*Reconciler)(nil)

// Reconcile compares the actual state with the desired, and attempts to
// converge the two.
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return nil
	}
	// Get the KnativeEventing resource with this namespace/name.
	original, err := r.starLister.Stars(namespace).Get(name)
	if apierrs.IsNotFound(err) {
		return nil

	} else if err != nil {
		return err
	}

	// Don't modify the informers copy.
	star := original.DeepCopy()

	// Reconcile this copy of the KnativeEventing resource and then write back any status
	// updates regardless of whether the reconciliation errored out.
	reconcileErr := r.ReconcileKind(ctx, star)
	if equality.Semantic.DeepEqual(original.Status, star.Status) {
		// If we didn't change anything then don't call updateStatus.
		// This is important because the copy we loaded from the informer's
		// cache may be stale and we don't want to overwrite a prior update
		// to status with this stale state.
	} else if _, err = r.updateStatus(star); err != nil {
		return err
	}

	if reconcileErr != nil {
		return reconcileErr
	}
	return nil
}

func (r *Reconciler) updateStatus(desired *samplesv1alpha1.Star) (*samplesv1alpha1.Star, error) {
	ke, err := r.starClient.ExampleV1alpha1().Stars(desired.Namespace).Get(desired.Name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	// If there's nothing to update, just return.
	if reflect.DeepEqual(ke.Status, desired.Status) {
		return ke, nil
	}
	// Don't modify the informers copy
	existing := ke.DeepCopy()
	existing.Status = desired.Status
	return r.starClient.ExampleV1alpha1().Stars(desired.Namespace).UpdateStatus(existing)
}


// ReconcileKind implements Interface.ReconcileKind.
func (r *Reconciler) ReconcileKind(ctx context.Context, o *samplesv1alpha1.Star) reconciler.Event {
	if o.GetDeletionTimestamp() != nil {
		logger := logging.FromContext(ctx)
		logger.Info("The sun is removed.")
		// Check for a DeletionTimestamp.  If present, elide the normal reconcile logic.
		// When a controller needs finalizer handling, it would go here.
		return nil
	}
	o.Status.InitializeConditions()

	if err := r.reconcileDeployment(ctx, o); err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) reconcileDeployment(ctx context.Context, star *samplesv1alpha1.Star) error {
	ns := star.Namespace
	deploymentName := "energy-source"
	logger := logging.FromContext(ctx).With(zap.String(logkey.Deployment, deploymentName))

	deployment, err := r.deploymentLister.Deployments(ns).Get(deploymentName)
	if apierrs.IsNotFound(err) {
		// Deployment does not exist. Create it.
		star.Status.MarkDeploymentUnavailable(deploymentName)
		dep := r.newDeployment(star, deploymentName)
		deployment, err = r.createDeployment(ctx, dep)
		if err != nil {
			return fmt.Errorf("failed to create deployment %q: %w", deploymentName, err)
		}
		logger.Infof("Created deployment %q", deploymentName)
	} else if err != nil {
		return fmt.Errorf("failed to get deployment %q: %w", deploymentName, err)
	} else if !metav1.IsControlledBy(deployment, star) {
		// Surface an error in the star's status, and return an error.
		star.Status.MarkDeploymentUnavailable(deploymentName)
		return fmt.Errorf("revision: %q does not own Deployment: %q", star.Name, deploymentName)
	} else {
		// The deployment exists, but make sure that it has the shape that we expect.
		deployment, err = r.checkDeployment(ctx, star, deployment)
		if err != nil {
			return fmt.Errorf("failed to update deployment %q: %w", deploymentName, err)
		}
	}

	logger.Info("The sun is ready with the source of energy.")
	return nil
}

func (r *Reconciler) createDeployment(ctx context.Context, deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	return r.KubeClientSet.AppsV1().Deployments(deployment.Namespace).Create(deployment)
}

func (r *Reconciler) newDeployment(star *samplesv1alpha1.Star, name string) *appsv1.Deployment {
	labels := map[string]string{
		"app":        "source-of-energy",
		"controller": name,
	}
	replicas := int32(1)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: star.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(star, schema.GroupVersionKind{
					Group:   samplesv1alpha1.SchemeGroupVersion.Group,
					Version: samplesv1alpha1.SchemeGroupVersion.Version,
					Kind:    "Star",
				}),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: ImagePath,
						},
					},
				},
			},
		},
	}
}

func (r *Reconciler) checkDeployment(ctx context.Context, star *samplesv1alpha1.Star,
	deployment *appsv1.Deployment) (*appsv1.Deployment, error) {
	available := func(d *appsv1.Deployment) bool {
		for _, c := range d.Status.Conditions {
			if c.Type == appsv1.DeploymentAvailable && c.Status == corev1.ConditionTrue {
				return true
			}
		}
		return false
	}
	deployment, err := r.KubeClientSet.AppsV1().Deployments(deployment.GetNamespace()).Get(deployment.GetName(), metav1.GetOptions{})
	if err != nil {
		star.Status.MarkDeploymentUnavailable(deployment.Name)
		if apierrs.IsNotFound(err) {
			return deployment, nil
		}
		return deployment, err
	}
	if !available(deployment) {
		star.Status.MarkDeploymentUnavailable(deployment.Name)
		return deployment, nil
	}

	star.Status.MarkStarReady()
	return deployment, nil
}
