package controller

import (
	"context"
	"reflect"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *OpenldapClusterReconciler) setService(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (bool, error) {
	logger := log.FromContext(ctx)
	existsService, err := r.getService(ctx, cluster)

	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error on Getting Service...")
			return false, err
		}

		newService := r.service(cluster)
		err = ctrl.SetControllerReference(cluster, newService, r.Scheme)
		if err != nil {
			logger.Error(err, "Error on Registering Service...")
			return false, err
		}

		if err := r.Create(ctx, newService); err != nil {
			logger.Error(err, "Error on Creating Service...")
			return false, err
		}

		logger.Info("Service Created")
		return true, nil
	}

	updatedService := r.service(cluster)

	if r.compareService(existsService, updatedService) {
		logger.Info("Nothing to update on Service")
		return false, nil
	}

	if err = r.Update(ctx, updatedService); err != nil {
		logger.Error(err, "Error on Updating Service...")
		return false, err
	}

	logger.Info("Service Updated")
	return true, nil
}

func (r *OpenldapClusterReconciler) getService(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (*corev1.Service, error) {

	service := &corev1.Service{}

	err := r.Client.Get(
		ctx,
		types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace},
		service,
	)

	if err != nil {
		return nil, err
	}

	return service, nil
}

func (r *OpenldapClusterReconciler) service(
	cluster *openldapv1.OpenldapCluster,
) *corev1.Service {

	ldapPort := intstr.FromInt(389)
	ldapsPort := intstr.FromInt(636)

	ports := []corev1.ServicePort{
		{
			Name:       "ldap",
			Protocol:   "TCP",
			Port:       cluster.LdapPort(),
			TargetPort: ldapPort,
		},
		{
			Name:     metricsPortName,
			Protocol: "TCP",
			Port:     metricsPort,
		},
	}

	if cluster.TlsEnabled() {
		ports = append(ports, corev1.ServicePort{
			Name:       "ldaps",
			Protocol:   "TCP",
			Port:       cluster.LdapsPort(),
			TargetPort: ldapsPort,
		})
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
			Labels:    cluster.SelectorLabels(),
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: cluster.SelectorLabels(),
			Ports:    ports,
		},
	}
}

func (r *OpenldapClusterReconciler) compareService(
	exists *corev1.Service,
	new *corev1.Service,
) bool {
	for label, value := range new.Labels {
		if exists.Labels[label] != value {
			return false
		}
	}

	for annotation, value := range new.Annotations {
		if exists.Annotations[annotation] != value {
			return false
		}
	}

	return reflect.DeepEqual(exists.Spec, new.Spec)
}
