package controller

import (
	"context"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	"github.com/qwp0905/openldap-operator/pkg/persistentvolumes"
	"github.com/qwp0905/openldap-operator/pkg/pods"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *OpenldapClusterReconciler) setCluster(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (bool, error) {
	logger := log.FromContext(ctx)

	resources, err := r.getManagedResources(ctx, cluster)
	if err != nil {
		logger.Error(err, "Error on getting managed resources")
		return false, err
	}

	count, err := r.countPods(ctx, cluster)
	if err != nil {
		logger.Error(err, "Error on Getting Pods....")
		return false, nil
	}

	if !cluster.IsMasterUpdated() && count == 0 {
		pvc := persistentvolumes.CreatePersistentVolumeClaim(cluster, 0)
		master := pods.CreateMasterPod(cluster, 0)

		cluster.UpdatePhase(openldapv1.PhaseCreating)
		cluster.UpdateMaster(0)
		cluster.Status.UpdatedReplicas = 0

		if err = r.Status().Update(ctx, cluster); err != nil {
			logger.Error(err, "Error on Update Status Cluster....")
			return false, err
		}

		if err = ctrl.SetControllerReference(cluster, master, r.Scheme); err != nil {
			logger.Error(err, "Error on Registering Master pod....")
			return false, err
		}

		if err = r.Create(ctx, pvc); err != nil {
			logger.Error(err, "Error on Creating Master Pvc")
			return false, err
		}

		if err = r.Create(ctx, master); err != nil {
			logger.Error(err, "Error on Creating Master")
			return false, err
		}

		logger.Info("Master pod created")
		return true, nil
	}

	master, err := r.getMasterPod(ctx, cluster)
	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error on Getting Master Pod....")
			return false, err
		}

		cluster.UpdatePhase(openldapv1.PhaseUpdating)

		if err = r.Status().Update(ctx, cluster); err != nil {
			logger.Error(err, "Error on Update Status Cluster....")
			return false, err
		}

		return r.election(ctx, cluster)
	}

	switch master.Status.Phase {
	case corev1.PodFailed:
	case corev1.PodUnknown:
		if err = r.Delete(ctx, master); err != nil {
			logger.Error(err, "Error on Deleting Master pod...")
			return false, err
		}

		logger.Info("Master pod Deleted")
		return true, nil

	case corev1.PodPending:
		logger.Info("Still waiting for master pod ready...")
		return true, nil
	}
}

func (r *OpenldapClusterReconciler) countPods(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (int, error) {
	count := 0
	for i := 0; i < int(cluster.Spec.Replicas); i++ {
		if _, err := r.getPod(ctx, cluster, i); err != nil {
			return 0, err
		}
		count++
	}

	return count, nil
}

func (r *OpenldapClusterReconciler) getPod(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
	index int,
) (*corev1.Pod, error) {
	pod := &corev1.Pod{}

	err := r.Get(
		ctx,
		types.NamespacedName{Name: cluster.PodName(index), Namespace: cluster.Name},
		pod,
	)

	return pod, err
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
