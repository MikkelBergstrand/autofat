package structure

func FlattenArray[T any](input []T) []T {
	ret := make([]T, 0)

	current := input

	for len(current) == 2 {
		ret = append(ret, current[0])
	}
	return nil
}
