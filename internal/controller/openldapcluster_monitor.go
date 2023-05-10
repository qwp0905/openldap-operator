package controller

import (
	"context"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	"github.com/qwp0905/openldap-operator/pkg/monitors"
	"github.com/qwp0905/openldap-operator/pkg/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *OpenldapClusterReconciler) setServiceMonitor(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (bool, error) {
	if !cluster.MonitorEnabled() {
		return false, nil
	}

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

		newServiceMonitor := monitors.CreateServiceMonitor(cluster)
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

	updatedServiceMonitor := monitors.CreateServiceMonitor(cluster)

	if r.compareServiceMonitor(existsServiceMonitor, updatedServiceMonitor) {
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

	if err := r.Get(
		ctx,
		types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace},
		monitor,
	); err != nil {
		return nil, err
	}

	return monitor, nil
}

func (r *OpenldapClusterReconciler) compareServiceMonitor(
	exists *monitoringv1.ServiceMonitor,
	new *monitoringv1.ServiceMonitor,
) bool {
	if !utils.CompareLabels(exists.Labels, new.Labels) {
		return false
	}

	if !utils.CompareLabels(exists.Annotations, new.Annotations) {
		return false
	}

	exEp := exists.Spec.Endpoints[0]
	neEp := new.Spec.Endpoints[0]

	return exEp.Path == neEp.Path &&
		exEp.Port == neEp.Port &&
		exEp.Interval == neEp.Interval
}
