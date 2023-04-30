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
) (bool, error) {
	logger := log.FromContext(ctx)
	existsConfigMap, err := r.getConfigMap(ctx, cluster)

	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error on Get ConfigMap...")
			return false, err
		}

		newConfigMap := r.configMap(cluster)
		err = ctrl.SetControllerReference(cluster, newConfigMap, r.Scheme)
		if err != nil {
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
		return false, nil
	}

	if existsConfigMap.Data["LDAP_TLS"] == "true" && !cluster.Spec.OpenldapConfig.Tls.Enabled {
		return false, fmt.Errorf("cannot disable tls if once enabled")
	}

	existsConfigMap.Data = updatedConfigMap.Data
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

	err := r.Client.Get(
		ctx,
		types.NamespacedName{
			Name:      cluster.ConfigMapName(),
			Namespace: cluster.Namespace,
		},
		configMap,
	)

	if err != nil {
		return nil, err
	}

	return configMap, nil
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
		"LDAP_BACKEND":             cluster.Spec.OpenldapConfig.Backend,
		"LDAP_REPLICATION":         "true",
		"LDAP_REPLICATION_HOSTS":   cluster.ReplicationHosts(),
	}
}

func (r *OpenldapClusterReconciler) setExporterConfigMap(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (bool, error) {
	logger := log.FromContext(ctx)
	existsConfigMap, err := r.getExporterConfigMap(ctx, cluster)

	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error on Get Exporter ConfigMap...")
			return false, err
		}

		newConfigMap := r.exporterConfigMap(cluster)
		err = ctrl.SetControllerReference(cluster, newConfigMap, r.Scheme)
		if err != nil {
			logger.Error(err, "Error on Registering Exporter ConfigMap...")
			return false, nil
		}

		if err = r.Create(ctx, newConfigMap); err != nil {
			logger.Error(err, "Error on Creating Exporter ConfigMap...")
			return false, err
		}

		logger.Info("Exporter ConfigMap Created!")
		return true, nil
	}

	updatedConfigMap := r.exporterConfigMap(cluster)

	if reflect.DeepEqual(existsConfigMap.Data, updatedConfigMap.Data) {
		logger.Info("Nothing to update on Exporter ConfigMap")
		return false, nil
	}

	if err = r.Update(ctx, updatedConfigMap); err != nil {
		logger.Error(err, "Error on Updating Exporter ConfigMap...")
		return false, err
	}

	logger.Info("Exporter ConfigMap Updated")
	return true, nil
}

func (r *OpenldapClusterReconciler) getExporterConfigMap(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{}

	err := r.Client.Get(
		ctx,
		types.NamespacedName{
			Name:      cluster.ExporterName(),
			Namespace: cluster.Namespace,
		},
		configMap,
	)

	if err != nil {
		return nil, err
	}

	return configMap, nil
}

func (r *OpenldapClusterReconciler) exporterConfigMap(
	cluster *openldapv1.OpenldapCluster,
) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.ExporterName(),
			Namespace: cluster.Namespace,
			Labels:    cluster.SelectorLabels(),
		},
		Data: map[string]string{
			"openldap_exporter.yml": fmt.Sprintf(
				`server: tcp,port=%s
client: tcp:host=127.0.0.1:port=%s`,
				strconv.Itoa(metricsPort),
				strconv.Itoa(int(cluster.LdapPort())),
			),
		},
	}
}
