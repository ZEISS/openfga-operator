package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StoreSpec defines the desired state of Store
type StoreSpec struct {
	StoreRef string `json:"storeRef"`
	Model    string `json:"model"`
}

type StorePhase string

const (
	StorePhaseNone         StorePhase = ""
	StorePhasePending      StorePhase = "Pending"
	StorePhaseCreating     StorePhase = "Creating"
	StorePhaseSynchronized StorePhase = "Synchronized"
	StorePhaseFailed       StorePhase = "Failed"
)

// StoreStatus defines the observed state of Store
// +k8s:openapi-gen=true
type StoreStatus struct {
	// Phase is the current state of Store.
	Phase StorePhase `json:"phase"`
	// ControlPaused indicates the operator pauses the control of the store.
	ControlPaused bool `json:"controlPaused,omitempty"`
	// StoreID is the unique identifier of the store.
	StoreID string `json:"storeID"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type Store struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StoreSpec   `json:"spec,omitempty"`
	Status StoreStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// StoreList contains a list of Stores
type StoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Store `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Store{}, &StoreList{})
}
