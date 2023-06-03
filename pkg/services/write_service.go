package services

import (
	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func CreateWriteService(cluster *openldapv1.OpenldapCluster) *corev1.Service {
	ports := []corev1.ServicePort{
		{
			Name:     "ldap",
			Port:     cluster.LdapPort(),
			Protocol: corev1.ProtocolTCP,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: cluster.LdapPort(),
			},
		},
	}

	if cluster.TlsEnabled() {
		ports = append(ports, corev1.ServicePort{
			Name:     "ldaps",
			Port:     cluster.LdapsPort(),
			Protocol: corev1.ProtocolTCP,
			TargetPort: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: cluster.LdapsPort(),
			},
		})
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.WriteServiceName(),
			Namespace: cluster.Namespace,
			Labels:    cluster.GetMasterLabels(),
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Ports:    ports,
			Selector: cluster.MasterSelectorLabels(),
		},
	}
}
