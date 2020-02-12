package solar

import (
	"context"
	"reflect"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/client-go/tools/cache"

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

// Reconciler implements controller.Reconciler for Star resources.
type Reconciler struct {
	// Tracker builds an index of what resources are watching other resources
	// so that we can immediately react to changes tracked resources.
	Tracker tracker.Interface

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

func (r *Reconciler) reconcileDeployment(ctx context.Context, asvc *samplesv1alpha1.Star) error {
	logger := logging.FromContext(ctx)
	//
	//if err := r.Tracker.TrackReference(tracker.Reference{
	//	APIVersion: "v1",
	//	Kind:       "Service",
	//	Name:       asvc.Spec.Location,
	//	Namespace:  asvc.Namespace,
	//}, asvc); err != nil {
	//	logger.Errorf("Error tracking service %s: %v", asvc.Spec.Location, err)
	//	return err
	//}
	//
	//_, err := r.ServiceLister.Services(asvc.Namespace).Get(asvc.Spec.Location)
	//if apierrs.IsNotFound(err) {
	//	logger.Info("Service does not yet exist:", asvc.Spec.Location)
	//	asvc.Status.MarkServiceUnavailable(asvc.Spec.Location)
	//	return nil
	//} else if err != nil {
	//	logger.Errorf("Error reconciling service %s: %v", asvc.Spec.Location, err)
	//	return err
	//}

	logger.Info("The sun is created.")
	asvc.Status.MarkStarReady()
	return nil
}
