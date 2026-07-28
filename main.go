package main

import (
	"fmt"
	"os"
	//"path/filepath"
    storage "github.com/Clorter/go-git/storage"
)

var writer storage.Storage


func ReadArg(command string, arg string) {
	if command == "add" {
		Add(arg) // arg is a file name
	} else if command == "commit" {
        Commit(arg) // arg is a commit message
	} else if command == "push" {
		fmt.Println("push is not implemented yet")
	} else if command == "init" {
        writer.Init()
	} else {
        // TODO: print help here
		fmt.Println("Failed command")
	}
}

func main() {

    writer = storage.Database_storage{}

	command := os.Args[1] 

    var NameFile string
    if len(os.Args) > 2 {
        NameFile = os.Args[2]
    }

	ReadArg(command, NameFile)

}
