package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type OpenFGAStoreSpec struct{}

type OpenFGAStoreStatus struct {
	StoreID string `json:"storeID"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type OpenFGAStore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenFGAStoreSpec   `json:"spec,omitempty"`
	Status OpenFGAStoreStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OpenFGAStoreList contains a list of NatsOperator
type OpenFGAStoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenFGAStore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OpenFGAStore{}, &OpenFGAStoreList{})
}
