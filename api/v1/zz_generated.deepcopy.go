//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MonitorConfig) DeepCopyInto(out *MonitorConfig) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MonitorConfig.
func (in *MonitorConfig) DeepCopy() *MonitorConfig {
	if in == nil {
		return nil
	}
	out := new(MonitorConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenldapCluster) DeepCopyInto(out *OpenldapCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenldapCluster.
func (in *OpenldapCluster) DeepCopy() *OpenldapCluster {
	if in == nil {
		return nil
	}
	out := new(OpenldapCluster)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OpenldapCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenldapClusterList) DeepCopyInto(out *OpenldapClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OpenldapCluster, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenldapClusterList.
func (in *OpenldapClusterList) DeepCopy() *OpenldapClusterList {
	if in == nil {
		return nil
	}
	out := new(OpenldapClusterList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OpenldapClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenldapClusterSpec) DeepCopyInto(out *OpenldapClusterSpec) {
	*out = *in
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(corev1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	if in.Storage != nil {
		in, out := &in.Storage, &out.Storage
		*out = new(StorageConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.OpenldapConfig != nil {
		in, out := &in.OpenldapConfig, &out.OpenldapConfig
		*out = new(OpenldapConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.Ports != nil {
		in, out := &in.Ports, &out.Ports
		*out = new(PortConfig)
		**out = **in
	}
	if in.Monitor != nil {
		in, out := &in.Monitor, &out.Monitor
		*out = new(MonitorConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.Affinity != nil {
		in, out := &in.Affinity, &out.Affinity
		*out = new(corev1.Affinity)
		(*in).DeepCopyInto(*out)
	}
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = new(corev1.NodeSelector)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenldapClusterSpec.
func (in *OpenldapClusterSpec) DeepCopy() *OpenldapClusterSpec {
	if in == nil {
		return nil
	}
	out := new(OpenldapClusterSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenldapClusterStatus) DeepCopyInto(out *OpenldapClusterStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenldapClusterStatus.
func (in *OpenldapClusterStatus) DeepCopy() *OpenldapClusterStatus {
	if in == nil {
		return nil
	}
	out := new(OpenldapClusterStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OpenldapConfig) DeepCopyInto(out *OpenldapConfig) {
	*out = *in
	if in.Tls != nil {
		in, out := &in.Tls, &out.Tls
		*out = new(TlsConfig)
		**out = **in
	}
	if in.AdminPassword != nil {
		in, out := &in.AdminPassword, &out.AdminPassword
		*out = new(corev1.SecretKeySelector)
		(*in).DeepCopyInto(*out)
	}
	if in.ConfigPassword != nil {
		in, out := &in.ConfigPassword, &out.ConfigPassword
		*out = new(corev1.SecretKeySelector)
		(*in).DeepCopyInto(*out)
	}
	if in.SeedData != nil {
		in, out := &in.SeedData, &out.SeedData
		*out = new(SecretOrConfigMapVolumeSource)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OpenldapConfig.
func (in *OpenldapConfig) DeepCopy() *OpenldapConfig {
	if in == nil {
		return nil
	}
	out := new(OpenldapConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PortConfig) DeepCopyInto(out *PortConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PortConfig.
func (in *PortConfig) DeepCopy() *PortConfig {
	if in == nil {
		return nil
	}
	out := new(PortConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretOrConfigMapVolumeSource) DeepCopyInto(out *SecretOrConfigMapVolumeSource) {
	*out = *in
	if in.Secret != nil {
		in, out := &in.Secret, &out.Secret
		*out = new(corev1.SecretVolumeSource)
		(*in).DeepCopyInto(*out)
	}
	if in.ConfigMap != nil {
		in, out := &in.ConfigMap, &out.ConfigMap
		*out = new(corev1.ConfigMapVolumeSource)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SecretOrConfigMapVolumeSource.
func (in *SecretOrConfigMapVolumeSource) DeepCopy() *SecretOrConfigMapVolumeSource {
	if in == nil {
		return nil
	}
	out := new(SecretOrConfigMapVolumeSource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StorageConfig) DeepCopyInto(out *StorageConfig) {
	*out = *in
	in.VolumeClaimTemplate.DeepCopyInto(&out.VolumeClaimTemplate)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StorageConfig.
func (in *StorageConfig) DeepCopy() *StorageConfig {
	if in == nil {
		return nil
	}
	out := new(StorageConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TlsConfig) DeepCopyInto(out *TlsConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TlsConfig.
func (in *TlsConfig) DeepCopy() *TlsConfig {
	if in == nil {
		return nil
	}
	out := new(TlsConfig)
	in.DeepCopyInto(out)
	return out
}
