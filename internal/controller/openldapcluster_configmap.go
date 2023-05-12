package controller

import (
	"context"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	"github.com/qwp0905/openldap-operator/pkg/configmaps"
	"github.com/qwp0905/openldap-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *OpenldapClusterReconciler) setConfigMap(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (bool, error) {
	logger := log.FromContext(ctx)
	existsConfigMap, err := r.getConfigMap(ctx, cluster)

	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error on Get ConfigMap...")
			return false, err
		}

		newConfigMap := configmaps.CreateConfigMap(cluster)
		if err = r.registerObject(cluster, newConfigMap); err != nil {
			logger.Error(err, "Error on Registering ConfigMap...")
			return false, err
		}

		if err = r.Create(ctx, newConfigMap); err != nil {
			logger.Error(err, "Error on Creating ConfigMap...")
			return false, err
		}

		logger.Info("ConfigMap Created!")
		return true, nil
	}

	updatedConfigMap := configmaps.CreateConfigMap(cluster)

	if r.compareConfigMap(existsConfigMap, updatedConfigMap) {
		return false, nil
	}

	existsConfigMap.Labels = updatedConfigMap.Labels
	existsConfigMap.Annotations = updatedConfigMap.Annotations

	if err = r.Update(ctx, existsConfigMap); err != nil {
		logger.Error(err, "Error on Updating ConfigMap...")
		return false, err
	}

	logger.Info("ConfigMap Updated")
	return true, nil
}

func (r *OpenldapClusterReconciler) getConfigMap(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{}

	if err := r.Get(
		ctx,
		types.NamespacedName{
			Name:      cluster.ConfigMapName(),
			Namespace: cluster.Namespace,
		},
		configMap,
	); err != nil {
		return nil, err
	}

	return configMap, nil
}

func (r *OpenldapClusterReconciler) compareConfigMap(
	exists *corev1.ConfigMap,
	new *corev1.ConfigMap,
) bool {
	if !utils.CompareLabels(exists.Labels, new.Labels) {
		return false
	}

	if !utils.CompareLabels(exists.Annotations, new.Annotations) {
		return false
	}

	return true
}
