/*
Copyright 2021.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NGINXWebServerSpec defines the desired state of NGINXWebServer
type NGINXWebServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of NGINXWebServer. Edit nginxwebserver_types.go to remove/update
	Replicas int32  `json:"replicas"`
	Image    string `json:"image,omitempty"`
	DiskSize int32  `json:"diskSize,omitempty"`
}

// NGINXWebServerStatus defines the observed state of NGINXWebServer
type NGINXWebServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Replicas []string `json:"replicas"`
	Image    string   `json:"Image,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NGINXWebServer is the Schema for the nginxwebservers API
type NGINXWebServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NGINXWebServerSpec   `json:"spec,omitempty"`
	Status NGINXWebServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NGINXWebServerList contains a list of NGINXWebServer
type NGINXWebServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NGINXWebServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NGINXWebServer{}, &NGINXWebServerList{})
}
