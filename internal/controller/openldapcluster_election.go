package controller

import (
	"context"
	"fmt"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	"github.com/qwp0905/openldap-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *OpenldapClusterReconciler) election(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (bool, error) {
	logger := log.FromContext(ctx)

	if cluster.Status.Master == "" {
		cluster.UpdateMaster(0)
		cluster.UpdatePhase(openldapv1.PhaseCreating)

		if err := r.Status().Update(ctx, cluster); err != nil {
			logger.Error(err, "Error on Updating Cluster Status....")
			return false, err
		}

		logger.Info(fmt.Sprintf("Master Pod Set 0"))
		return true, nil
	}

	masterPod, err := r.getMasterPod(ctx, cluster)
	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error on getting master pod....")
			return false, err
		}

		if cluster.Status.Phase == openldapv1.PhaseCreating {
			logger.Info("Waiting for first pod...")
			return true, nil
		}

		nextIndex, err := r.getAlivePodIndex(ctx, cluster)
		if err != nil {
			logger.Error(err, "Error on get pods...")
			return false, err
		}

		cluster.UpdateMaster(nextIndex)
		cluster.UpdatePhase(openldapv1.PhasePending)

		if err := r.Status().Update(ctx, cluster); err != nil {
			logger.Error(err, "Error on Updating Cluster Status....")
			return false, err
		}

		logger.Info(fmt.Sprintf("Master Pod Set %s", nextIndex))
		return true, nil
	}

}

func (r *OpenldapClusterReconciler) getAlivePodIndex(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (int, error) {
	for i := 0; i < cluster.GetReplicas(); i++ {
		pod, err := r.getPod(ctx, cluster, i)
		if err != nil {
			if !errors.IsNotFound(err) {
				return 0, err
			}
		}

		if utils.IsPodAlive(*pod) {
			return i, nil
		}
	}

	return 0, fmt.Errorf("No Pod Alive")
}
