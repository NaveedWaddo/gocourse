package main

import "fmt"


type Person struct {
		Name string
		Last string
		age int
	}



func (p Person) printFullName() string {
		return "Full Name is  " + p.Name +" " + p.Last;
	}

func main() {

	

	p := Person {
		Name: "Naveed",
		Last: "Waddo",
		age: 22,
	}

	fmt.Println(p)

	fmt.Println("First Name: ",p.Name)

	fmt.Println(p.printFullName())
}
