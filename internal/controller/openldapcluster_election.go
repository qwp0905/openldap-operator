package controller

import (
	"context"
	"fmt"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	"github.com/qwp0905/openldap-operator/pkg/jobs"
	"github.com/qwp0905/openldap-operator/pkg/utils"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *OpenldapClusterReconciler) election(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (bool, error) {
	logger := log.FromContext(ctx)

	if cluster.GetDesiredMaster() == "" {
		cluster.UpdateDesiredMaster(0)
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

		if err = r.newElection(ctx, cluster); err != nil {
			return false, err
		}

		return true, nil
	}

	if !utils.IsPodAlive(*masterPod) {
		if err = r.newElection(ctx, cluster); err != nil {
			return false, err
		}

		if err = r.Delete(ctx, masterPod); err != nil {
			logger.Error(err, "Error on deleting failed pod...")
			return false, err
		}

		return true, nil
	}

	if cluster.IsMasterSame() {
		logger.Info("Nothing to update on pod")
		return false, nil
	}

	existsJob, err := r.getJob(ctx, cluster)
	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error on Get job....")
			return false, err
		}

		newJob := jobs.CreateSlaveToMasterJob(cluster)

		if err = ctrl.SetControllerReference(cluster, newJob, r.Scheme); err != nil {
			logger.Error(err, "Error on registering job...")
			return false, err
		}

		if err = r.Create(ctx, newJob); err != nil {
			logger.Error(err, "Error on Creating Job...")
			return false, err
		}

		logger.Info("Master Job Created")
		return true, nil
	}

	if !utils.JobHasOneCompletion(*existsJob) {
		logger.Info("Waiting for job complete")
		return true, nil
	}

	cluster.UpdateCurrentMaster()

	if err = r.Status().Update(ctx, cluster); err != nil {
		logger.Error(err, "Error on Updating Status Current master....")
		return false, err
	}

	if err = r.Delete(ctx, existsJob); err != nil {
		logger.Error(err, "Error on Deleting job...")
		return false, err
	}

	logger.Info("Master Pod Updated")
	return true, nil
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

func (r *OpenldapClusterReconciler) newElection(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) error {
	logger := log.FromContext(ctx)

	nextIndex, err := r.getAlivePodIndex(ctx, cluster)
	if err != nil {
		logger.Error(err, "Error on get pods...")
		return err
	}

	cluster.UpdateDesiredMaster(nextIndex)
	cluster.UpdatePhase(openldapv1.PhasePending)

	if err := r.Status().Update(ctx, cluster); err != nil {
		logger.Error(err, "Error on Updating Cluster Status....")
		return err
	}

	return nil
}

func (r *OpenldapClusterReconciler) getJob(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (*batchv1.Job, error) {
	job := &batchv1.Job{}

	if err := r.Get(
		ctx,
		types.NamespacedName{Name: cluster.GetDesiredMaster(), Namespace: cluster.Namespace},
		job,
	); err != nil {
		return nil, err
	}

	return job, nil
}

func (r *OpenldapClusterReconciler) getMasterPod(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (*corev1.Pod, error) {
	pod := &corev1.Pod{}

	err := r.Get(
		ctx,
		types.NamespacedName{Name: cluster.Status.DesiredMaster, Namespace: cluster.Name},
		pod,
	)

	return pod, err
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
