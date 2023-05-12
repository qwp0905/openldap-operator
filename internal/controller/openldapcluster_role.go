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

func (r *OpenldapClusterReconciler) setRole(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (bool, error) {
	logger := log.FromContext(ctx)
	existsRole, err := r.getRole(ctx, cluster)

	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error on getting Role....")
			return false, err
		}

		newRole := rbac.CreateRole(cluster)

		if err = r.registerObject(cluster, newRole); err != nil {
			logger.Error(err, "Error on Registering Role...")
			return false, err
		}

		if err = r.Create(ctx, newRole); err != nil {
			logger.Error(err, "Error on Creating Role...")
			return false, err
		}

		logger.Info("Role Created")
		return true, nil
	}

	newRole := rbac.CreateRole(cluster)

	if r.compareRole(existsRole, newRole) {
		return false, nil
	}

	if err = r.Update(ctx, newRole); err != nil {
		logger.Error(err, "Error on Updating Role...")
		return false, err
	}

	logger.Info("Role Updated")
	return true, nil
}

func (r *OpenldapClusterReconciler) getRole(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (*rbacv1.Role, error) {
	role := &rbacv1.Role{}

	if err := r.Get(
		ctx,
		types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace},
		role,
	); err != nil {
		return nil, err
	}

	return role, nil
}

func (r *OpenldapClusterReconciler) compareRole(
	exists *rbacv1.Role,
	new *rbacv1.Role,
) bool {
	return utils.CompareLabels(exists.Labels, new.Labels)
}
