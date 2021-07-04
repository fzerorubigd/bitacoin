package helper

func ExtractKeysFromMap(m map[string]struct{}) []string {
	keys := make([]string, len(m))
	index := 0
	for node := range m {
		keys[index] = node
		index++
	}

	return keys
}
