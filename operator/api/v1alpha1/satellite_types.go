/*
Copyright 2026.

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
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SatelliteSpec defines the desired state of Satellite
// SatelliteSpec defines the desired state of Satellite
type SatelliteSpec struct {
	// Region is the logical region/site identifier (e.g. "hcm", "hn", "dn")
	// +kubebuilder:validation:Required
	Region string `json:"region"`
}

// SatelliteStatus defines the observed state of Satellite
type SatelliteStatus struct {
	// Phase reflects the Satellite's current lifecycle state
	// +kubebuilder:validation:Enum=Pending;Ready;Error;Unreachable
	Phase string `json:"phase,omitempty"`

	// ManagedBy indicates whether this resource is managed via GitOps/Operator
	ManagedBy string `json:"managedBy,omitempty"`

	// LastHeartbeat is the last heartbeat timestamp received from the Satellite Agent
	LastHeartbeat *metav1.Time `json:"lastHeartbeat,omitempty"`

	// conditions represent the current state of the Satellite resource.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Satellite is the Schema for the satellites API
type Satellite struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of Satellite
	// +required
	Spec SatelliteSpec `json:"spec"`

	// status defines the observed state of Satellite
	// +optional
	Status SatelliteStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// SatelliteList contains a list of Satellite
type SatelliteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []Satellite `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &Satellite{}, &SatelliteList{})
		return nil
	})
}
