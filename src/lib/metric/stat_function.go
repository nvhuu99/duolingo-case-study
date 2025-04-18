package metric

import (
	"sort"

	"golang.org/x/exp/constraints"
)

func mean[T constraints.Integer | constraints.Float](arr []T) float64 {
	var sum float64
	for _, v := range arr {
		sum += float64(v)
	}
	return sum / float64(len(arr))
}

func sum[T constraints.Integer | constraints.Float](arr []T) float64 {
	var sum float64
	for _, v := range arr {
		sum += float64(v)
	}
	return sum
}

func percentile[T constraints.Integer | constraints.Float](arr []T, p float64) float64 {
	sort.Slice(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})
	index := int(p * float64(len(arr)-1))
	return float64(arr[index])
}

func min[T constraints.Integer | constraints.Float](arr []T) float64 {
	minValue := arr[0]
	for _, v := range arr {
		if v < minValue {
			minValue = v
		}
	}
	return float64(minValue)
}

func max[T constraints.Integer | constraints.Float](arr []T) float64 {
	maxValue := arr[0]
	for _, v := range arr {
		if v > maxValue {
			maxValue = v
		}
	}
	return float64(maxValue)
}
