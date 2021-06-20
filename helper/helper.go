package helper

func InArray(in int, arr []int) bool {
	for i := range arr {
		if arr[i] == in {
			return true
		}
	}

	return false
}

func ExtractKeysFromMap(m map[string]struct{}) []string {
	keys := make([]string, len(m))
	index := 0
	for node := range m {
		keys[index] = node
		index++
	}

	return keys
}
