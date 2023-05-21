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
	template := cluster.GetTemplate()

	if cluster.Spec.OpenldapConfig.SeedData != nil {
		initVolumeMounts = append(volumeMounts, pods.SeedVolumeMount(cluster))
		volumes = append(volumes, pods.SeedVolumes(cluster))
	}

	initContainers := []corev1.Container{{
		Name:            cluster.InitContainerName(),
		Image:           template.Image,
		ImagePullPolicy: template.ImagePullPolicy,
		Command:         []string{"/opt/bitnami/scripts/openldap/setup.sh"},
		Resources:       template.Resources,
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
			Image:           template.Image,
			ImagePullPolicy: template.ImagePullPolicy,
			Resources:       template.Resources,
			ReadinessProbe:  cluster.ReadinessProbe(),
			LivenessProbe:   cluster.LivenessProbe(),
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
		Affinity:                      template.Affinity,
		NodeSelector:                  template.NodeSelector,
		TerminationGracePeriodSeconds: &gracefulPeriod,
		ImagePullSecrets:              cluster.Spec.ImagePullSecrets,
	}

	if template.Tolerations != nil {
		podSpec.Tolerations = template.Tolerations
	}

	if template.PriorityClassName != "" {
		podSpec.PriorityClassName = template.PriorityClassName
	}

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cluster.Name,
			Namespace:   cluster.Namespace,
			Labels:      cluster.DefaultLabels(),
			Annotations: cluster.GetAnnotations(),
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
					Labels:      cluster.GetSlaveLabels(),
					Annotations: cluster.GetAnnotations(),
				},
				Spec: *podSpec,
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "data",
						Labels:      cluster.SelectorLabels(),
						Annotations: cluster.GetAnnotations(),
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
