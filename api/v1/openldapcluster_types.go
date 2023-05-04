package v1

import (
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	Ports *PortConfig `json:"port,omitempty"`

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
	VolumeClaimTemplate corev1.PersistentVolumeClaimSpec `json:"volumeClaimTemplate,omitempty"`
}

type OpenldapConfig struct {
	//+optional
	Tls *TlsConfig `json:"tls,omitempty"`

	//+optional
	AdminPassword *corev1.SecretKeySelector `json:"adminPassword,omitempty"`

	//+optional
	AdminUsername *corev1.SecretKeySelector `json:"adminUsername,omitempty"`

	//+optional
	ConfigPassword *corev1.SecretKeySelector `json:"configPassword,omitempty"`

	//+optional
	ConfigUsername *corev1.SecretKeySelector `json:"configUsername,omitempty"`

	//+kubebuilder:validation:Required
	Domain string `json:"domain,omitempty"`

	//+optional
	Organization string `json:"organization,omitempty"`

	//+kubebuilder:default:="mdb"
	Backend string `json:"backend,omitempty"`

	//+optional
	SeedData *SecretOrConfigMapVolumeSource `json:"seedData,omitempty"`

	//+kubebuilder:default:=info
	//+kubebuilder:validation:Enum:=info;error;warn;debug;trace
	LogLevel string `json:"logLevel,omitempty"`
}

type SecretOrConfigMapVolumeSource struct {
	//+optional
	Secret *corev1.SecretVolumeSource `json:"secret,omitempty" protobuf:"bytes,6,opt,name=secret"`

	//+optional
	ConfigMap *corev1.ConfigMapVolumeSource `json:"configMap,omitempty" protobuf:"bytes,19,opt,name=configMap"`
}

type TlsConfig struct {
	//+kubebuilder:default:=true
	Enabled bool `json:"enabled,omitempty"`

	//+kubebuilder:default:=false
	Enforced bool `json:"enforced,omitempty"`

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
	defaultTlsEnabled     = true
	defaultTlsEnforced    = false
	defaultDomain         = "example.com"
	defaultBackend        = "mdb"
	defaultLogLevel       = "info"
	defaultMonitorEnabled = false
)

func (r *OpenldapCluster) SetDefault() {
	if r.Spec.OpenldapConfig == nil {
		r.Spec.OpenldapConfig = &OpenldapConfig{
			Backend:  defaultBackend,
			LogLevel: defaultLogLevel,
			Domain:   defaultDomain,
		}
	}

	if r.Spec.OpenldapConfig.Tls == nil {
		r.Spec.OpenldapConfig.Tls = &TlsConfig{
			Enabled:  defaultTlsEnabled,
			Enforced: defaultTlsEnforced,
		}
	}

	if r.Spec.OpenldapConfig.Domain == "" {
		r.Spec.OpenldapConfig.Domain = defaultDomain
	}

	if r.Spec.OpenldapConfig.Backend == "" {
		r.Spec.OpenldapConfig.Backend = defaultBackend
	}

	if r.Spec.Monitor == nil {
		r.Spec.Monitor = &MonitorConfig{Enabled: defaultMonitorEnabled}
	}

	if r.Spec.Ports == nil {
		r.Spec.Ports = &PortConfig{
			Ldap:  389,
			Ldaps: 636,
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

func (r *OpenldapCluster) SelectorLabels() map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":     r.Name,
		"app.kubernetes.io/instance": "openldap",
	}
}

func (r *OpenldapCluster) ConfigMapName() string {
	return fmt.Sprintf("%s-config", r.Name)
}

func (r *OpenldapCluster) InitContainerName() string {
	return fmt.Sprintf("init-%s", r.Name)
}

func (r *OpenldapCluster) MonitorConfigMapName() string {
	return fmt.Sprintf("%s-monitor", r.Name)
}

func (r *OpenldapCluster) ExporterName() string {
	return fmt.Sprintf("%s-exporter", r.Name)
}

func (r *OpenldapCluster) BindDn() string {
	dnList := []string{}
	for _, el := range strings.Split(r.Spec.OpenldapConfig.Domain, ".") {
		dnList = append(dnList, fmt.Sprintf("dc=%s", el))
	}

	return fmt.Sprintf("cn=admin,%s", strings.Join(dnList, ","))
}

func (r *OpenldapCluster) ReplicationHosts() string {
	pods := []string{}
	for i := 0; i < int(r.Spec.Replicas); i++ {
		pods = append(
			pods,
			fmt.Sprintf(
				"'%s-%s.%s.%s.svc:%s'",
				r.Name,
				strconv.Itoa(i),
				r.Name,
				r.Namespace,
				strconv.Itoa(389),
			),
		)
	}

	return fmt.Sprintf("#PYTHON2BASH:[%s]", strings.Join(pods, ","))
}

func (r *OpenldapCluster) TlsEnabled() bool {
	return r.Spec.OpenldapConfig.Tls.Enabled
}

func (r *OpenldapCluster) MonitorEnabled() bool {
	return r.Spec.Monitor.Enabled
}

func (r *OpenldapCluster) ModuleMonitorLdif() string {
	return `dn: cn=module{0},cn=config
changetype: modify
add: olcModuleLoad
olcModuleLoad: {1}back_monitor`
}

func (r *OpenldapCluster) DatabaseModuleLdif() string {
	return fmt.Sprintf(`dn: olcDatabase={2}Monitor,cn=config
objectClass: olcDatabaseConfig
objectClass: olcMonitorConfig
olcDatabase: {2}Monitor
olcAccess: {0}to dn.subtree="cn=Monitor" by dn.base="%s" read by * none`, r.BindDn())
}

func (r *OpenldapCluster) MonitorInitScript(
	moduleMonitor string,
	databaseMonitor string,
) string {
	return strings.Join([]string{
		"if [ -z \"$(ldapsearch -Y EXTERNAL -H ldapi:/// -b \"cn=module{0},cn=config\" | grep back_monitor)\" ]",
		fmt.Sprintf("then ldapmodify -Y EXTERNAL -H ldapi:/// -f %s", moduleMonitor),
		fmt.Sprintf("ldapadd -Y EXTERNAL -H ldapi:/// -f %s", databaseMonitor),
		"fi",
	}, ";")
}

func (r *OpenldapCluster) LdapPort() int32 {
	return r.Spec.Ports.Ldap
}

func (r *OpenldapCluster) LdapsPort() int32 {
	return r.Spec.Ports.Ldaps
}

func init() {
	SchemeBuilder.Register(&OpenldapCluster{}, &OpenldapClusterList{})
}
