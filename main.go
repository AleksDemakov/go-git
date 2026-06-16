package main


import (
	"fmt"
    "os"
	//"path/filepath"
)

func ReadArg(command string, NameFile string) {
	if command == "add" {
        Add(NameFile)
		
	} else if command == "commit" {
		fmt.Println(2)
	} else if command == "push" {
		fmt.Println(3)
	} else {
		fmt.Println("Failed command")
	}
}

func main() {

	command := os.Args[1]

	NameFile := os.Args[2]

	ReadArg(command, NameFile)

}
