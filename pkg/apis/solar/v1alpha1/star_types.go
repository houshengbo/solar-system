package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	duckv1 "knative.dev/pkg/apis/duck/v1"

	"knative.dev/pkg/apis"
)

const (
	CreationSucceeded     apis.ConditionType = "CreationSucceeded"
	DeploymentsAvailable  apis.ConditionType = "DeploymentsAvailable"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Star is a the basic unit to form the solar system.
// +k8s:openapi-gen=true
type Star struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the Star (from the client).
	// +optional
	Spec StarSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the Star (from the controller).
	// +optional
	Status StarStatus `json:"status,omitempty"`
}

// StarSpec holds the desired state of the Star (from the client).
// +k8s:openapi-gen=true
type StarSpec struct {
	Type string `json:"type"`
	Location string `json:"location"`
}

// StarStatus communicates the observed state of the Star (from the controller).
// +k8s:openapi-gen=true
type StarStatus struct {
	duckv1.Status `json:",inline"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StarList is a list of Star resources
type StarList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Star `json:"items"`
}
