package pods

import (
	"strconv"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
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

func ConfigEnvFrom(cluster *openldapv1.OpenldapCluster) corev1.EnvFromSource {
	return corev1.EnvFromSource{
		ConfigMapRef: &corev1.ConfigMapEnvSource{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: cluster.ConfigMapName(),
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
		{
			Name: "LDAP_CONFIG_ADMIN_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: cluster.Spec.OpenldapConfig.ConfigPassword,
			},
		},
	}

	if cluster.Spec.OpenldapConfig.Env == nil || len(cluster.Spec.OpenldapConfig.Env) == 0 {
		return envVars
	}

	return append(envVars, cluster.Spec.OpenldapConfig.Env...)
}

func ContainerPorts(cluster *openldapv1.OpenldapCluster) []corev1.ContainerPort {
	ports := []corev1.ContainerPort{
		{
			Name:          "ldap",
			Protocol:      "TCP",
			ContainerPort: cluster.Spec.Ports.Ldap,
		},
	}

	if !cluster.TlsEnabled() {
		return ports
	}

	return append(
		ports,
		corev1.ContainerPort{
			Name:          "ldaps",
			Protocol:      "TCP",
			ContainerPort: cluster.Spec.Ports.Ldaps,
		},
	)
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
