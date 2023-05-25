package goutil

import (
	"io"
)

type Foo struct {
	Bar string
	Baz int
}

func ExampleXXDump() {
	foo := Foo{
		Bar: "Hello, World!",
		Baz: 123,
	}

	Dump(foo)
	// Output:
	//(goutil.Foo) {
	//  Bar: (string) (len=13) "Hello, World!",
	//  Baz: (int) 123
	//}

	Sdump(foo)
	// Output:
	//(goutil.Foo) {
	//  Bar: (string) (len=13) "Hello, World!",
	//  Baz: (int) 123
	//}

	Fdump(io.Discard, foo)
}
