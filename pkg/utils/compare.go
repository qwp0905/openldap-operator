package utils

import (
	corev1 "k8s.io/api/core/v1"
)

func CompareLabels(exists map[string]string, new map[string]string) bool {
	for key, val := range new {
		if exists[key] != val {
			return false
		}
	}

	return true
}

func ConvertBool(flag bool) string {
	if flag {
		return "yes"
	} else {
		return "no"
	}
}

func CompareEnv(c1, c2 corev1.Container) bool {
	for _, de := range c2.Env {
		be := Find(c1.Env, func(e corev1.EnvVar) bool {
			return e.Name == de.Name
		})

		if be == nil {
			return false
		}

		if be.Value != de.Value {
			return false
		}
	}

	return true
}
