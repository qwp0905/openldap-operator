package utils

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
