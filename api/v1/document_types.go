/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DocumentSpec defines the desired state of Document
type DocumentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Document. Edit document_types.go to remove/update
	Name  string `json:"name,omitempty"`
	Image string `json:"image,omitempty"`
	Route string `json:"route,omitempty"`
}

// DocumentStatus defines the observed state of Document
type DocumentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

//+kubebuilder:rbac:groups=uccps.uccps.document.domain,resources=documents,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=uccps.uccps.document.domain,resources=documents/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=uccps.uccps.document.domain,resources=documents/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;update;patch;watch;list;delete;create
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;update;patch;watch;list;delete;create

// Document is the Schema for the documents API
type Document struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DocumentSpec   `json:"spec,omitempty"`
	Status DocumentStatus `json:"status,omitempty"`
}

func (in *Document) DeepCopyObject() runtime.Object {
	//panic("implement me")
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

//+kubebuilder:object:root=true

// DocumentList contains a list of Document
type DocumentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Document `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Document{}, &DocumentList{})
}
