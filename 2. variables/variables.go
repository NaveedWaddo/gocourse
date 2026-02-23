package main

import (
	"fmt"
	_ "math"
)


var last = "Waddo"

func main() {

	 city := "Mumbai"
	 pin := 413004

	 var num1 int = 3
	 var num2 = float32(num1)

	 const daysInWeek  = 7

	 var d float64 = daysInWeek

	 fmt.Println(num2)
	 fmt.Println(daysInWeek)


	city , pin , country  := "Solapur" , 41305 , "india"

	fmt.Printf("city : %s | Pin : %d | Country : %s" , city , pin , country)

}
