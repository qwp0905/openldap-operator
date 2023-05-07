package rbac

import (
	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateRoleBinding(cluster *openldapv1.OpenldapCluster) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
			Labels:    cluster.JobLabels(),
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     "Role",
			Name:     cluster.Name,
		},
		Subjects: []rbacv1.Subject{
			{
				Name:      cluster.Name,
				Namespace: cluster.Namespace,
				Kind:      rbacv1.ServiceAccountKind,
			},
		},
	}
}
