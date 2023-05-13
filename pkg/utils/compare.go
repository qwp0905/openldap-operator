package utils

import (
	"reflect"

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

func ComparePVC(vc1, vc2 corev1.PersistentVolumeClaim) bool {
	if vc1.Spec.StorageClassName != nil && vc2.Spec.StorageClassName != nil {
		if *vc1.Spec.StorageClassName != *vc2.Spec.StorageClassName {
			return false
		}
	}

	if !reflect.DeepEqual(vc1.Spec.Resources, vc2.Spec.Resources) {
		return false
	}

	if !reflect.DeepEqual(vc1.Spec.AccessModes, vc2.Spec.AccessModes) {
		return false
	}

	if vc1.Spec.VolumeName != vc2.Spec.VolumeName {
		return false
	}

	if vc1.Spec.VolumeMode != nil && vc2.Spec.VolumeMode != nil {
		if !reflect.DeepEqual(*vc1.Spec.VolumeMode, *vc2.Spec.VolumeMode) {
			return false
		}
	}

	if vc1.Spec.Selector != nil && vc2.Spec.Selector != nil {
		if !reflect.DeepEqual(*vc1.Spec.Selector, *vc2.Spec.Selector) {
			return false
		}
	}

	return true
}
