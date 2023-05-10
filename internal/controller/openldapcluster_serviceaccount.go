package controller

import (
	"context"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	"github.com/qwp0905/openldap-operator/pkg/rbac"
	"github.com/qwp0905/openldap-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *OpenldapClusterReconciler) setServiceAccount(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (bool, error) {
	logger := log.FromContext(ctx)
	existsServiceAccount, err := r.getServiceAccount(ctx, cluster)

	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error on Get ServiceAccount...")
			return false, err
		}

		newServiceAccount := rbac.CreateServiceAccount(cluster)
		err = ctrl.SetControllerReference(cluster, newServiceAccount, r.Scheme)
		if err != nil {
			logger.Error(err, "Error on Registering ServiceAccount...")
			return false, err
		}

		if err = r.Create(ctx, newServiceAccount); err != nil {
			logger.Error(err, "Error on Creating ServiceAccount...")
			return false, err
		}

		logger.Info("ServiceAccount Created!")
		return true, nil
	}

	updatedServiceAccount := rbac.CreateServiceAccount(cluster)

	if r.compareServiceAccount(existsServiceAccount, updatedServiceAccount) {
		return false, nil
	}

	existsServiceAccount.Labels = updatedServiceAccount.Labels
	existsServiceAccount.Annotations = updatedServiceAccount.Annotations

	if err = r.Update(ctx, existsServiceAccount); err != nil {
		logger.Error(err, "Error on Updating ServiceAccount...")
		return false, err
	}

	logger.Info("ServiceAccount Updated")
	return true, nil
}

func (r *OpenldapClusterReconciler) getServiceAccount(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (*corev1.ServiceAccount, error) {
	serviceAccount := &corev1.ServiceAccount{}

	if err := r.Get(
		ctx,
		types.NamespacedName{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
		serviceAccount,
	); err != nil {
		return nil, err
	}

	return serviceAccount, nil
}

func (r *OpenldapClusterReconciler) compareServiceAccount(
	exists *corev1.ServiceAccount,
	new *corev1.ServiceAccount,
) bool {
	if !utils.CompareLabels(exists.Labels, new.Labels) {
		return false
	}

	if !utils.CompareLabels(exists.Annotations, new.Annotations) {
		return false
	}

	return true
}
