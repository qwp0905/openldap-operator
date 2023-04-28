package controller

import (
	"context"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *OpenldapClusterReconciler) setStatefulset(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) error {
	statefulset, err := r.getStatefulset(ctx, cluster)

	if errors.IsNotFound(err) {
		err = r.createStatefulset(ctx, cluster)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
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

func (r *OpenldapClusterReconciler) createStatefulset(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) error {
	return r.Create(ctx, &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":     cluster.Name,
				"app.kubernetes.io/instance": "openldap",
			},
		},
	})
}
