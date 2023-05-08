package pods

import (
	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateSlavePod(cluster *openldapv1.OpenldapCluster, index int) *corev1.Pod {
	envVars := append(
		defaultEnvs(cluster),
		corev1.EnvVar{
			Name:  "ROLE",
			Value: "slave",
		},
	)

	containers := []corev1.Container{
		{
			Name:            cluster.Name,
			Image:           cluster.Spec.Image,
			ImagePullPolicy: cluster.Spec.ImagePullPolicy,
			Resources:       *cluster.Spec.Resources,
			ReadinessProbe:  cluster.ContainerProbe(),
			LivenessProbe:   cluster.ContainerProbe(),
			Env:             envVars,
			EnvFrom:         []corev1.EnvFromSource{configEnvFrom(cluster)},
			Ports:           containerPorts(cluster),
			VolumeMounts:    []corev1.VolumeMount{dataVolumeMounts(cluster)},
		},
	}

	if cluster.MonitorEnabled() {
		containers = append(containers, createExporterContainer(cluster))
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.PodName(index),
			Namespace: cluster.Namespace,
			Labels:    cluster.SlaveSelectorLabels(),
		},
		Spec: corev1.PodSpec{
			Containers: containers,
			Volumes:    []corev1.Volume{dataVolume(cluster, index)},
		},
	}
}
