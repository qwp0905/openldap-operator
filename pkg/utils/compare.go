package utils

import (
	"reflect"
	"sort"

	corev1 "k8s.io/api/core/v1"
)

func CompareMap(exists map[string]string, new map[string]string) bool {
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
	e1 := c1.DeepCopy().Env
	e2 := c2.DeepCopy().Env

	sort.Slice(e1, func(i, j int) bool {
		return e1[i].Name > e1[j].Name
	})
	sort.Slice(e2, func(i, j int) bool {
		return e2[i].Name > e2[j].Name
	})

	return reflect.DeepEqual(e1, e2)
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

func CompareServicePorts(p1, p2 []corev1.ServicePort) bool {
	sort.Slice(p1, func(i, j int) bool {
		return p1[i].Port > p1[j].Port
	})
	sort.Slice(p2, func(i, j int) bool {
		return p2[i].Port > p2[j].Port
	})

	return reflect.DeepEqual(p1, p2)
}
