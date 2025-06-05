package writecounter

import "fmt"

func Example() {
	counter := new(WriteCounter)
	if _, err := counter.Write([]byte("123")); err != nil {
		// Handle the error
		return
	}
	fmt.Println(counter.Len())
	// Output: 3
}
