package controller

import (
	"context"
	"fmt"
	"reflect"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	configMountPath      = "/etc/ldap/slapd.d"
	configVolume         = "openldap-config"
	dataMountPath        = "/var/lib/ldap"
	dataVolume           = "openldap-data"
	tlsMountPath         = "/container/service/slapd/assets/certs"
	tlsVolume            = "openldap-tls"
	seedDataMountPath    = "/var/init/seed"
	seedDataVolume       = "openldap-seed"
	monitorDataMountPath = "/var/init/monitor"
	monitorDataVolume    = "openldap-monitor"
	monitorInitScript    = "init.sh"
	exporterVolume       = "exporter-config"
	exporterMountPath    = "/config/"

	exporterImage = "qwp1216/openldap-exporter:0.0.2"

	metricsPort     = 9142
	metricsPortName = "metrics"
	metricsPath     = "/metrics"
)

func (r *OpenldapClusterReconciler) setStatefulset(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (bool, error) {
	logger := log.FromContext(ctx)
	existsStatefulset, err := r.getStatefulset(ctx, cluster)

	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error on Get Statefulset...")
			return false, err
		}
		newStatefulset := r.statefulset(cluster)
		err = ctrl.SetControllerReference(cluster, newStatefulset, r.Scheme)
		if err != nil {
			logger.Error(err, "Error on Registering Statefulset...")
			return false, err
		}

		if err := r.Create(ctx, newStatefulset); err != nil {
			logger.Error(err, "Error on Creating Statefulset...")
			return false, err
		}

		logger.Info("Statefulset Created")
		return true, nil
	}

	updatedStatefulset := r.statefulset(cluster)

	if r.compareStatefulset(cluster, existsStatefulset, updatedStatefulset) {
		logger.Info("Nothing to Update On Statefulset")
		return false, nil
	}

	if err = r.Update(ctx, updatedStatefulset); err != nil {
		logger.Error(err, "Error on Updating Statefulset...")
		return false, err
	}

	logger.Info("Statefulset Updated")
	return true, nil
}

func (r *OpenldapClusterReconciler) getStatefulset(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (*appsv1.StatefulSet, error) {
	statefulset := &appsv1.StatefulSet{}

	err := r.Client.Get(
		ctx,
		types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace},
		statefulset,
	)

	if err != nil {
		return nil, err
	}

	return statefulset, nil
}

func (r *OpenldapClusterReconciler) initContainers(
	cluster *openldapv1.OpenldapCluster,
) []corev1.Container {

	initVolumeMounts := append(
		r.defaultVolumeMounts(cluster),
		corev1.VolumeMount{
			Name:      seedDataVolume,
			MountPath: seedDataMountPath,
		},
		corev1.VolumeMount{
			Name:      monitorDataVolume,
			MountPath: monitorDataMountPath,
		},
	)

	initEnvVars := append(
		r.defaultEnvVars(cluster),
		corev1.EnvVar{
			Name:  "KEEP_EXISTING_CONFIG",
			Value: "false",
		},
	)

	return []corev1.Container{
		{
			Name:            cluster.InitContainerName(),
			Image:           cluster.Spec.Image,
			ImagePullPolicy: cluster.Spec.ImagePullPolicy,
			Resources:       *cluster.Spec.Resources,
			Args: []string{
				fmt.Sprintf("sh %s/bootstrap", monitorDataMountPath),
			},
			EnvFrom: []corev1.EnvFromSource{
				{
					ConfigMapRef: &corev1.ConfigMapEnvSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cluster.ConfigMapName(),
						},
					},
				},
			},
			Env:          initEnvVars,
			VolumeMounts: initVolumeMounts,
		},
	}
}

func (r *OpenldapClusterReconciler) defaultVolumeMounts(
	cluster *openldapv1.OpenldapCluster,
) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		{Name: configVolume, MountPath: configMountPath},
		{Name: dataVolume, MountPath: dataMountPath},
	}

	if cluster.Spec.OpenldapConfig.Tls.Enabled {
		volumeMounts = append(
			volumeMounts,
			corev1.VolumeMount{
				Name:      tlsVolume,
				MountPath: tlsMountPath,
			},
		)
	}

	return volumeMounts
}

func (r *OpenldapClusterReconciler) defaultEnvVars(
	cluster *openldapv1.OpenldapCluster,
) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: "LDAP_ADMIN_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: cluster.Spec.OpenldapConfig.AdminPassword,
			},
		},
	}
}

func (r *OpenldapClusterReconciler) openldapContainer(
	cluster *openldapv1.OpenldapCluster,
) corev1.Container {

	containerEnvs := append(
		r.defaultEnvVars(cluster),
		corev1.EnvVar{
			Name:  "KEEP_EXISTING_CONFIG",
			Value: "true",
		},
	)

	volumeMounts := r.defaultVolumeMounts(cluster)

	return corev1.Container{
		Name:            cluster.Name,
		Image:           cluster.Spec.Image,
		ImagePullPolicy: cluster.Spec.ImagePullPolicy,
		Resources:       *cluster.Spec.Resources,
		Args:            []string{"-l", cluster.Spec.OpenldapConfig.LogLevel},
		Ports: []corev1.ContainerPort{
			{Name: "ldap", Protocol: "TCP", ContainerPort: 389},
			{Name: "ldaps", Protocol: "TCP", ContainerPort: 636},
		},
		ReadinessProbe: cluster.ContainerProbe(),
		LivenessProbe:  cluster.ContainerProbe(),
		Env:            containerEnvs,
		EnvFrom: []corev1.EnvFromSource{
			{
				ConfigMapRef: &corev1.ConfigMapEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cluster.ConfigMapName(),
					},
				},
			},
		},
		VolumeMounts: volumeMounts,
	}
}

func (r *OpenldapClusterReconciler) statefulset(cluster *openldapv1.OpenldapCluster) *appsv1.StatefulSet {

	containers := []corev1.Container{
		r.openldapContainer(cluster),
		r.exporterContainer(cluster),
	}

	maxUnavailable := intstr.FromInt(1)

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
			Labels:    cluster.SelectorLabels(),
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName:     cluster.Name,
			Replicas:        &cluster.Spec.Replicas,
			MinReadySeconds: 60,
			Selector: &metav1.LabelSelector{
				MatchLabels: cluster.SelectorLabels(),
			},
			PersistentVolumeClaimRetentionPolicy: &appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy{
				WhenDeleted: appsv1.RetainPersistentVolumeClaimRetentionPolicyType,
				WhenScaled:  appsv1.DeletePersistentVolumeClaimRetentionPolicyType,
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
					MaxUnavailable: &maxUnavailable,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: cluster.SelectorLabels(),
				},
				Spec: corev1.PodSpec{
					Containers:     containers,
					InitContainers: r.initContainers(cluster),
					Volumes:        r.statefulsetVolumes(cluster),
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   dataVolume,
						Labels: cluster.SelectorLabels(),
					},
					Spec: cluster.Spec.Storage.VolumeClaimTemplate,
				},
			},
		},
	}
}

func (r *OpenldapClusterReconciler) statefulsetVolumes(
	cluster *openldapv1.OpenldapCluster,
) []corev1.Volume {

	volumes := []corev1.Volume{
		{
			Name: configVolume,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: monitorDataVolume,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cluster.MonitorConfigMapName(),
					},
				},
			},
		},
		{
			Name: exporterVolume,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: cluster.ExporterName(),
					},
				},
			},
		},
	}

	if cluster.Spec.OpenldapConfig.Tls.Enabled {
		defaultMode := int32(420)

		volumes = append(
			volumes,
			corev1.Volume{
				Name: tlsVolume,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  cluster.Spec.OpenldapConfig.Tls.SecretName,
						DefaultMode: &defaultMode,
					},
				},
			},
		)
	}

	if cluster.Spec.OpenldapConfig.SeedData != nil &&
		(cluster.Spec.OpenldapConfig.SeedData.ConfigMap != nil ||
			cluster.Spec.OpenldapConfig.SeedData.Secret != nil) {
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

		volumes = append(
			volumes,
			corev1.Volume{
				Name:         seedDataVolume,
				VolumeSource: volumeSource,
			},
		)
	}

	return volumes
}

func (r *OpenldapClusterReconciler) exporterContainer(
	cluster *openldapv1.OpenldapCluster,
) corev1.Container {

	return corev1.Container{
		Name:            cluster.ExporterName(),
		Image:           exporterImage,
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
				Name:  "BIND_DN",
				Value: cluster.BindDn(),
			},
			{
				Name: "BIND_PW",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: cluster.Spec.OpenldapConfig.AdminPassword,
				},
			},
		},
		Ports: []corev1.ContainerPort{
			{
				Name:          metricsPortName,
				Protocol:      "TCP",
				ContainerPort: metricsPort,
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      exporterVolume,
				MountPath: exporterMountPath,
			},
		},
	}
}

func (r *OpenldapClusterReconciler) compareStatefulset(
	cluster *openldapv1.OpenldapCluster,
	exists *appsv1.StatefulSet,
	new *appsv1.StatefulSet,
) bool {

	for label, value := range new.Labels {
		if exists.Labels[label] != value {
			return false
		}
	}

	for annotation, value := range new.Annotations {
		if exists.Annotations[annotation] != value {
			return false
		}
	}

	// var exOpenldapCon corev1.Container
	// var neOpenldapCon corev1.Container

	// var exExporter corev1.Container
	// var neExporter corev1.Container

	// existsContainers := exists.Spec.Template.Spec.Containers
	// newContainers := new.Spec.Template.Spec.Containers

	// for _, con := range existsContainers {
	// 	if con.Name == cluster.Name {
	// 		exOpenldapCon = con
	// 	}

	// 	if con.Name == cluster.ExporterName() {
	// 		exExporter = con
	// 	}
	// }

	// for _, con := range newContainers {
	// 	if con.Name == cluster.Name {
	// 		neOpenldapCon = con
	// 	}

	// 	if con.Name == cluster.ExporterName() {
	// 		neExporter = con
	// 	}
	// }

	return reflect.DeepEqual(exists.Spec.VolumeClaimTemplates, new.Spec.VolumeClaimTemplates) &&
		exists.Spec.Replicas == new.Spec.Replicas
}
