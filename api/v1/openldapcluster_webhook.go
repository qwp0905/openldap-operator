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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const (
	defaultTlsEnabled     = false
	defaultRoot           = "dc=example,dc=com"
	defaultMonitorEnabled = false
	defaultAdmin          = "admin"
	defaultConfig         = "config"
)

// log is for logging in this package.
var openldapclusterlog = logf.Log.WithName("openldapcluster-resource")

func (r *OpenldapCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-openldap-kwonjin-click-v1-openldapcluster,mutating=true,failurePolicy=fail,sideEffects=None,groups=openldap.kwonjin.click,resources=openldapclusters,verbs=create;update,versions=v1,name=mopenldapcluster.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &OpenldapCluster{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *OpenldapCluster) Default() {
	openldapclusterlog.Info("default", "name", r.Name)

	r.SetDefault()
}

//+kubebuilder:webhook:path=/validate-openldap-kwonjin-click-v1-openldapcluster,mutating=false,failurePolicy=fail,sideEffects=None,groups=openldap.kwonjin.click,resources=openldapclusters,verbs=create;update,versions=v1,name=vopenldapcluster.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &OpenldapCluster{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *OpenldapCluster) ValidateCreate() error {
	openldapclusterlog.Info("validate create", "name", r.Name)

	apierrs := field.ErrorList{}

	if err := r.validateTlsSecret(); err != nil {
		apierrs = append(apierrs, err)
	}

	if len(apierrs) > 0 {
		return r.createError(apierrs)
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *OpenldapCluster) ValidateUpdate(old runtime.Object) error {
	openldapclusterlog.Info("validate update", "name", r.Name)

	oldCluster := old.(*OpenldapCluster)
	oldCluster.SetDefault()

	apierrs := field.ErrorList{}

	if err := r.validateTlsSecret(); err != nil {
		apierrs = append(apierrs, err)
	}

	if err := r.validateTlsChanged(oldCluster); err != nil {
		apierrs = append(apierrs, err)
	}

	if len(apierrs) > 0 {
		return r.createError(apierrs)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *OpenldapCluster) ValidateDelete() error {
	openldapclusterlog.Info("validate delete", "name", r.Name)

	return nil
}

func (r *OpenldapCluster) SetDefault() {
	if r.Spec.OpenldapConfig == nil {
		r.Spec.OpenldapConfig = &OpenldapConfig{
			Root:           defaultRoot,
			AdminUsername:  defaultAdmin,
			ConfigUsername: defaultConfig,
		}
	}

	if r.Spec.OpenldapConfig.Tls == nil {
		r.Spec.OpenldapConfig.Tls = &TlsConfig{
			Enabled: defaultTlsEnabled,
		}
	} else if r.Spec.OpenldapConfig.Tls.Enabled {
		if r.Spec.OpenldapConfig.Tls.CaFile == "" {
			r.Spec.OpenldapConfig.Tls.CaFile = "ca.crt"
		}

		if r.Spec.OpenldapConfig.Tls.KeyFile == "" {
			r.Spec.OpenldapConfig.Tls.KeyFile = "cert.key"
		}

		if r.Spec.OpenldapConfig.Tls.CertFile == "" {
			r.Spec.OpenldapConfig.Tls.CertFile = "cert.crt"
		}
	}

	if r.Spec.OpenldapConfig.Root == "" {
		r.Spec.OpenldapConfig.Root = defaultRoot
	}

	if r.Spec.Monitor == nil {
		r.Spec.Monitor = &MonitorConfig{Enabled: defaultMonitorEnabled}
	}

	if r.Spec.Ports == nil {
		r.Spec.Ports = &PortConfig{
			Ldap:  1389,
			Ldaps: 1636,
		}
	}

	if r.Spec.NodeSelector == nil {
		r.Spec.NodeSelector = map[string]string{}
	}

	if r.Spec.ImagePullSecrets == nil {
		r.Spec.ImagePullSecrets = []corev1.LocalObjectReference{}
	}

	if r.Spec.Affinity == nil {
		r.Spec.Affinity = &corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
					{
						Weight: 100,
						PodAffinityTerm: corev1.PodAffinityTerm{
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: r.SelectorLabels(),
							},
							TopologyKey: "kubernetes.io/hostname",
						},
					},
				},
			},
		}
	}
}

func (r *OpenldapCluster) createError(errList field.ErrorList) *errors.StatusError {
	return errors.NewInvalid(
		schema.GroupKind{Group: r.GroupVersionKind().Group, Kind: r.Kind},
		r.Name,
		errList,
	)
}

func (r *OpenldapCluster) validateTlsSecret() *field.Error {
	if r.TlsEnabled() && r.Spec.OpenldapConfig.Tls.SecretName == "" {
		return &field.Error{
			Type:     field.ErrorTypeForbidden,
			Field:    "spec.openldapConfig.tls.secretName",
			BadValue: r.Spec.OpenldapConfig.Tls.SecretName,
			Detail:   "If tls enabled, secret name must be provided",
		}
	}

	return nil
}

func (r *OpenldapCluster) validateTlsChanged(old *OpenldapCluster) *field.Error {
	if old.TlsEnabled() != r.TlsEnabled() {
		return &field.Error{
			Type:     field.ErrorTypeForbidden,
			Field:    "spec.openldapConfig.tls",
			BadValue: r.TlsEnabled(),
			Detail:   "Cannot change tls configuration after created",
		}
	}

	return nil
}
