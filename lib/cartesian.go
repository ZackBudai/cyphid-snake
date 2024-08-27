package lib

// CartesianProduct takes slices of any type T and returns a channel receiving cartesian products
func CartesianProduct[T any](params ...[]T) chan []T {
	// create channel
	c := make(chan []T)
	if len(params) == 0 {
		close(c)
		return c // Return a safe value for nil/empty params.
	}
	go func() {
		iterate(c, params[0], []T{}, params[1:]...)
		close(c)
	}()
	return c
}

func iterate[T any](channel chan []T, topLevel, result []T, needUnpacking ...[]T) {
	if len(needUnpacking) == 0 {
		for _, p := range topLevel {
			newResult := make([]T, len(result), len(result)+1)
			copy(newResult, result)
			channel <- append(newResult, p)
		}
		return
	}
	for _, p := range topLevel {
		newResult := make([]T, len(result), len(result)+1)
		copy(newResult, result)
		iterate(channel, needUnpacking[0], append(newResult, p), needUnpacking[1:]...)
	}
}
