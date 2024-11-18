package v1alpha1

import (
	openfga "github.com/openfga/go-sdk"
	"github.com/openfga/go-sdk/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StoreData is the data structure to store the data of the store.
type StoreData struct {
	Name      string                            `json:"name"       yaml:"name"`
	Model     string                            `json:"model"      yaml:"model"`
	ModelFile string                            `json:"model_file" yaml:"model_file,omitempty"` //nolint:tagliatelle
	Tuples    []client.ClientContextualTupleKey `json:"tuples"     yaml:"tuples"`
	TupleFile string                            `json:"tuple_file" yaml:"tuple_file,omitempty"` //nolint:tagliatelle
	Tests     []ModelTest                       `json:"tests"      yaml:"tests"`
}

// ModelTest is the data structure to store the test of the model.
type ModelTest struct {
	Name        string                            `json:"name"         yaml:"name"`
	Description string                            `json:"description"  yaml:"description,omitempty"`
	Tuples      []client.ClientContextualTupleKey `json:"tuples"       yaml:"tuples,omitempty"`
	TupleFile   string                            `json:"tuple_file"   yaml:"tuple_file,omitempty"` //nolint:tagliatelle
	Check       []ModelTestCheck                  `json:"check"        yaml:"check"`
	ListObjects []ModelTestListObjects            `json:"list_objects" yaml:"list_objects,omitempty"` //nolint:tagliatelle
	ListUsers   []ModelTestListUsers              `json:"list_users"   yaml:"list_users,omitempty"`   //nolint:tagliatelle
}

// ModelTestCheck is the data structure to store the check of the model test.

type ModelTestCheck struct {
	User       string                  `json:"user"       yaml:"user"`
	Object     string                  `json:"object"     yaml:"object"`
	Context    *map[string]interface{} `json:"context"    yaml:"context,omitempty"`
	Assertions map[string]bool         `json:"assertions" yaml:"assertions"`
}

// ModelTestListObjects is the data structure to store the list objects of the model test.
type ModelTestListObjects struct {
	User       string                  `json:"user"       yaml:"user"`
	Type       string                  `json:"type"       yaml:"type"`
	Context    *map[string]interface{} `json:"context"    yaml:"context"`
	Assertions map[string][]string     `json:"assertions" yaml:"assertions"`
}

// ModelTestListUsers is the data structure to store the list users of the model test.
type ModelTestListUsers struct {
	Object     string                                 `json:"object"      yaml:"object"`
	UserFilter []openfga.UserTypeFilter               `json:"user_filter" yaml:"user_filter"` //nolint:tagliatelle
	Context    *map[string]interface{}                `json:"context"     yaml:"context,omitempty"`
	Assertions map[string]ModelTestListUsersAssertion `json:"assertions"  yaml:"assertions"`
}

// ModelTestListUsersAssertion is the data structure to store the assertion of the list users of the model test.
type ModelTestListUsersAssertion struct {
	Users []string `json:"users" yaml:"users"`
}

// OpenFGAStoreSpec defines the desired state of OpenFGAStore
type OpenFGAStoreSpec struct {
	StoreRef string    `json:"storeRef"`
	Store    StoreData `json:"store"`
}

type OpenFGAStorePhase string

const (
	OpenFGAStorePhaseNone         OpenFGAStorePhase = ""
	OpenFGAStorePhasePending      OpenFGAStorePhase = "Pending"
	OpenFGAStorePhaseCreating     OpenFGAStorePhase = "Creating"
	OpenFGAStorePhaseSynchronized OpenFGAStorePhase = "Synchronized"
	OpenFGAStorePhaseFailed       OpenFGAStorePhase = "Failed"
)

// OpenFGAStoreStatus defines the observed state of OpenFGAStore
// +k8s:openapi-gen=true
type OpenFGAStoreStatus struct {
	// Phase is the current state of OpenFGAStore.
	Phase OpenFGAStorePhase `json:"phase"`

	// ControlPaused indicates the operator pauses the control of
	// Octopinger.
	ControlPaused bool `json:"controlPaused,omitempty"`

	// StoreID is the unique identifier of the store.
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
