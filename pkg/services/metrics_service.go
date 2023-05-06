package services

import (
	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateMetricsService(cluster *openldapv1.OpenldapCluster) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.MetricsServiceName(),
			Namespace: cluster.Namespace,
			Labels:    cluster.SelectorLabels(),
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:     cluster.MetricsPortName(),
					Port:     cluster.MetricsPort(),
					Protocol: corev1.ProtocolTCP,
				},
			},
			Selector: cluster.SelectorLabels(),
		},
	}
}
