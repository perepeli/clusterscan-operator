/*
Copyright 2024.

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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterScanSpec defines the desired state of ClusterScan
type ClusterScanSpec struct {
	// Specify the schedule for recurring jobs (cron syntax). Empty for one-off jobs.
	Schedule string `json:"schedule,omitempty"`
	// JobTemplate specifies the job to run for the scan
	JobTemplate batchv1.JobSpec `json:"jobTemplate"`
}

type ClusterScanStatus struct {
	LastScheduleTime *metav1.Time             `json:"lastScheduleTime,omitempty"`
	Active           []corev1.ObjectReference `json:"active,omitempty"`
	Conditions       []metav1.Condition       `json:"conditions,omitempty"`
	UnifiedStatus    string                   `json:"unifiedStatus,omitempty"`
	Message          string                   `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type ClusterScan struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterScanSpec   `json:"spec,omitempty"`
	Status ClusterScanStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type ClusterScanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterScan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterScan{}, &ClusterScanList{})
}
