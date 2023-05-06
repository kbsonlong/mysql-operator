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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MysqlRestoreSpec defines the desired state of MysqlRestore
type MysqlRestoreSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of MysqlRestore. Edit mysqlrestore_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// MysqlRestoreStatus defines the observed state of MysqlRestore
type MysqlRestoreStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MysqlRestore is the Schema for the mysqlrestores API
type MysqlRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MysqlRestoreSpec   `json:"spec,omitempty"`
	Status MysqlRestoreStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MysqlRestoreList contains a list of MysqlRestore
type MysqlRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MysqlRestore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MysqlRestore{}, &MysqlRestoreList{})
}
