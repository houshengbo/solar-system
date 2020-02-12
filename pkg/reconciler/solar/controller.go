package solar

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	appsv1listers "k8s.io/client-go/listers/apps/v1"

	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"
	"knative.dev/pkg/tracker"
	samplesv1alpha1 "my.dev/solar-system/pkg/apis/solar/v1alpha1"
	starreconciler "my.dev/solar-system/pkg/client/injection/reconciler/solar/v1alpha1/star"
)

// newReconciledNormal makes a new reconciler event with event type Normal, and
// reason StarReconciled.
func newReconciledNormal(namespace, name string) reconciler.Event {
	return reconciler.NewEvent(corev1.EventTypeNormal, "StarReconciled", "Star reconciled: \"%s/%s\"", namespace, name)
}

// Reconciler implements addressableservicereconciler.Interface for
// AddressableService resources.
type Reconciler struct {
	// Tracker builds an index of what resources are watching other resources
	// so that we can immediately react to changes tracked resources.
	Tracker tracker.Interface

	// Listers index properties about resources
	deploymentLister    appsv1listers.DeploymentLister
}

// Check that our Reconciler implements Interface
var _ starreconciler.Interface = (*Reconciler)(nil)

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

	o.Status.ObservedGeneration = o.Generation
	return newReconciledNormal(o.Namespace, o.Name)
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
