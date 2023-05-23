package controller

import (
	"context"
	"fmt"
	"reflect"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	"github.com/qwp0905/openldap-operator/pkg/statefulsets"
	"github.com/qwp0905/openldap-operator/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *OpenldapClusterReconciler) ensureStatefulset(
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

		r.Recorder.Eventf(
			cluster,
			"Normal",
			"StatefulsetCreated",
			"Statefulset %s created",
			newStatefulset.Name,
		)
		logger.Info("Statefulset Created")
		return true, nil
	}

	updatedStatefulset := statefulsets.CreateStatefulset(cluster)

	reason, result := compareStatefulset(existsStatefulset, updatedStatefulset)
	if result {
		return false, nil
	}

	existsStatefulset.SetLabels(updatedStatefulset.GetLabels())
	existsStatefulset.SetAnnotations(updatedStatefulset.GetAnnotations())
	existsStatefulset.Spec = updatedStatefulset.Spec

	if err = r.Update(ctx, existsStatefulset); err != nil {
		logger.Error(err, "Error on Updating Statefulset...")
		return false, err
	}

	r.Recorder.Eventf(
		cluster,
		"Normal",
		"StatefulsetUpdated",
		"Statefulset %s updated",
		updatedStatefulset.Name,
	)
	logger.Info(fmt.Sprintf("Statefulset Updated %s changed", reason))
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

func compareStatefulset(exists *appsv1.StatefulSet, new *appsv1.StatefulSet) (string, bool) {
	if !utils.CompareMap(exists.Labels, new.Labels) {
		return "labels", false
	}

	if !utils.CompareMap(exists.Annotations, new.Annotations) {
		return "annotations", false
	}

	exCon := statefulsets.GetContainer(exists, exists.Name)
	neCon := statefulsets.GetContainer(new, new.Name)

	if *exists.Spec.Replicas != *new.Spec.Replicas {
		return "replicas", false
	}

	if !utils.CompareEnv(*exCon, *neCon) {
		return "env", false
	}

	if !reflect.DeepEqual(exCon.Resources, neCon.Resources) {
		return "resources", false
	}

	if exCon.Image != neCon.Image {
		return "image", false
	}

	if exCon.ImagePullPolicy != neCon.ImagePullPolicy {
		return "imagePullPolicy", false
	}

	if !utils.ComparePVC(
		exists.Spec.VolumeClaimTemplates[0],
		new.Spec.VolumeClaimTemplates[0],
	) {
		return "pvc", false
	}

	return "", true
}
