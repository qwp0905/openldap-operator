package utils

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	// PodReasonEvicted is set inside the status as the Pod failure reason
	// when the Kubelet evicts a Pod
	PodReasonEvicted = "Evicted"
)

// PodStatus represent the possible status of pods
type PodStatus string

const (
	// PodHealthy means that a Pod is active and ready
	PodHealthy = "healthy"

	// PodReplicating means that a Pod is still not ready but still active
	PodReplicating = "replicating"

	// PodFailed means that a Pod will not be scheduled again (deleted or evicted)
	PodFailed = "failed"
)

// IsPodReady check if a Pod is ready or not
func IsPodReady(pod corev1.Pod) bool {
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.ContainersReady && c.Status == corev1.ConditionTrue {
			return true
		}
	}

	return false
}

// IsPodActive checks if a pod is active, copied from:
// https://github.com/kubernetes/kubernetes/blob/1bd0077/test/e2e/framework/pod/resource.go#L664
func IsPodActive(p corev1.Pod) bool {
	return corev1.PodSucceeded != p.Status.Phase &&
		corev1.PodPending != p.Status.Phase &&
		corev1.PodFailed != p.Status.Phase &&
		p.DeletionTimestamp == nil
}

// IsPodEvicted checks if a pod has been evicted by the
// Kubelet
func IsPodEvicted(p *corev1.Pod) bool {
	return corev1.PodFailed == p.Status.Phase &&
		PodReasonEvicted == p.Status.Reason
}

// IsPodUnscheduled check if a Pod is unscheduled
func IsPodUnscheduled(p *corev1.Pod) bool {
	if corev1.PodPending != p.Status.Phase && corev1.PodFailed != p.Status.Phase {
		return false
	}
	for _, c := range p.Status.Conditions {
		if c.Type == corev1.PodScheduled &&
			c.Status == corev1.ConditionFalse &&
			c.Reason == corev1.PodReasonUnschedulable {
			return true
		}
	}

	return false
}

// IsPodAlive check if a pod is active and not crash-looping
func IsPodAlive(p corev1.Pod) bool {
	if corev1.PodRunning == p.Status.Phase {
		for _, container := range append(p.Status.InitContainerStatuses, p.Status.ContainerStatuses...) {
			if container.State.Waiting != nil && container.State.Waiting.Reason == "CrashLoopBackOff" {
				return false
			}
		}
	}
	return IsPodActive(p)
}

// FilterActivePods returns pods that have not terminated.
func FilterActivePods(pods []corev1.Pod) []corev1.Pod {
	var result []corev1.Pod
	for _, p := range pods {
		if IsPodActive(p) {
			result = append(result, p)
		}
	}
	return result
}

// CountReadyPods counts the number of Pods which are ready
func CountReadyPods(podList []corev1.Pod) int {
	readyPods := 0
	for _, pod := range podList {
		if IsPodReady(pod) {
			readyPods++
		}
	}
	return readyPods
}

// ListStatusPods return a list of active Pods
func ListStatusPods(podList []corev1.Pod) map[PodStatus][]string {
	podsNames := make(map[PodStatus][]string)

	for _, pod := range podList {
		switch {
		case IsPodReady(pod):
			podsNames[PodHealthy] = append(podsNames[PodHealthy], pod.Name)
		case IsPodActive(pod):
			podsNames[PodReplicating] = append(podsNames[PodReplicating], pod.Name)
		default:
			podsNames[PodFailed] = append(podsNames[PodFailed], pod.Name)
		}
	}

	return podsNames
}