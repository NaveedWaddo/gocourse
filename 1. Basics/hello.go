package main

import "fmt"

func main() {
	var num1,num2 int

	fmt.Print("Enter First Value: ")
	fmt.Scan(&num1)

	fmt.Print("Enter Second Value: ")
	fmt.Scan(&num2)

	total,err := add(num1 , num2)

	if err != nil {
		fmt.Println("Error" , err)
	}
	fmt.Println("Addition is: ", total)
}

func add(num1  , num2 int ) (int , error) { 
	if num1 <= 0 || num2 <= 0 {
		return 0, fmt.Errorf("Numbers must be greater than 0!")
	}

	return num1+num2 , nil
}