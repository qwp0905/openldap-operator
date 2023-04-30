package controller

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *OpenldapClusterReconciler) setServiceMonitor(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (bool, error) {
	logger := log.FromContext(ctx)
	existsServiceMonitor, err := r.getServiceMonitor(ctx, cluster)

	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error on Get ServiceMonitor...")
			return false, err
		}

		if !cluster.MonitorEnabled() {
			logger.Info("Nothing to update on service monitor")
			return false, nil
		}

		newServiceMonitor := r.serviceMonitor(cluster)
		err = ctrl.SetControllerReference(cluster, newServiceMonitor, r.Scheme)
		if err != nil {
			logger.Error(err, "Error on Registering ServiceMonitor...")
			return false, err
		}

		if err = r.Create(ctx, newServiceMonitor); err != nil {
			logger.Error(err, "Error on Creating ServiceMonitor...")
			return false, err
		}

		logger.Info("ServiceMonitor Created")
		return true, nil
	}

	if !cluster.MonitorEnabled() {
		if err = r.Delete(ctx, existsServiceMonitor); err != nil {
			logger.Error(err, "Error on Deleting ServiceMonitor...")
			return false, err
		}

		logger.Info("ServiceMonitor Deleted")
		return true, nil
	}

	updatedServiceMonitor := r.serviceMonitor(cluster)

	if reflect.DeepEqual(existsServiceMonitor, updatedServiceMonitor) {
		logger.Info("NotThing to change on ServiceMonitor")
		return false, nil
	}

	if err = r.Update(ctx, updatedServiceMonitor); err != nil {
		logger.Error(err, "Error on Updating ServiceMonitor...")
		return false, err
	}

	logger.Info("ServiceMonitor Updated")
	return true, nil
}

func (r *OpenldapClusterReconciler) getServiceMonitor(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (*monitoringv1.ServiceMonitor, error) {
	monitor := &monitoringv1.ServiceMonitor{}

	if err := r.Client.Get(
		ctx,
		types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace},
		monitor,
	); err != nil {
		return nil, err
	}

	return monitor, nil
}

func (r *OpenldapClusterReconciler) serviceMonitor(
	cluster *openldapv1.OpenldapCluster,
) *monitoringv1.ServiceMonitor {

	trueValue := true
	falseValue := false

	labels := cluster.SelectorLabels()

	for key, val := range cluster.Spec.Monitor.Labels {
		labels[key] = val
	}

	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
			Labels:    labels,
		},
		Spec: monitoringv1.ServiceMonitorSpec{
			JobLabel: "app.kubernetes.io/name",
			NamespaceSelector: monitoringv1.NamespaceSelector{
				MatchNames: []string{cluster.Namespace},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: cluster.SelectorLabels(),
			},
			Endpoints: []monitoringv1.Endpoint{
				{
					Port:            metricsPortName,
					Path:            metricsPath,
					Interval:        monitoringv1.Duration(cluster.Spec.Monitor.Interval),
					HonorTimestamps: &trueValue,
					HonorLabels:     true,
					EnableHttp2:     &falseValue,
					FilterRunning:   &trueValue,
					FollowRedirects: &falseValue,
				},
			},
		},
	}
}

func (r *OpenldapClusterReconciler) setMonitorConfigMap(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (bool, error) {
	logger := log.FromContext(ctx)
	existsMonitorConfigMap, err := r.getMonitorConfigMap(ctx, cluster)

	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error on Getting Monitoring configmap...")
			return false, err
		}

		newMonitorConfig, err := r.monitorConfigMap(ctx, cluster)
		if err != nil {
			logger.Error(err, "Error on Making Monitor ConfigMap...")
			return false, err
		}

		err = ctrl.SetControllerReference(cluster, newMonitorConfig, r.Scheme)
		if err != nil {
			logger.Error(err, "Error on Registering Monitor ConfigMap...")
			return false, err
		}

		if err = r.Create(ctx, newMonitorConfig); err != nil {
			logger.Error(err, "Error on Creating Monitor ConfigMap...")
			return false, err
		}

		logger.Info("Monitor ConfigMap Created")
		return true, nil
	}

	newMonitorConfigMap, err := r.monitorConfigMap(ctx, cluster)
	if err != nil {
		logger.Error(err, "Error on Creating Monitor ConfigMap...")
		return false, err
	}

	if r.compareMonitorConfigMap(existsMonitorConfigMap, newMonitorConfigMap) {
		logger.Info("Nothing to Update on MonitorConfigMap")
		return false, nil
	}

	if err = r.Update(ctx, newMonitorConfigMap); err != nil {
		logger.Error(err, "Error on Updating Monitor ConfigMap...")
		return false, err
	}

	logger.Info("Monitor ConfigMap updated")
	return true, nil
}

func (r *OpenldapClusterReconciler) getMonitorConfigMap(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{}

	err := r.Client.Get(
		ctx,
		types.NamespacedName{
			Name:      cluster.MonitorConfigMapName(),
			Namespace: cluster.Namespace,
		},
		configMap,
	)

	if err != nil {
		return nil, err
	}

	return configMap, nil
}

func (r *OpenldapClusterReconciler) monitorConfigMap(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (*corev1.ConfigMap, error) {

	initScript, err := r.initScript(ctx, cluster)
	if err != nil {
		return nil, err
	}

	monitorScript := cluster.MonitorInitScript(
		fmt.Sprintf("%s/module_monitor.ldif", monitorDataMountPath),
		fmt.Sprintf("%s/database_monitor.ldif", monitorDataMountPath),
	)

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.MonitorConfigMapName(),
			Namespace: cluster.Namespace,
			Labels:    cluster.SelectorLabels(),
		},
		Data: map[string]string{
			"module_monitor.ldif":   cluster.ModuleMonitorLdif(),
			"database_monitor.ldif": cluster.DatabaseModuleLdif(),
			"bootstrap": fmt.Sprintf(`#!/bin/bash
%s
%s`, initScript, monitorScript),
		},
	}, nil
}

func (r *OpenldapClusterReconciler) compareMonitorConfigMap(
	exists *corev1.ConfigMap,
	new *corev1.ConfigMap,
) bool {
	for key, val := range new.Labels {
		if exists.Labels[key] != val {
			return false
		}
	}

	for key, val := range new.Annotations {
		if exists.Annotations[key] != val {
			return false
		}
	}

	return true
}

func (r *OpenldapClusterReconciler) initScript(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (string, error) {

	if (cluster.Spec.OpenldapConfig.SeedData == nil) ||
		(cluster.Spec.OpenldapConfig.SeedData.ConfigMap == nil &&
			cluster.Spec.OpenldapConfig.SeedData.Secret == nil) {
		return "", nil
	}

	var seedData map[string]string
	if cluster.Spec.OpenldapConfig.SeedData.ConfigMap != nil {
		seedConfigMap := &corev1.ConfigMap{}
		err := r.Client.Get(
			ctx,
			types.NamespacedName{
				Name:      cluster.Spec.OpenldapConfig.SeedData.ConfigMap.Name,
				Namespace: cluster.Namespace,
			},
			seedConfigMap,
		)

		if err != nil {
			return "", err
		}
		seedData = seedConfigMap.Data
	} else {
		seedSecret := &corev1.Secret{}
		err := r.Client.Get(
			ctx,
			types.NamespacedName{
				Name:      cluster.Spec.OpenldapConfig.SeedData.Secret.SecretName,
				Namespace: cluster.Namespace,
			},
			seedSecret,
		)
		if err != nil {
			return "", err
		}
		seedData = seedSecret.StringData
	}

	args := []string{}
	for key := range seedData {
		args = append(
			args, fmt.Sprintf(
				"ldapadd -x -D \"%s\" -w $LDAP_ADMIN_PASSWORD -f %s/%s",
				cluster.BindDn(),
				seedDataMountPath,
				key,
			),
		)
	}

	return strings.Join(args, ";"), nil
}
