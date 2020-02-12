package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)

var condSet = apis.NewLivingConditionSet(
	DeploymentsAvailable,
	CreationSucceeded,
)

func (es *StarStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return condSet.Manage(es).GetCondition(t)
}

func (es *StarStatus) duck() *duckv1.Status {
	return &es.Status
}

// GetGroupVersionKind implements kmeta.OwnerRefable
func (as *Star) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("Star")
}

func (ass *StarStatus) InitializeConditions() {
	condSet.Manage(ass).InitializeConditions()
}

func (ass *StarStatus) MarkDeploymentUnavailable(name string) {
	condSet.Manage(ass).MarkFalse(
		DeploymentsAvailable,
		"DeploymentUnavailable",
		"Deployment %q wasn't found.", name)
}

func (es *StarStatus) MarkStarReady() {
	condSet.Manage(es).MarkTrue(CreationSucceeded)
}

func (ass *StarStatus) MarkDeploymentAvailable() {
	condSet.Manage(ass).MarkTrue(DeploymentsAvailable)
}
