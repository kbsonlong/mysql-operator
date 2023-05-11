/*
 * @Author: kbsonlong kbsonlong@gmail.com
 * @Date: 2023-05-06 16:09:35
 * @LastEditors: kbsonlong kbsonlong@gmail.com
 * @LastEditTime: 2023-05-11 13:08:22
 * @FilePath: /mysql-operator/Users/zengshenglong/Code/GoWorkSpace/operators/mysql-operator/api/v1/mysqlcluster_types.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
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

// MysqlClusterSpec defines the desired state of MysqlCluster
type MysqlClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The number of pods. This updates replicas filed
	// Defaults to 0
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Represents the MySQL version that will be run. The available version can be found here:
	// This field should be set even if the Image is set to let the operator know which mysql version is running.
	// Based on this version the operator can take decisions which features can be used.
	// 默认版本  5.7
	// +kubebuilder:validation:Required
	// +kubebuilder:default="5.7"
	MysqlVersion string `json:"mysqlVersion,omitempty"`

	Image     string `json:"image,omitempty"`
	InitImage string `json:"initImage,omitempty"`
}

// MysqlClusterStatus defines the observed state of MysqlCluster
type MysqlClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Replica int32 `json:"replica"`
	// LastScheduleTime metav1.Time `json:"lastScheduleTime,omitempty" protobuf:"bytes,8,opt,name=lastScheduleTime"`
	LastScheduleTime int32 `json:"lastScheduleTime"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Image",type="string",JSONPath=".spec.image",description="The Docker Image of MyAPP"
// +kubebuilder:printcolumn:JSONPath=".status.replica",name=Replica,type=integer
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="LastScheduleTime",type="integer",JSONPath=".status.lastScheduleTime"
// +kubebuilder:resource:shortName=msc
// +kubebuilder:subresource:status

// MysqlCluster is the Schema for the mysqlclusters API
type MysqlCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MysqlClusterSpec   `json:"spec,omitempty"`
	Status MysqlClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MysqlClusterList contains a list of MysqlCluster
type MysqlClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MysqlCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MysqlCluster{}, &MysqlClusterList{})
}
