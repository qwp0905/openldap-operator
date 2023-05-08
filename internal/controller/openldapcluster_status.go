package controller

import (
	"context"
	"sort"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type managedResources struct {
	pods corev1.PodList
	pvcs corev1.PersistentVolumeClaimList
}

func (r *OpenldapClusterReconciler) getManagedResources(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (*managedResources, error) {
	pods, err := r.getManagedPods(ctx, cluster)
	if err != nil {
		return nil, err
	}

	pvcs, err := r.getManagedPvcs(ctx, cluster)
	if err != nil {
		return nil, err
	}

	return &managedResources{pods: pods, pvcs: pvcs}, nil
}

func (r *OpenldapClusterReconciler) getManagedPods(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (corev1.PodList, error) {
	var childPods corev1.PodList
	if err := r.List(ctx, &childPods,
		client.InNamespace(cluster.Namespace),
		client.MatchingFields{"podOwnerKey": cluster.Name},
	); err != nil {
		log.FromContext(ctx).Error(err, "Unable to list child pods resource")
		return corev1.PodList{}, err
	}

	sort.Slice(childPods.Items, func(i, j int) bool {
		return childPods.Items[i].Name < childPods.Items[j].Name
	})

	return childPods, nil
}

func (r *OpenldapClusterReconciler) getManagedPvcs(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (corev1.PersistentVolumeClaimList, error) {
	var childPVCs corev1.PersistentVolumeClaimList
	if err := r.List(ctx, &childPVCs,
		client.InNamespace(cluster.Namespace),
		client.MatchingFields{"pvcOwnerKey": cluster.Name},
	); err != nil {
		log.FromContext(ctx).Error(err, "Unable to list child PVCs")
		return corev1.PersistentVolumeClaimList{}, err
	}

	sort.Slice(childPVCs.Items, func(i, j int) bool {
		return childPVCs.Items[i].Name < childPVCs.Items[j].Name
	})

	return childPVCs, nil
}

func (r *OpenldapClusterReconciler) updateResourcesStatus(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
	resources *managedResources,
) error {
	existsStatus := cluster.Status

}
