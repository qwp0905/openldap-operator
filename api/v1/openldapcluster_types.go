package v1

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qwp0905/openldap-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	ConditionInitialized = "Initialized"
	ConditionReady       = "Ready"
	ConditionElected     = "Elected"
)

// OpenldapClusterSpec defines the desired state of OpenldapCluster
type OpenldapClusterSpec struct {
	// Openldap image based on qwp1216/openldap
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
	Ports *PortConfig `json:"ports,omitempty"`

	//+optional
	Monitor *MonitorConfig `json:"monitor,omitempty"`

	//+optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	//+optional
	NodeSelector *corev1.NodeSelector `json:"nodeSelector,omitempty"`
}

type PortConfig struct {
	//+kubebuilder:default:=1389
	//+kubebuilder:validation:Minimum:=1
	Ldap int32 `json:"ldap,omitempty"`

	//+kubebuilder:default:=1636
	//+kubebuilder:validation:Minimum:=1
	Ldaps int32 `json:"ldaps,omitempty"`
}

type StorageConfig struct {
	//+kubebuilder:validation:Required
	VolumeClaimTemplate corev1.PersistentVolumeClaimSpec `json:"volumeClaimTemplate,omitempty"`
}

type OpenldapConfig struct {
	//+optional
	Tls *TlsConfig `json:"tls,omitempty"`

	//+optional
	AdminPassword *corev1.SecretKeySelector `json:"adminPassword,omitempty"`

	//+optional
	//+kubebuilder:default:=admin
	AdminUsername string `json:"adminUsername,omitempty"`

	//+optional
	ConfigPassword *corev1.SecretKeySelector `json:"configPassword,omitempty"`

	//+optional
	//+kubebuilder:default:=config
	ConfigUsername string `json:"configUsername,omitempty"`

	//+kubebuilder:validation:Required
	Root string `json:"root,omitempty"`

	//+optional
	SeedData *SecretOrConfigMapVolumeSource `json:"seedData,omitempty"`

	//+optional
	Env []corev1.EnvVar `json:"env,omitempty"`
}

type SecretOrConfigMapVolumeSource struct {
	//+optional
	Secret *corev1.SecretVolumeSource `json:"secret,omitempty" protobuf:"bytes,6,opt,name=secret"`

	//+optional
	ConfigMap *corev1.ConfigMapVolumeSource `json:"configMap,omitempty" protobuf:"bytes,19,opt,name=configMap"`
}

type TlsConfig struct {
	//+kubebuilder:default:=false
	Enabled bool `json:"enabled,omitempty"`

	//+optional
	SecretName string `json:"secretName,omitempty"`

	//+optional
	CaFile string `json:"caFile,omitempty"`

	//+optional
	CertFile string `json:"certFile,omitempty"`

	//+optional
	KeyFile string `json:"keyFile,omitempty"`
}

type MonitorConfig struct {
	//+kubebuilder:default:=false
	Enabled bool `json:"enabled,omitempty"`

	//+optional
	Labels map[string]string `json:"labels,omitempty"`

	//+kubebuilder:default:="30s"
	Interval string `json:"interval,omitempty"`
}

// OpenldapClusterStatus defines the observed state of OpenldapCluster
type OpenldapClusterStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`

	CurrentMaster string `json:"currentMaster,omitempty"`

	DesiredMaster string `json:"desiredMaster,omitempty"`

	MasterFailure int `json:"masterFailure"`
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

const (
	defaultTlsEnabled     = false
	defaultRoot           = "dc=example,dc=com"
	defaultLogLevel       = "info"
	defaultMonitorEnabled = false
	defaultAdmin          = "admin"
	defaultConfig         = "config"
)

func (r *OpenldapCluster) SetDefault() {
	if r.Spec.OpenldapConfig == nil {
		r.Spec.OpenldapConfig = &OpenldapConfig{
			Root:           defaultRoot,
			AdminUsername:  defaultAdmin,
			ConfigUsername: defaultConfig,
		}
	}

	if r.Spec.OpenldapConfig.Tls == nil {
		r.Spec.OpenldapConfig.Tls = &TlsConfig{
			Enabled: defaultTlsEnabled,
		}
	} else if r.Spec.OpenldapConfig.Tls.Enabled {
		if r.Spec.OpenldapConfig.Tls.CaFile == "" {
			r.Spec.OpenldapConfig.Tls.CaFile = "ca.crt"
		}

		if r.Spec.OpenldapConfig.Tls.KeyFile == "" {
			r.Spec.OpenldapConfig.Tls.KeyFile = "cert.key"
		}

		if r.Spec.OpenldapConfig.Tls.CertFile == "" {
			r.Spec.OpenldapConfig.Tls.CertFile = "cert.crt"
		}
	}

	if r.Spec.OpenldapConfig.Root == "" {
		r.Spec.OpenldapConfig.Root = defaultRoot
	}

	if r.Spec.Monitor == nil {
		r.Spec.Monitor = &MonitorConfig{Enabled: defaultMonitorEnabled}
	}

	if r.Spec.Ports == nil {
		r.Spec.Ports = &PortConfig{
			Ldap:  1389,
			Ldaps: 1636,
		}
	}

	if r.Spec.Affinity == nil {
		r.Spec.Affinity = &corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						LabelSelector: &metav1.LabelSelector{
							MatchExpressions: []metav1.LabelSelectorRequirement{
								{
									Key:      "app.kubernetes.io/name",
									Operator: "In",
									Values:   []string{r.Name},
								},
							},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		}
	}
}

func (r *OpenldapCluster) ContainerProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: r.Spec.Ports.Ldap,
				},
			},
		},
		InitialDelaySeconds: 5,
		PeriodSeconds:       10,
		TimeoutSeconds:      1,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
}

func (r *OpenldapCluster) PodName(index int) string {
	return fmt.Sprintf(
		"%s-%s",
		r.Name,
		strconv.Itoa(index),
	)
}

func (r *OpenldapCluster) GetReplicas() int {
	return int(r.Spec.Replicas)
}

func (r *OpenldapCluster) SelectorLabels() map[string]string {
	version := "latest"

	if strings.Contains(r.Spec.Image, ":") {
		version = strings.Split(r.Spec.Image, ":")[1]
	}

	return map[string]string{
		"app.kubernetes.io/name":     r.Name,
		"app.kubernetes.io/instance": "openldap",
		"app.kubernetes.io/version":  version,
	}
}

func (r *OpenldapCluster) MasterSelectorLabels() map[string]string {
	return utils.MergeMap(
		r.SelectorLabels(),
		map[string]string{"app.kubernetes.io/component": "master"},
	)
}

func (r *OpenldapCluster) SlaveSelectorLabels() map[string]string {
	return utils.MergeMap(
		r.SelectorLabels(),
		map[string]string{"app.kubernetes.io/component": "slave"},
	)
}

func (r *OpenldapCluster) JobLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":     r.Name,
		"app.kubernetes.io/instance": "election",
	}
}

func (r *OpenldapCluster) ReadServiceName() string {
	return fmt.Sprintf("%s-read", r.Name)
}

func (r *OpenldapCluster) WriteServiceName() string {
	return fmt.Sprintf("%s-write", r.Name)
}

func (r *OpenldapCluster) MetricsServiceName() string {
	return fmt.Sprintf("%s-metrics", r.Name)
}

func (r *OpenldapCluster) ConfigMapName() string {
	return fmt.Sprintf("%s-config", r.Name)
}

func (r *OpenldapCluster) ExporterName() string {
	return fmt.Sprintf("%s-exporter", r.Name)
}

func (r *OpenldapCluster) ExporterImage() string {
	return "qwp1216/openldap-exporter:0.0.4"
}

func (r *OpenldapCluster) AdminDn() string {
	return fmt.Sprintf("cn=%s,%s", r.Spec.OpenldapConfig.AdminUsername, r.Spec.OpenldapConfig.Root)
}

func (r *OpenldapCluster) TlsEnabled() bool {
	return r.Spec.OpenldapConfig.Tls.Enabled
}

func (r *OpenldapCluster) TlsMountPath() string {
	return "/opt/bitnami/openldap/certs"
}

func (r *OpenldapCluster) MonitorEnabled() bool {
	return r.Spec.Monitor.Enabled
}

func (r *OpenldapCluster) MetricsPort() int32 {
	return 9142
}

func (r *OpenldapCluster) MetricsPortName() string {
	return "metrics"
}

func (r *OpenldapCluster) MetricsPath() string {
	return "/metrics"
}

func (r *OpenldapCluster) LdapPort() int32 {
	return r.Spec.Ports.Ldap
}

func (r *OpenldapCluster) LdapsPort() int32 {
	return r.Spec.Ports.Ldaps
}

func (r *OpenldapCluster) SeedDataPath() string {
	return "/ldifs"
}

func (r *OpenldapCluster) JobName() string {
	return fmt.Sprintf("%s-job", r.GetDesiredMaster())
}

func (r *OpenldapCluster) GetDesiredMaster() string {
	return r.Status.DesiredMaster
}

func (r *OpenldapCluster) GetCurrentMaster() string {
	return r.Status.CurrentMaster
}

func (r *OpenldapCluster) IsMasterSame() bool {
	return r.Status.CurrentMaster == r.Status.DesiredMaster
}

func (r *OpenldapCluster) UpdateDesiredMaster(index int) {
	r.Status.DesiredMaster = r.PodName(index)
}

func (r *OpenldapCluster) UpdateCurrentMaster() {
	r.Status.CurrentMaster = r.Status.DesiredMaster
}

func (r *OpenldapCluster) DeleteCurrentMaster() {
	r.Status.CurrentMaster = ""
}

func (r *OpenldapCluster) IsConditionsEmpty() bool {
	if r.Status.Conditions == nil {
		return true
	}

	return len(r.Status.Conditions) == 0
}

func (r *OpenldapCluster) GetCondition() metav1.Condition {
	return r.Status.Conditions[0]
}

func (r *OpenldapCluster) SetInitCondition() {
	r.Status.Conditions = []metav1.Condition{
		{
			Type:               ConditionInitialized,
			Status:             metav1.ConditionTrue,
			Reason:             ConditionInitialized,
			LastTransitionTime: metav1.Now(),
		},
		{
			Type:               ConditionReady,
			Status:             metav1.ConditionFalse,
			Reason:             string(metav1.StatusReasonServiceUnavailable),
			LastTransitionTime: metav1.Now(),
		},
		{
			Type:               ConditionElected,
			Status:             metav1.ConditionFalse,
			Reason:             string(metav1.StatusReasonServiceUnavailable),
			LastTransitionTime: metav1.Now(),
		},
	}
}

func (r *OpenldapCluster) IsInitialized() bool {
	for _, con := range r.Status.Conditions {
		if con.Type == ConditionInitialized && con.Status == metav1.ConditionTrue {
			return true
		}
	}

	return false
}

func (r *OpenldapCluster) IsReady() bool {
	for _, con := range r.Status.Conditions {
		if con.Type == ConditionReady && con.Status == metav1.ConditionTrue {
			return true
		}
	}

	return false
}

func (r *OpenldapCluster) IsElected() bool {
	for _, con := range r.Status.Conditions {
		if con.Type == ConditionElected && con.Status == metav1.ConditionTrue {
			return true
		}
	}

	return false
}

func (r *OpenldapCluster) SetConditionReady(condition bool) {
	conditions := []metav1.Condition{}

	for _, con := range r.Status.Conditions {
		if con.Type != ConditionReady {
			conditions = append(conditions, con)
		}
	}

	var status metav1.ConditionStatus
	if condition {
		status = metav1.ConditionTrue
	} else {
		status = metav1.ConditionFalse
	}

	r.Status.Conditions = append(conditions, metav1.Condition{
		Type:               ConditionReady,
		Status:             status,
		LastTransitionTime: metav1.Now(),
		Reason:             ConditionReady,
	})
}

func (r *OpenldapCluster) SetConditionElected(condition bool) {
	conditions := []metav1.Condition{}

	for _, con := range r.Status.Conditions {
		if con.Type != ConditionElected {
			conditions = append(conditions, con)
		}
	}

	var status metav1.ConditionStatus
	if condition {
		status = metav1.ConditionTrue
	} else {
		status = metav1.ConditionFalse
	}

	r.Status.Conditions = append(conditions, metav1.Condition{
		Type:               ConditionElected,
		Status:             status,
		LastTransitionTime: metav1.Now(),
		Reason:             ConditionElected,
	})
}

func (r *OpenldapCluster) DeleteInitializedCondition() {
	conditions := []metav1.Condition{}
	for _, con := range r.Status.Conditions {
		if con.Type != ConditionInitialized {
			conditions = append(conditions, con)
		}
	}

	r.Status.Conditions = conditions
}

func (r *OpenldapCluster) GetMasterFailure() int {
	return r.Status.MasterFailure
}

func init() {
	SchemeBuilder.Register(&OpenldapCluster{}, &OpenldapClusterList{})
}
