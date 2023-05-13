package controller

import (
	"context"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	"github.com/qwp0905/openldap-operator/pkg/services"
	"github.com/qwp0905/openldap-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *OpenldapClusterReconciler) setService(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (bool, error) {
	requeue, err := r.setWriteService(ctx, cluster)
	if err != nil {
		return false, err
	}
	if requeue {
		return true, nil
	}

	requeue, err = r.setReadService(ctx, cluster)
	if err != nil {
		return false, err
	}
	if requeue {
		return true, nil
	}

	requeue, err = r.setMetricsService(ctx, cluster)
	if err != nil {
		return false, err
	}
	if requeue {
		return true, nil
	}

	return false, nil
}

func (r *OpenldapClusterReconciler) getService(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
	name string,
) (*corev1.Service, error) {
	service := &corev1.Service{}

	if err := r.Get(
		ctx,
		types.NamespacedName{Name: name, Namespace: cluster.Namespace},
		service,
	); err != nil {
		return nil, err
	}

	return service, nil
}

func (r *OpenldapClusterReconciler) setWriteService(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (bool, error) {
	logger := log.FromContext(ctx)
	existsService, err := r.getService(ctx, cluster, cluster.WriteServiceName())

	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error on getting write service....")
			return false, err
		}

		newService := services.CreateWriteService(cluster)

		if err = r.registerObject(cluster, newService); err != nil {
			logger.Error(err, "Error on Registering Write Service...")
			return false, nil
		}

		if err = r.Create(ctx, newService); err != nil {
			logger.Error(err, "Error on Creating Write Service...")
			return false, nil
		}

		logger.Info("Write Service Created")
		return true, nil
	}

	newService := services.CreateWriteService(cluster)

	if r.compareService(existsService, newService) {
		return false, nil
	}

	if err = r.Update(ctx, newService); err != nil {
		logger.Error(err, "Error on Updating Write Service...")
		return false, err
	}

	logger.Info("Write Service Updated")
	return true, nil
}

func (r *OpenldapClusterReconciler) setReadService(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (bool, error) {
	logger := log.FromContext(ctx)
	existsService, err := r.getService(ctx, cluster, cluster.Name)

	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error on getting read service....")
			return false, err
		}

		newService := services.CreateReadService(cluster)

		if err = r.registerObject(cluster, newService); err != nil {
			logger.Error(err, "Error on Registering Read Service...")
			return false, nil
		}

		if err = r.Create(ctx, newService); err != nil {
			logger.Error(err, "Error on Creating read Service...")
			return false, nil
		}

		logger.Info("Read Service Created")
		return true, nil
	}

	newService := services.CreateReadService(cluster)

	if r.compareService(existsService, newService) {
		return false, nil
	}

	if err = r.Update(ctx, newService); err != nil {
		logger.Error(err, "Error on Updating Read Service...")
		return false, err
	}

	logger.Info("Read Service Updated")
	return true, nil
}

func (r *OpenldapClusterReconciler) setMetricsService(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (bool, error) {
	logger := log.FromContext(ctx)
	existsService, err := r.getService(ctx, cluster, cluster.MetricsServiceName())

	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error on getting metrics service....")
			return false, err
		}

		newService := services.CreateMetricsService(cluster)

		if err = r.registerObject(cluster, newService); err != nil {
			logger.Error(err, "Error on Registering Metrics Service...")
			return false, nil
		}

		if err = r.Create(ctx, newService); err != nil {
			logger.Error(err, "Error on Creating Metrics Service...")
			return false, nil
		}

		logger.Info("Metrics Service Created")
		return true, nil
	}

	newService := services.CreateMetricsService(cluster)

	if r.compareService(existsService, newService) {
		return false, nil
	}

	if err = r.Update(ctx, newService); err != nil {
		logger.Error(err, "Error on Updating Metrics Service...")
		return false, err
	}

	logger.Info("Metrics Service Updated")
	return true, nil
}

func (r *OpenldapClusterReconciler) compareService(
	exists *corev1.Service,
	new *corev1.Service,
) bool {
	if !utils.CompareLabels(exists.Labels, new.Labels) {
		return false
	}

	if !utils.CompareLabels(exists.Annotations, new.Annotations) {
		return false
	}

	exPorts := []int32{}
	for _, p := range exists.Spec.Ports {
		exPorts = append(exPorts, p.Port)
	}

	result := true
	for _, ne := range new.Spec.Ports {
		rr := false
		for _, ex := range exPorts {
			rr = (ne.Port == ex) || rr
		}

		result = result && rr
	}

	return result
}
