package iter_test

import (
	"fmt"
	"github.com/qiafan666/gotato/commons/gconc/iter"
)

func ExampleIterator() {
	input := []int{1, 2, 3, 4}
	iterator := iter.Iterator[int]{
		MaxGoroutines: len(input) / 2,
	}

	iterator.ForEach(input, func(v *int) {
		if *v%2 != 0 {
			*v = -1
		}
	})

	fmt.Println(input)
	// Output:
	// [-1 2 -1 4]
}

func ExampleMapper() {
	input := []int{1, 2, 3, 4}
	mapper := iter.Mapper[int, bool]{
		MaxGoroutines: len(input) / 2,
	}

	results := mapper.Map(input, func(v *int) bool { return *v%2 == 0 })
	fmt.Println(results)
	// Output:
	// [false true false true]
}
