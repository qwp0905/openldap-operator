package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// OpenldapClusterSpec defines the desired state of OpenldapCluster
type OpenldapClusterSpec struct {
	// Openldap image based on osixia/openldap
	Image string `json:"image,omitempty"`

	//+optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	//+kubebuilder:default:=1
	//+kubebuilder:validation:Minimum:=1
	Replicas int32 `json:"replicas,omitempty"`

	//+optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	//+optional
	Storage *StorageConfig `json:"storage,omitempty"`

	//+optional
	OpenldapConfig *OpenldapConfig `json:"openldapConfig,omitempty"`

	//+optional
	Ports int32 `json:"port,omitempty"`

	//+optional
	Monitor *MonitorConfig `json:"monitor,omitempty"`

	//+optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	//+optional
	NodeSelector *corev1.NodeSelector `json:"nodeSelector,omitempty"`
}

type PortConfig struct {
	//+kubebuilder:default:=389
	//+kubebuilder:validation:Minimum:=1
	Ldap int32 `json:"ldap,omitempty"`

	//+kubebuilder:default:=636
	//+kubebuilder:validation:Minimum:=1
	Ldaps int32 `json:"ldaps,omitempty"`
}

type StorageConfig struct {
	//+kubebuilder:validation:Required
	VolumeClaimTemplate *corev1.PersistentVolumeClaimSpec `json:"volumeClaimTemplate,omitempty"`
}

type OpenldapConfig struct {
	//+optional
	Tls *TlsConfig `json:"tls,omitempty"`

	//+optional
	AdminPassword *corev1.SecretKeySelector `json:"adminPassword,omitempty"`

	//+kubebuilder:validation:Required
	Domain string `json:"domain,omitempty"`

	//+optional
	Organization string `json:"organization,omitempty"`

	//+kubebuilder:default:=mdb
	Backend string `json:"backend,omitempty"`

	//+optional
	SeedData *SecretOrConfigMapVolume `json:"seedData,omitempty"`

	//+kubebuilder:default:=256
	//+kubebuilder:validation:Enum:=-1;0;1;2;4;8;16;32;64;128;256;512;1024;2048;16384;32768
	LogLevel int32 `json:"logLevel,omitempty"`
}

type SecretOrConfigMapVolume struct {
	//+kubebuilder:validation:Required
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`

	SecretOrConfigMapVolumeSource `json:",inline" protobuf:"bytes,2,opt,name=volumeSource"`
}

type SecretOrConfigMapVolumeSource struct {
	Secret *corev1.SecretVolumeSource `json:"secret,omitempty" protobuf:"bytes,6,opt,name=secret"`

	ConfigMap *corev1.ConfigMapVolumeSource `json:"configMap,omitempty" protobuf:"bytes,19,opt,name=configMap"`
}

type TlsConfig struct {
	//+kubebuilder:default:=true
	Enabled bool `json:"enabled,omitempty"`

	//+kubebuilder:default:=false
	Enforced bool `json:"enforced,omitempty"`

	//+optional
	CaFile *corev1.SecretKeySelector `json:"caFile,omitempty"`

	//+optional
	CertFile *corev1.SecretKeySelector `json:"certFile,omitempty"`

	//+optional
	KeyFile *corev1.SecretKeySelector `json:"keyFile,omitempty"`
}

type MonitorConfig struct {
	//+kubebuilder:default:=false
	Enabled bool `json:"enabled,omitempty"`

	//+optional
	Labels map[string]string `json:"labels,omitempty"`

	//+kubebuilder:default:=30s
	Interval int32 `json:"interval,omitempty"`
}

// OpenldapClusterStatus defines the observed state of OpenldapCluster
type OpenldapClusterStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OpenldapCluster is the Schema for the openldapclusters API
type OpenldapCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenldapClusterSpec   `json:"spec,omitempty"`
	Status OpenldapClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OpenldapClusterList contains a list of OpenldapCluster
type OpenldapClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenldapCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OpenldapCluster{}, &OpenldapClusterList{})
}
