package controller

import (
	"context"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *OpenldapClusterReconciler) setPod(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
	index int,
) (bool, error) {
	logger := log.FromContext(ctx)

	existsPod, err := r.getPod(ctx, cluster, index)
	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error on Get Pod...")
		}

		newPod := r.createPod(ctx, cluster, index)
	}
}

func (r *OpenldapClusterReconciler) getPod(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
	index int,
) (*corev1.Pod, error) {
	pod := &corev1.Pod{}

	err := r.Get(
		ctx,
		types.NamespacedName{Name: cluster.PodName(index), Namespace: cluster.Namespace},
		pod,
	)
	if err != nil {
		return nil, err
	}

	return pod, nil
}

func (r *OpenldapClusterReconciler) createPod(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
	index int,
) *corev1.Pod {

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.PodName(index),
			Namespace: cluster.Namespace,
			Labels:    cluster.SelectorLabels(),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:            cluster.Name,
					Image:           cluster.Spec.Image,
					ImagePullPolicy: cluster.Spec.ImagePullPolicy,
					Resources:       *cluster.Spec.Resources,
					ReadinessProbe:  cluster.ContainerProbe(),
					LivenessProbe:   cluster.ContainerProbe(),
					Ports: []corev1.ContainerPort{
						{
							Name:          "ldap",
							Protocol:      "TCP",
							ContainerPort: cluster.Spec.Ports.Ldap,
						},
						{
							Name:          "ldaps",
							Protocol:      "TCP",
							ContainerPort: cluster.Spec.Ports.Ldaps,
						},
					},
					EnvFrom: []corev1.EnvFromSource{
						{
							ConfigMapRef: &corev1.ConfigMapEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: cluster.Name,
								},
							},
						},
					},
					Env: []corev1.EnvVar{
						{
							Name: "POD_NAME",
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{
									FieldPath: "metadata.name",
								},
							},
						},
						{
							Name: "LDAP_ADMIN_USERNAME",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: cluster.Spec.OpenldapConfig.AdminUsername,
							},
						},
						{
							Name: "LDAP_ADMIN_PASSWORD",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: cluster.Spec.OpenldapConfig.AdminPassword,
							},
						},
						{
							Name: "LDAP_CONFIG_USERNAME",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: cluster.Spec.OpenldapConfig.ConfigUsername,
							},
						},
						{
							Name: "LDAP_CONFIG_PASSWORD",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: cluster.Spec.OpenldapConfig.ConfigPassword,
							},
						},
					},
				},
			},
		},
	}
}
