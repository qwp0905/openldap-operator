package statefulsets

import (
	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	"github.com/qwp0905/openldap-operator/pkg/pods"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func CreateStatefulset(cluster *openldapv1.OpenldapCluster) *appsv1.StatefulSet {
	volumeMounts := []corev1.VolumeMount{pods.DataVolumeMounts(cluster)}
	volumes := []corev1.Volume{}
	initVolumeMounts := []corev1.VolumeMount{pods.DataVolumeMounts(cluster)}
	rootUser := int64(0)
	gracefulPeriod := int64(10)

	if cluster.Spec.OpenldapConfig.SeedData != nil {
		initVolumeMounts = append(volumeMounts, pods.SeedVolumeMount(cluster))
		volumes = append(volumes, pods.SeedVolumes(cluster))
	}

	initContainers := []corev1.Container{{
		Name:            cluster.InitContainerName(),
		Image:           cluster.Spec.Image,
		ImagePullPolicy: cluster.Spec.ImagePullPolicy,
		Command:         []string{"/opt/bitnami/scripts/openldap/setup.sh"},
		Resources:       *cluster.Spec.Resources,
		Env:             pods.DefaultEnvs(cluster),
		EnvFrom:         []corev1.EnvFromSource{pods.ConfigEnvFrom(cluster)},
		VolumeMounts:    initVolumeMounts,
		SecurityContext: &corev1.SecurityContext{
			RunAsUser: &rootUser,
		},
	}}

	containers := []corev1.Container{
		{
			Name:            cluster.Name,
			Image:           cluster.Spec.Image,
			ImagePullPolicy: cluster.Spec.ImagePullPolicy,
			Resources:       *cluster.Spec.Resources,
			ReadinessProbe:  cluster.ContainerProbe(),
			LivenessProbe:   cluster.ContainerProbe(),
			Env:             pods.DefaultEnvs(cluster),
			EnvFrom:         []corev1.EnvFromSource{pods.ConfigEnvFrom(cluster)},
			Ports:           pods.ContainerPorts(cluster),
			VolumeMounts:    volumeMounts,
		},
	}

	if cluster.MonitorEnabled() {
		containers = append(containers, pods.CreateExporterContainer(cluster))
	}

	podSpec := &corev1.PodSpec{
		InitContainers:                initContainers,
		Containers:                    containers,
		Volumes:                       volumes,
		Affinity:                      cluster.Spec.Affinity,
		NodeSelector:                  cluster.Spec.NodeSelector,
		TerminationGracePeriodSeconds: &gracefulPeriod,
		ImagePullSecrets:              cluster.Spec.ImagePullSecrets,
	}

	if cluster.Spec.Tolerations != nil {
		podSpec.Tolerations = cluster.Spec.Tolerations
	}

	if cluster.Spec.PriorityClassName != "" {
		podSpec.PriorityClassName = cluster.Spec.PriorityClassName
	}

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
			Labels:    cluster.DefaultLabels(),
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: cluster.SelectorLabels(),
			},
			Replicas: &cluster.Spec.Replicas,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 1,
					},
				},
			},
			PersistentVolumeClaimRetentionPolicy: &appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy{
				WhenDeleted: appsv1.RetainPersistentVolumeClaimRetentionPolicyType,
				WhenScaled:  appsv1.DeletePersistentVolumeClaimRetentionPolicyType,
			},
			MinReadySeconds: 0,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: cluster.SlaveSelectorLabels(),
				},
				Spec: *podSpec,
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "data",
						Labels: cluster.SelectorLabels(),
					},
					Spec: cluster.Spec.Storage.VolumeClaimTemplate,
				},
			},
		},
	}
}

func GetContainer(st *appsv1.StatefulSet, name string) *corev1.Container {
	for _, c := range st.Spec.Template.Spec.Containers {
		if c.Name == name {
			return &c
		}
	}

	return nil
}
