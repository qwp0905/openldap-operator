package controller

import (
	"context"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *OpenldapClusterReconciler) setCluster(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (bool, error) {
	logger := log.FromContext(ctx)
	masterPod, err := r.getMasterPod(ctx, cluster)

	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error on Getting Master Pod...")
			return false, err
		}
	}

	if masterPod.Status.Phase == corev1.PodFailed {
	}
}

func (r *OpenldapClusterReconciler) getMasterPod(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (*corev1.Pod, error) {
	pod := &corev1.Pod{}

	err := r.Get(
		ctx,
		types.NamespacedName{Name: cluster.Status.Master, Namespace: cluster.Name},
		pod,
	)

	return pod, err
}
