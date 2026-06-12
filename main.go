package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func ReadArg(command string, NameFile string) {
	if command == "add" {
		// fmt.Println(1)
		SelfPathFile, Myerror := os.Executable()
		PathFile, Nameerror := filepath.Abs(NameFile)

		fmt.Println(SelfPathFile)
		if Myerror != nil || Nameerror != nil {
			fmt.Println(Myerror)
			fmt.Println(Nameerror)
		} else {

			RelPath, error := filepath.Rel(SelfPathFile, PathFile)
			if error != nil {
				fmt.Println(error)
			} else {
				fmt.Println(RelPath)
			}
		}
	} else if command == "commit" {
		fmt.Println(2)
	} else if command == "push" {
		fmt.Println(3)
	} else {
		fmt.Println("Faild command")
	}
}

func main() {

	command := os.Args[1]

	NameFile := os.Args[2]

	ReadArg(command, NameFile)

}
