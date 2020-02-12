package solar

import (
	"context"
	"k8s.io/client-go/tools/cache"

	"knative.dev/pkg/tracker"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	deploymentinformer "knative.dev/pkg/client/injection/kube/informers/apps/v1/deployment"

	"my.dev/solar-system/pkg/apis/solar/v1alpha1"
	starinformer "my.dev/solar-system/pkg/client/injection/informers/solar/v1alpha1/star"
	starreconciler "my.dev/solar-system/pkg/client/injection/reconciler/solar/v1alpha1/star"
)

// NewController creates a Reconciler and returns the result of NewImpl.
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	logger := logging.FromContext(ctx)

	starInformer := starinformer.Get(ctx)
	deploymentInformer := deploymentinformer.Get(ctx)

	r := &Reconciler{
		deploymentLister: deploymentInformer.Lister(),
	}
	impl := starreconciler.NewImpl(ctx, r)
	r.Tracker = tracker.New(impl.EnqueueKey, controller.GetTrackerLease(ctx))

	logger.Info("Setting up event handlers.")

	starInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	deploymentInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.Filter(v1alpha1.SchemeGroupVersion.WithKind("Star")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}
