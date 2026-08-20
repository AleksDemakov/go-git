package main

import (
	"fmt"
	"os"
	"flag"
	//"path/filepath"
	core "github.com/Clorter/go-git/core"
)



func ReadArg(command string, arg string) {
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
	} else if command == "branch" {
		core.Branch()
	} else{
        // TODO: print help here
		fmt.Println("Failed command")
	}
}

func main() {
	restoreCmd := flag.NewFlagSet("restore", flag.ExitOnError)
	stagedFlag := restoreCmd.Bool("staged" , false, "where to restore")

	command := os.Args[1] 
	switch command {
	//TODO: file validation
	case "restore" :
		restoreCmd.Parse(os.Args[2:])
		fmt.Println(*stagedFlag)
		fmt.Println(restoreCmd.Args())
		core.Restore(*stagedFlag, restoreCmd.Args()[0])
	default: 
		var NameFile string
		if len(os.Args) > 2 {
			NameFile = os.Args[2]
		}
		ReadArg(command, NameFile)
	}

}
