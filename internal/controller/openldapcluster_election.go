package controller

import (
	"context"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
)

func (r *OpenldapClusterReconciler) checkMaster(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (bool, error) {

	masterPod := cluster.Status.Master
}
