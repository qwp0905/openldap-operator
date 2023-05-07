package utils

import (
	batchv1 "k8s.io/api/batch/v1"
)

// JobHasOneCompletion Completion check if a certain job is complete
func JobHasOneCompletion(job batchv1.Job) bool {
	requestedCompletions := int32(1)
	if job.Spec.Completions != nil {
		requestedCompletions = *job.Spec.Completions
	}
	return job.Status.Succeeded == requestedCompletions
}

// FilterJobsWithOneCompletion returns jobs that have one completion
func FilterJobsWithOneCompletion(jobList []batchv1.Job) []batchv1.Job {
	var result []batchv1.Job
	for _, job := range jobList {
		if JobHasOneCompletion(job) {
			result = append(result, job)
		}
	}
	return result
}

// CountJobsWithOneCompletion count the number complete jobs
func CountJobsWithOneCompletion(jobList []batchv1.Job) int {
	result := 0

	for _, job := range jobList {
		if JobHasOneCompletion(job) {
			result++
		}
	}

	return result
}
