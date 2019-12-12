package bitacoin

func inArray(in int, arr []int) bool {
	for i := range arr {
		if arr[i] == in {
			return true
		}
	}

	return false
}
