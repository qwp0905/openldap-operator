package jobs

import (
	"fmt"

	openldapv1 "github.com/qwp0905/openldap-operator/api/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateSlaveToMasterJob(cluster *openldapv1.OpenldapCluster) *batchv1.Job {
	ttl := int32(30)

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.JobName(),
			Namespace: cluster.Namespace,
			Labels:    cluster.JobLabels(),
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttl,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"sidecar.istio.io/inject": "false",
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: cluster.Name,
					RestartPolicy:      corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{
						{
							Name:            "exec",
							Image:           "alpine/k8s:1.25.6",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Command: []string{
								"/bin/bash",
								"-c",
								fmt.Sprintf(
									`kubectl exec %s -n %s -- /bin/bash -c "sh /opt/repl/master"`,
									cluster.GetDesiredMaster(),
									cluster.Namespace,
								),
							},
						},
					},
				},
			},
		},
	}
}
