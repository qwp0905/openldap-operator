package controller

import (
	"context"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	"github.com/qwp0905/openldap-operator/pkg/rbac"
	"github.com/qwp0905/openldap-operator/pkg/utils"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *OpenldapClusterReconciler) ensureRoleBinding(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (bool, error) {
	logger := log.FromContext(ctx)
	existsRoleBinding, err := r.getRoleBinding(ctx, cluster)

	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error on getting RoleBinding....")
			return false, err
		}

		newRoleBinding := rbac.CreateRoleBinding(cluster)

		if err = r.registerObject(cluster, newRoleBinding); err != nil {
			logger.Error(err, "Error on Registering RoleBinding...")
			return false, err
		}

		if err = r.Create(ctx, newRoleBinding); err != nil {
			logger.Error(err, "Error on Creating RoleBinding...")
			return false, err
		}

		r.Recorder.Eventf(
			cluster,
			"Normal",
			"RoleBindingCreated",
			"RoleBinding %s created",
			newRoleBinding.Name,
		)
		logger.Info("RoleBinding Created")
		return true, nil
	}

	updatedRoleBinding := rbac.CreateRoleBinding(cluster)

	if r.compareRoleBinding(existsRoleBinding, updatedRoleBinding) {
		return false, nil
	}

	existsRoleBinding.SetLabels(updatedRoleBinding.GetLabels())
	existsRoleBinding.SetAnnotations(updatedRoleBinding.GetAnnotations())

	if err = r.Update(ctx, existsRoleBinding); err != nil {
		logger.Error(err, "Error on Updating RoleBinding....")
		return false, err
	}

	r.Recorder.Eventf(
		cluster,
		"Normal",
		"RoleBindingUpdated",
		"RoleBinding %s updated",
		updatedRoleBinding.Name,
	)
	logger.Info("RoleBinding Updated")
	return true, nil
}

func (r *OpenldapClusterReconciler) getRoleBinding(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (*rbacv1.RoleBinding, error) {
	roleBinding := &rbacv1.RoleBinding{}

	if err := r.Get(
		ctx,
		types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace},
		roleBinding,
	); err != nil {
		return nil, err
	}

	return roleBinding, nil
}

func (r *OpenldapClusterReconciler) compareRoleBinding(
	exists *rbacv1.RoleBinding,
	new *rbacv1.RoleBinding,
) bool {
	return utils.CompareMap(exists.Labels, new.Labels)
}
