package main

import (
	"fmt"
	"regexp"
)

func main() {
	re := regexp.MustCompile(`a.`)
	fmt.Println(re.FindAllString("paranormal", -1)) // [ar an al]
	fmt.Println(re.FindAllString("paranormal", 2))  // [ar an]
	fmt.Println(re.FindAllString("graal", -1))      // [aa]
	fmt.Println(re.FindAllString("none", -1))       // []
}
