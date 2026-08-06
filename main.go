package main

import (
	"fmt"
	"os"
	//"path/filepath"
	core "github.com/Clorter/go-git/core"
)



func ReadArg(command string, arg string) {
	fmt.Println("MAIN:", command )
	if command == "add" {
		core.Add(arg) // arg is a file name
	} else if command == "commit" {
        core.Commit(arg) // arg is a commit message
	} else if command == "push" {
		fmt.Println("push is not implemented yet")
	} else if command == "init" {
        core.Init()
	} else if command == "status"{
		core.Status()
	}else{
        // TODO: print help here
		fmt.Println("Failed command")
	}
}

func main() {

	command := os.Args[1] 
    var NameFile string
    if len(os.Args) > 2 {
        NameFile = os.Args[2]
    }
	ReadArg(command, NameFile)

}
