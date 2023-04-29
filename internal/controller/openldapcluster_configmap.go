package controller

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *OpenldapClusterReconciler) setConfigMap(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) error {
	logger := log.FromContext(ctx)
	existsConfigMap, err := r.getConfigMap(ctx, cluster)

	if errors.IsNotFound(err) {

		newConfigMap := r.configMap(cluster)
		ctrl.SetControllerReference(cluster, newConfigMap, r.Scheme)

		if err = r.Create(ctx, newConfigMap); err != nil {
			logger.Error(err, "Error on Creating ConfigMap...")
			return err
		}

		logger.Info("ConfigMap Created!")
		return nil
	}

	updatedConfigMap := r.configMap(cluster)

	if reflect.DeepEqual(existsConfigMap.Data, updatedConfigMap.Data) &&
		reflect.DeepEqual(
			existsConfigMap.Labels,
			updatedConfigMap.Labels,
		) &&
		reflect.DeepEqual(
			existsConfigMap.Annotations,
			updatedConfigMap.Annotations,
		) {
		logger.Info("NotThing to change on ConfigMap")
		return nil
	}

	if existsConfigMap.Data["LDAP_TLS"] == "true" && !cluster.Spec.OpenldapConfig.Tls.Enabled {
		return fmt.Errorf("Cannot disable tls if once enabled")
	}

	existsConfigMap.Data = updatedConfigMap.Data
	existsConfigMap.Labels = updatedConfigMap.Labels
	existsConfigMap.Annotations = updatedConfigMap.Annotations

	err = r.Update(ctx, existsConfigMap)
	if err != nil {
		logger.Error(err, "Error on Updating ConfigMap...")
		return err
	}

	return nil
}

func (r *OpenldapClusterReconciler) getConfigMap(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{}

	err := r.Client.Get(
		ctx,
		types.NamespacedName{Name: cluster.ConfigMapName(), Namespace: cluster.Namespace},
		configMap,
	)

	if err != nil {
		return nil, err
	}

	return configMap, nil
}

func (r *OpenldapClusterReconciler) createConfigMap(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) error {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.ConfigMapName(),
			Namespace: cluster.Namespace,
			Labels:    cluster.SelectorLabels(),
		},
		Data: r.getConfigMapData(cluster),
	}

	return r.Create(ctx, configMap)
}

func (r *OpenldapClusterReconciler) configMap(
	cluster *openldapv1.OpenldapCluster,
) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.ConfigMapName(),
			Namespace: cluster.Namespace,
			Labels:    cluster.SelectorLabels(),
		},
		Data: r.getConfigMapData(cluster),
	}
}

func (r *OpenldapClusterReconciler) getConfigMapData(
	cluster *openldapv1.OpenldapCluster,
) map[string]string {
	return map[string]string{
		"LDAP_TLS":                 strconv.FormatBool(cluster.Spec.OpenldapConfig.Tls.Enabled),
		"LDAP_TLS_ENFORCE":         strconv.FormatBool(cluster.Spec.OpenldapConfig.Tls.Enforced),
		"LDAP_TLS_CRT_FILENAME":    cluster.Spec.OpenldapConfig.Tls.CertFile,
		"LDAP_TLS_KEY_FILENAME":    cluster.Spec.OpenldapConfig.Tls.KeyFile,
		"LDAP_TLS_CA_CRT_FILENAME": cluster.Spec.OpenldapConfig.Tls.CaFile,
		"LDAP_ORGANISATION":        cluster.Spec.OpenldapConfig.Organization,
		"LDAP_DOMAIN":              cluster.Spec.OpenldapConfig.Domain,
		"LDAP_LOG_LEVEL":           strconv.Itoa(int(cluster.Spec.OpenldapConfig.LogLevel)),
		"LDAP_BACKEND":             cluster.Spec.OpenldapConfig.Backend,
		"LDAP_REPLICATION":         "true",
		"LDAP_REPLICATION_HOSTS":   cluster.ReplicationHosts(),
	}
}
