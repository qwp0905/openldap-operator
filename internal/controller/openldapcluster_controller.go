/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"reflect"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
)

// OpenldapClusterReconciler reconciles a OpenldapCluster object
type OpenldapClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=openldap.kwonjin.click,resources=openldapclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=openldap.kwonjin.click,resources=openldapclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=openldap.kwonjin.click,resources=openldapclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the OpenldapCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *OpenldapClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	cluster, err := r.getCluster(ctx, req)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.setDefault(ctx, cluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.setConfigMap(ctx, cluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.setStatefulset(ctx, cluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *OpenldapClusterReconciler) getCluster(ctx context.Context, req ctrl.Request) (*openldapv1.OpenldapCluster, error) {
	logger := log.FromContext(ctx)
	cluster := &openldapv1.OpenldapCluster{}

	if err := r.Client.Get(ctx, req.NamespacedName, cluster); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Delete")
			return nil, nil
		}

		logger.Error(err, "error")
		return cluster, err
	}

	return cluster, nil
}

func (r *OpenldapClusterReconciler) setDefault(ctx context.Context, cluster *openldapv1.OpenldapCluster) error {
	logger := log.FromContext(ctx)
	origin := cluster.DeepCopy()

	cluster.SetDefault()

	if !reflect.DeepEqual(origin.Spec, cluster.Spec) {
		logger.Info("Admission controllers (webhooks) appear to have been disabled. " +
			"Please enable them for this object/namespace")

		err := r.Patch(ctx, cluster, client.MergeFrom(origin))
		if err != nil {
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OpenldapClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&openldapv1.OpenldapCluster{}).
		Complete(r)
}
