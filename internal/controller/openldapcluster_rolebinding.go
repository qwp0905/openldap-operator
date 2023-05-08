package controller

import (
	"context"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	"github.com/qwp0905/openldap-operator/pkg/rbac"
	"github.com/qwp0905/openldap-operator/pkg/utils"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *OpenldapClusterReconciler) setRoleBinding(
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

		if err = ctrl.SetControllerReference(cluster, newRoleBinding, r.Scheme); err != nil {
			logger.Error(err, "Error on Registering RoleBinding...")
			return false, err
		}

		if err = r.Create(ctx, newRoleBinding); err != nil {
			logger.Error(err, "Error on Creating RoleBinding...")
			return false, err
		}

		logger.Info("RoleBinding Created")
		return true, nil
	}

	newRoleBinding := rbac.CreateRoleBinding(cluster)

	if r.compareRoleBinding(existsRoleBinding, newRoleBinding) {
		logger.Info("Nothing to change on RoleBinding")
		return false, nil
	}

	if err = r.Update(ctx, newRoleBinding); err != nil {
		logger.Error(err, "Error on Updating RoleBinding....")
		return false, err
	}

	logger.Info("RoleBinding Updated")
	return true, nil
}

func (r *OpenldapClusterReconciler) getRoleBinding(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (*rbacv1.RoleBinding, error) {
	roleBinding := &rbacv1.RoleBinding{}

	err := r.Get(
		ctx,
		types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace},
		roleBinding,
	)

	return roleBinding, err
}

func (r *OpenldapClusterReconciler) compareRoleBinding(
	exists *rbacv1.RoleBinding,
	new *rbacv1.RoleBinding,
) bool {
	if !utils.CompareLabels(exists.Labels, new.Labels) {
		return false
	}

	return true
}