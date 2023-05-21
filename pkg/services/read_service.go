package services

import (
	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateReadService(cluster *openldapv1.OpenldapCluster) *corev1.Service {
	ports := []corev1.ServicePort{
		{
			Name:     "ldap",
			Port:     cluster.LdapPort(),
			Protocol: corev1.ProtocolTCP,
		},
	}

	if cluster.TlsEnabled() {
		ports = append(ports, corev1.ServicePort{
			Name:     "ldaps",
			Port:     cluster.LdapsPort(),
			Protocol: corev1.ProtocolTCP,
		})
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        cluster.Name,
			Namespace:   cluster.Namespace,
			Labels:      cluster.DefaultLabels(),
			Annotations: cluster.GetAnnotations(),
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Ports:    ports,
			Selector: cluster.SelectorLabels(),
		},
	}
}
