package helper

func ReverseSlice[T any](arr []T) {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
}

func InArray[T comparable](value T, arr []T) bool {
	for _, arrItem := range arr {
		if value == arrItem {
			return true
		}
	}

	return false
}