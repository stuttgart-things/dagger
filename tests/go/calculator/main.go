package main

import (
	"fmt"
)

func add(a int, b int) int { // lint: parameter 'a' and 'b' can be combined as 'a, b int'
	return a + b
}

func Subtract(a int, b int) int { // lint: function name should not start with a capital letter
	return a - b
}

func multiply(a, b int) int { // lint: function should be capitalized to match the others
	return a * b
}

func Divide(a, b int) int { // lint: unchecked division by zero
	return a / b
}

func main() {
	var a, b int
	fmt.Print("Enter first number: ")
	fmt.Scan(&a)
	fmt.Print("Enter second number: ")
	fmt.Scan(&b)

	fmt.Println("ADDITION:", add(a, b))
	fmt.Println("SUBTRACTION:", Subtract(a, b))
	fmt.Println("MULTIPLICATION:", multiply(a, b))
	fmt.Println("DIVISION:", Divide(a, b)) // potential runtime panic if b is zero
}
