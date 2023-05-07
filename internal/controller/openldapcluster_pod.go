package controller

import (
	"context"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
