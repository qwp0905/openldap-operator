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
			Name:      cluster.Status.Master,
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
									`kubectl exec %s -n %s -- /bin/bash -c "/opt/repl/initialize && /opt/repl/master"`,
									cluster.Status.Master,
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

func CreateMasterToSlaveJob(cluster *openldapv1.OpenldapCluster, index int) *batchv1.Job {
	ttl := int32(30)

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.PodName(index),
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
									`#!/bin/bash
									while [ "$(kubectl get pod %s -n %s -o jsonpath='{.status.phase}')" != "Ready" ];do
										sleep 5;
									done

									kubectl exec %s -n %s -- /bin/bash -c "/opt/repl/delete-master && /opt/repl/create-slave"`,
									cluster.PodName(index),
									cluster.Namespace,
									cluster.PodName(index),
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

func CreateSlaveToSlaveJob(cluster *openldapv1.OpenldapCluster, index int) *batchv1.Job {
	ttl := int32(30)

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.PodName(index),
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
									`#!/bin/bash
									while [ "$(kubectl get pod %s -n %s -o jsonpath='{.status.phase}')" != "Ready" ];do
										sleep 5;
									done

									kubectl exec %s -n %s -- /bin/bash -c "/opt/repl/delete-slave && /opt/repl/create-slave"`,
									cluster.PodName(index),
									cluster.Namespace,
									cluster.PodName(index),
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
