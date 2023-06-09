package controller

import (
	"context"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	"github.com/qwp0905/openldap-operator/pkg/rbac"
	"github.com/qwp0905/openldap-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *OpenldapClusterReconciler) ensureServiceAccount(
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

		if err = r.registerObject(cluster, newServiceAccount); err != nil {
			logger.Error(err, "Error on Registering ServiceAccount...")
			return false, err
		}

		if err = r.Create(ctx, newServiceAccount); err != nil {
			logger.Error(err, "Error on Creating ServiceAccount...")
			return false, err
		}

		r.Recorder.Eventf(
			cluster,
			"Normal",
			"ServiceAccountCreated",
			"ServiceAccount %s created",
			newServiceAccount.Name,
		)
		logger.Info("ServiceAccount Created!")
		return true, nil
	}

	updatedServiceAccount := rbac.CreateServiceAccount(cluster)

	if r.compareServiceAccount(existsServiceAccount, updatedServiceAccount) {
		return false, nil
	}

	existsServiceAccount.SetLabels(updatedServiceAccount.GetLabels())
	existsServiceAccount.SetAnnotations(updatedServiceAccount.GetAnnotations())

	if err = r.Update(ctx, existsServiceAccount); err != nil {
		logger.Error(err, "Error on Updating ServiceAccount...")
		return false, err
	}

	r.Recorder.Eventf(
		cluster,
		"Normal",
		"ServiceAccountUpdated",
		"ServiceAccount %s updated",
		updatedServiceAccount.Name,
	)
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
	if !utils.CompareMap(exists.Labels, new.Labels) {
		return false
	}

	if !utils.CompareMap(exists.Annotations, new.Annotations) {
		return false
	}

	return true
}
