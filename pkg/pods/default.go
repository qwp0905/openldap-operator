package pods

import (
	"fmt"
	"strconv"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	"github.com/qwp0905/openldap-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func CreateExporterContainer(cluster *openldapv1.OpenldapCluster) corev1.Container {
	return corev1.Container{
		Name:            cluster.ExporterName(),
		Image:           cluster.ExporterImage(),
		ImagePullPolicy: corev1.PullIfNotPresent,
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				"cpu":    resource.MustParse("50m"),
				"memory": resource.MustParse("64Mi"),
			},
			Requests: corev1.ResourceList{
				"cpu":    resource.MustParse("50m"),
				"memory": resource.MustParse("64Mi"),
			},
		},
		Env: []corev1.EnvVar{
			{
				Name: "BIND_PW",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: cluster.Spec.OpenldapConfig.AdminPassword,
				},
			},
			{
				Name:  "BIND_DN",
				Value: cluster.AdminDn(),
			},
			{
				Name:  "METRICS_PORT",
				Value: strconv.Itoa(int(cluster.MetricsPort())),
			},
			{
				Name:  "LDAP_PORT",
				Value: strconv.Itoa(int(cluster.LdapPort())),
			},
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          cluster.MetricsPortName(),
				ContainerPort: cluster.MetricsPort(),
			},
		},
	}
}

func SeedVolumes(cluster *openldapv1.OpenldapCluster) corev1.Volume {
	var volumeSource corev1.VolumeSource

	if cluster.Spec.OpenldapConfig.SeedData.ConfigMap != nil {
		volumeSource = corev1.VolumeSource{
			ConfigMap: cluster.Spec.OpenldapConfig.SeedData.ConfigMap,
		}
	} else if cluster.Spec.OpenldapConfig.SeedData.Secret != nil {
		volumeSource = corev1.VolumeSource{
			Secret: cluster.Spec.OpenldapConfig.SeedData.Secret,
		}
	}

	return corev1.Volume{
		Name:         "ldifs",
		VolumeSource: volumeSource,
	}
}

func DataVolume(cluster *openldapv1.OpenldapCluster, index int) corev1.Volume {
	return corev1.Volume{
		Name: "data",
		VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
				ClaimName: cluster.PodName(index),
			},
		},
	}
}

func DefaultEnvs(cluster *openldapv1.OpenldapCluster) []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		{
			Name: "LDAP_ADMIN_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: cluster.Spec.OpenldapConfig.AdminPassword,
			},
		},
		{Name: "LDAP_CUSTOM_LDIF_DIR", Value: cluster.SeedDataPath()},
		{Name: "LDAP_ENABLE_TLS", Value: utils.ConvertBool(cluster.TlsEnabled())},
		{Name: "LDAP_PORT_NUMBER", Value: strconv.Itoa(int(cluster.LdapPort()))},
		{Name: "LDAP_ROOT", Value: strconv.Itoa(int(cluster.LdapPort()))},
		{Name: "LDAP_PORT_NUMBER", Value: cluster.Spec.OpenldapConfig.Root},
		{Name: "LDAP_CONFIG_ADMIN_ENABLED", Value: "yes"},
		{Name: "LDAP_ADMIN_USERNAME", Value: cluster.Spec.OpenldapConfig.AdminUsername},
		{Name: "LDAP_CONFIG_ADMIN_USERNAME", Value: cluster.Spec.OpenldapConfig.ConfigUsername},
		{
			Name: "MASTER_HOST",
			Value: fmt.Sprintf(
				"ldap://%s.%s.svc.cluster.local:%s",
				cluster.WriteServiceName(),
				cluster.Namespace,
				strconv.Itoa(int(cluster.LdapPort())),
			),
		},
	}

	if cluster.TlsEnabled() {
		envVars = append(
			envVars,
			corev1.EnvVar{
				Name:  "LDAP_LDAPS_PORT_NUMBER",
				Value: strconv.Itoa(int(cluster.LdapsPort())),
			},
			corev1.EnvVar{
				Name: "LDAP_TLS_CERT_FILE",
				Value: fmt.Sprintf(
					"%s/%s",
					cluster.TlsMountPath(),
					cluster.Spec.OpenldapConfig.Tls.CertFile,
				),
			},
			corev1.EnvVar{
				Name: "LDAP_TLS_KEY_FILE",
				Value: fmt.Sprintf(
					"%s/%s",
					cluster.TlsMountPath(),
					cluster.Spec.OpenldapConfig.Tls.KeyFile,
				),
			},
			corev1.EnvVar{
				Name: "LDAP_TLS_CA_FILE",
				Value: fmt.Sprintf(
					"%s/%s",
					cluster.TlsMountPath(),
					cluster.Spec.OpenldapConfig.Tls.CaFile,
				),
			},
		)
	}

	if cluster.Spec.OpenldapConfig.ConfigPassword != nil {
		envVars = append(envVars, corev1.EnvVar{
			Name: "LDAP_CONFIG_ADMIN_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: cluster.Spec.OpenldapConfig.ConfigPassword,
			},
		})
	}

	if cluster.GetTemplate().Env != nil && len(cluster.GetTemplate().Env) > 0 {
		envVars = append(envVars, cluster.GetTemplate().Env...)
	}

	return envVars
}

func ContainerPorts(cluster *openldapv1.OpenldapCluster) []corev1.ContainerPort {
	ports := []corev1.ContainerPort{
		{
			Name:          "ldap",
			Protocol:      "TCP",
			ContainerPort: cluster.LdapPort(),
		},
	}

	if cluster.TlsEnabled() {
		ports = append(
			ports,
			corev1.ContainerPort{
				Name:          "ldaps",
				Protocol:      "TCP",
				ContainerPort: cluster.LdapsPort(),
			},
		)
	}

	return ports
}

func DataVolumeMounts(cluster *openldapv1.OpenldapCluster) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      "data",
		MountPath: "/bitnami/openldap",
	}
}

func SeedVolumeMount(cluster *openldapv1.OpenldapCluster) corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      "ldifs",
		MountPath: cluster.SeedDataPath(),
	}
}
