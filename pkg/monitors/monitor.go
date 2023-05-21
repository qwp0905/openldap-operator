package monitors

import (
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	"github.com/qwp0905/openldap-operator/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateServiceMonitor(cluster *openldapv1.OpenldapCluster) *monitoringv1.ServiceMonitor {
	trueValue := true
	falseValue := false

	return &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
			Labels: utils.MergeMap(
				cluster.SelectorLabels(),
				cluster.Spec.Monitor.Labels,
			),
			Annotations: cluster.GetAnnotations(),
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
					Port:            cluster.MetricsPortName(),
					Path:            cluster.MetricsPath(),
					Interval:        monitoringv1.Duration(cluster.Spec.Monitor.Interval),
					ScrapeTimeout:   monitoringv1.Duration(cluster.Spec.Monitor.ScrapeTimeout),
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
