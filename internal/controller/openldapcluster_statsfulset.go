package controller

import (
	"context"
	"reflect"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	"github.com/qwp0905/openldap-operator/pkg/statefulsets"
	"github.com/qwp0905/openldap-operator/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *OpenldapClusterReconciler) setStatefulset(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (bool, error) {
	logger := log.FromContext(ctx)
	existsStatefulset, err := r.getStatefulset(ctx, cluster)

	if err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, "Error on getting Statefulset...")
			return false, err
		}

		newStatefulset := statefulsets.CreateStatefulset(cluster)

		if err = r.registerObject(cluster, newStatefulset); err != nil {
			logger.Error(err, "Error on Registering Statefulset...")
			return false, err
		}

		if err = r.Create(ctx, newStatefulset); err != nil {
			logger.Error(err, "Error on Creating Statefulset...")
			return false, err
		}

		logger.Info("Statefulset Created")
		return true, nil
	}

	updatedStatefulset := statefulsets.CreateStatefulset(cluster)

	if compareStatefulset(existsStatefulset, updatedStatefulset) {
		return false, nil
	}

	existsStatefulset.Spec.Replicas = updatedStatefulset.Spec.Replicas
	existsStatefulset.Spec.Template.Spec = updatedStatefulset.Spec.Template.Spec
	existsStatefulset.Spec.VolumeClaimTemplates = updatedStatefulset.Spec.VolumeClaimTemplates
	existsStatefulset.SetLabels(updatedStatefulset.GetLabels())
	existsStatefulset.SetAnnotations(updatedStatefulset.GetAnnotations())

	if err = r.Update(ctx, existsStatefulset); err != nil {
		logger.Error(err, "Error on Updating Statefulset...")
		return false, err
	}

	logger.Info("Statefulset Updated")
	return true, nil
}

func (r *OpenldapClusterReconciler) getStatefulset(
	ctx context.Context,
	cluster *openldapv1.OpenldapCluster,
) (*appsv1.StatefulSet, error) {

	statefulset := &appsv1.StatefulSet{}

	if err := r.Get(
		ctx,
		types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace},
		statefulset,
	); err != nil {
		return nil, err
	}

	return statefulset, nil
}

func compareStatefulset(exists *appsv1.StatefulSet, new *appsv1.StatefulSet) bool {
	if !utils.CompareLabels(exists.Labels, new.Labels) {
		return false
	}

	if !utils.CompareLabels(exists.Annotations, new.Annotations) {
		return false
	}

	exCon := statefulsets.GetContainer(exists, exists.Name)
	neCon := statefulsets.GetContainer(new, new.Name)

	return *exists.Spec.Replicas == *new.Spec.Replicas &&
		reflect.DeepEqual(exCon.Resources, neCon.Resources) &&
		utils.CompareEnv(*exCon, *neCon) &&
		reflect.DeepEqual(exists.Spec.VolumeClaimTemplates, new.Spec.VolumeClaimTemplates)
}
