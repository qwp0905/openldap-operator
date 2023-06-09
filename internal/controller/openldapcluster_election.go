package controller

import (
	"context"
	"fmt"
	"strconv"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	"github.com/qwp0905/openldap-operator/pkg/jobs"
	"github.com/qwp0905/openldap-operator/pkg/utils"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *OpenldapClusterReconciler) election(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (int, error) {
	logger := log.FromContext(ctx)

	if cluster.IsConditionsEmpty() {
		cluster.SetInitCondition()

		if err := r.Status().Update(ctx, cluster); err != nil {
			logger.Error(err, "Error on Updating Cluster Conditions....")
			return 0, err
		}

		logger.Info("Cluster Condition Initialized")
		return 2, nil
	}

	if cluster.GetDesiredMaster() == "" {
		cluster.UpdateDesiredMaster(0)

		if err := r.Status().Update(ctx, cluster); err != nil {
			logger.Error(err, "Error on Updating Cluster Desired Master....")
			return 0, err
		}
		r.Recorder.Eventf(
			cluster,
			"Normal",
			"DesiredMasterCreated",
			"First Desired Master to be set %s",
			cluster.GetDesiredMaster(),
		)
		logger.Info("Master Pod Set 0")
		return 2, nil
	}

	masterPod, err := r.getMasterPod(ctx, cluster)
	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error on getting master pod....")
			return 0, err
		}

		if cluster.IsInitialized() {
			logger.Info("Waiting for first pod...")
			return 5, nil
		}

		logger.Info("Election Triggered because of Pod Not Found....")
		if err = r.electMaster(ctx, cluster); err != nil {
			return 0, err
		}

		return 2, nil
	}

	if !utils.IsPodAlive(*masterPod) || !utils.IsPodReady(*masterPod) {
		if cluster.IsInitialized() {
			logger.Info("Waiting for pod ready")
			return 10, nil
		}

		logger.Info("Election Triggered because of Pod Unhealthy....")
		if err = r.electMaster(ctx, cluster); err != nil {
			return 0, err
		}

		if cluster.GetReplicas() == 1 {
			return 2, nil
		}

		if err = r.Delete(ctx, masterPod); err != nil {
			logger.Error(err, "Error on deleting failed pod...")
			return 0, err
		}

		logger.Info("Deleting Exists pod...")
		return 2, nil
	}

	if utils.IsPodRestart(*masterPod) {
		logger.Info("Election Triggered because of Pod Restarted....")
		if err = r.electMaster(ctx, cluster); err != nil {
			return 0, err
		}

		if err = r.Delete(ctx, masterPod); err != nil {
			logger.Error(err, "Error on deleting restarted pod...")
			return 0, err
		}

		logger.Info("Deleting Exists pod...")
		return 2, nil
	}

	if !cluster.IsElected() {
		existsJob, err := r.getJob(ctx, cluster)
		if err != nil {
			if !errors.IsNotFound(err) {
				logger.Error(err, "Error on Get job....")
				return 0, err
			}

			newJob := jobs.CreateSlaveToMasterJob(cluster)

			if err = r.registerObject(cluster, newJob); err != nil {
				logger.Error(err, "Error on registering job...")
				return 0, err
			}

			if err = r.Create(ctx, newJob); err != nil {
				logger.Error(err, "Error on Creating Job...")
				return 0, err
			}

			cluster.SetConditionElected(true)

			if err = r.Status().Update(ctx, cluster); err != nil {
				logger.Error(err, "Error on Updating Status Elected...")
				return 0, err
			}

			logger.Info("Master Job Created")
			return 10, nil
		}

		if !utils.JobHasOneCompletion(*existsJob) {
			logger.Info("Waiting for job complete")
			return 5, nil
		}
	}

	if !cluster.IsMasterSame() {
		origin := masterPod.DeepCopy()
		masterPod.SetLabels(
			utils.MergeMap(
				masterPod.GetLabels(),
				cluster.MasterSelectorLabels(),
			),
		)
		if err = r.Patch(ctx, masterPod, client.MergeFrom(origin)); err != nil {
			logger.Error(err, "Error on Updating labels to master...")
			return 0, err
		}

		cluster.UpdateCurrentMaster()
		if err = r.Status().Update(ctx, cluster); err != nil {
			logger.Error(err, "Error on Updating Status Current master....")
			return 0, err
		}

		r.Recorder.Eventf(
			cluster,
			"Normal",
			"ElectionCompleted",
			"Current master updated to %s",
			cluster.GetCurrentMaster(),
		)
		logger.Info("Master Pod Updated")
		return 10, nil
	}

	if !cluster.IsReady() {
		cluster.DeleteInitializedCondition()
		cluster.SetConditionReady(true)

		r.Recorder.Eventf(
			cluster,
			"Normal",
			"ClusterReady",
			"Cluster %s has ready status",
			cluster.Name,
		)
		if err := r.Status().Update(ctx, cluster); err != nil {
			logger.Error(err, "Error on Updating Cluster status Running...")
			return 0, err
		}
	}

	logger.Info("Everything ok")
	return 10, nil
}

func (r *OpenldapClusterReconciler) getAlivePodIndex(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (int, error) {
	for i := 0; i < cluster.GetReplicas(); i++ {
		pod, err := r.getPod(ctx, cluster, i)
		if err != nil {
			continue
		}

		if utils.IsPodAlive(*pod) && utils.IsPodReady(*pod) {
			return i, nil
		}
	}

	return 0, fmt.Errorf("no Pod Alive")
}

func (r *OpenldapClusterReconciler) electMaster(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) error {
	logger := log.FromContext(ctx)
	r.Recorder.Eventf(
		cluster,
		"Warning",
		"MasterUnhealthy",
		"Election triggered because of current master pod %s unhealthy",
		cluster.GetCurrentMaster(),
	)

	cluster.DeleteCurrentMaster()
	cluster.SetConditionElected(false)
	if err := r.Status().Update(ctx, cluster); err != nil {
		logger.Error(err, "Error on Updating Cluster Condition Elected....")
		return err
	}

	nextIndex, err := r.getAlivePodIndex(ctx, cluster)
	if err != nil {
		cluster.SetConditionReady(false)
		if err := r.Status().Update(ctx, cluster); err != nil {
			logger.Error(err, "Error on Updating Cluster Condition Ready....")
			return err
		}
		r.Recorder.Eventf(
			cluster,
			"Warning",
			"ClusterUnhealthy",
			"Cannot find healthy pod in cluster %s",
			cluster.Name,
		)
		logger.Error(err, "Error on get pods...")
		return err
	}

	cluster.UpdateDesiredMaster(nextIndex)
	if err := r.Status().Update(ctx, cluster); err != nil {
		logger.Error(err, "Error on Updating Cluster Desired Master....")
		return err
	}

	logger.Info(fmt.Sprintf("Desired master updated to %s", strconv.Itoa(nextIndex)))
	return nil
}

func (r *OpenldapClusterReconciler) getJob(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (*batchv1.Job, error) {
	job := &batchv1.Job{}

	if err := r.Get(
		ctx,
		types.NamespacedName{
			Name:      cluster.JobName(),
			Namespace: cluster.Namespace,
		},
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

	if err := r.Get(
		ctx,
		types.NamespacedName{
			Name:      cluster.GetDesiredMaster(),
			Namespace: cluster.Namespace,
		},
		pod,
	); err != nil {
		return nil, err
	}

	return pod, nil
}

func (r *OpenldapClusterReconciler) getPod(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
	index int,
) (*corev1.Pod, error) {
	pod := &corev1.Pod{}

	if err := r.Get(
		ctx,
		types.NamespacedName{
			Name:      cluster.PodName(index),
			Namespace: cluster.Namespace,
		},
		pod,
	); err != nil {
		return nil, err
	}

	return pod, nil
}
