package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ModelSpec defines the desired state of Store
type ModelSpec struct {
	StoreRef StoreRef `json:"storeRef"`
	Model    string   `json:"model"`
}

// StoreRef defines the reference to the store.
type StoreRef struct {
	// Name is the name of the store.
	Name string `json:"name"`
}

type ModelPhase string

const (
	ModelPhaseNone         ModelPhase = ""
	ModelPhasePending      ModelPhase = "Pending"
	ModelPhaseCreating     ModelPhase = "Creating"
	ModelPhaseSynchronized ModelPhase = "Synchronized"
	ModelPhaseFailed       ModelPhase = "Failed"
)

// ModelStatus defines the observed state of the Model
// +k8s:openapi-gen=true
type ModelStatus struct {
	// Phase is the current state of Store.
	Phase ModelPhase `json:"phase"`
	// ControlPaused indicates the operator pauses the control of the store.
	ControlPaused bool `json:"controlPaused,omitempty"`
	// InstanceID is the unique identifier of the store.
	InstanceID string `json:"instanceID"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type Model struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ModelSpec   `json:"spec,omitempty"`
	Status ModelStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ModelList contains a list of Models
type ModelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Model `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Model{}, &ModelList{})
}
