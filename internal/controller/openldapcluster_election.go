package controller

import (
	"context"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	"github.com/qwp0905/openldap-operator/pkg/jobs"
	"github.com/qwp0905/openldap-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *OpenldapClusterReconciler) election(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (bool, error) {
	logger := log.FromContext(ctx)

	var next int
	var pod *corev1.Pod

	for i := 0; i < int(cluster.Spec.Replicas); i++ {
		p, err := r.getPod(ctx, cluster, i)
		if cluster.PodName(i) != cluster.Status.Master &&
			err == nil &&
			utils.IsPodReady(*p) {
			pod = p
			next = i
			break
		}
	}

	cluster.UpdateMaster(next)
	pod.Labels = cluster.MasterSelectorLabels()

	if err := r.Update(ctx, pod); err != nil {
		logger.Error(err, "Error on Election....")
		return false, err
	}

	masterJob := jobs.CreateSlaveToMasterJob(cluster)

	if err := ctrl.SetControllerReference(cluster, masterJob, r.Scheme); err != nil {
		logger.Error(err, "Error on Registering Master Job...")
		return false, err
	}

	if err := r.Create(ctx, masterJob); err != nil {
		logger.Error(err, "Error on Creating Master Job...")
		return false, err
	}

	for i := 0; i < int(cluster.Spec.Replicas); i++ {
		if cluster.Status.Master == "" {
		}
	}

}
