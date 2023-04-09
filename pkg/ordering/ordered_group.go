package ordering

import "sort"

type OrderedItem[T any] struct {
	Order int
	Value T
}

type OrderedItems[T any] []OrderedItem[T]

func (o OrderedItems[T]) Get() []T {
	sort.Slice(o, func(i, j int) bool {
		return o[i].Order < o[j].Order
	})

	result := make([]T, len(o))
	for i, item := range o {
		result[i] = item.Value
	}

	return result
}
