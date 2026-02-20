package main

import (
	"fmt"
)

type Abser interface{
	Abs() float64
}
func main() {
	var a Abser
	


	fmt.Println(a.Abs())

}

type Myfloat float64