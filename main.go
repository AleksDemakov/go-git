package main

import (
	"fmt"
	"os"
	"flag"
	//"path/filepath"
	core "github.com/Clorter/go-git/core"
)



func printHelp() {
	fmt.Println(`go-git - A simple git implementation

Usage: go-git [--help] <command> [<args>]

Commands:
  init                    Initialize a git repository
  add <file>             Stage a file for commit
  commit <message>       Commit staged changes
  push                   Push commits to remote (not implemented)
  status                 Show working tree status
  checkout <branch>      Switch branches
  branch [-d <name>]     List branches, create new branch, or delete with -d
  merge [-s <strategy>]  Merge a branch (strategy optional)
  diff <file1> <file2> [output]  Generate conflict-style diff between files
  restore [--staged] <file>  Restore file from HEAD

Options:
  --help, -h             Show this help message`)
}

func ReadArg(command string, arg string) {
	if command == "add" {
		core.Add(arg) // arg is a file name
	} else if command == "commit" {
        core.Commit(arg) // arg is a commit message
	} else if command == "push" {
		fmt.Println("push is not implemented yet")
	} else if command == "init" {
		//TODO: validation
        core.Init()
	} else if command == "status"{
		core.Status()
	} else if command == "checkout" {
		core.Checkout(arg)
	} else {
        // TODO: print help here
		fmt.Println("Failed command")
	}
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	if os.Args[1] == "--help" || os.Args[1] == "-h" {
		printHelp()
		os.Exit(0)
	}

	restoreCmd := flag.NewFlagSet("restore", flag.ExitOnError)
	stagedFlag := restoreCmd.Bool("staged" , false, "where to restore")

	branchCmd := flag.NewFlagSet("branch", flag.ExitOnError)
	deleteBranchFlag := branchCmd.String("d", "", "name of branch to delete")

	mergeCmd := flag.NewFlagSet("merge", flag.ExitOnError)
	mergeStrategyFlag := mergeCmd.String("s", "", "merge strategy")

	command := os.Args[1] 
	switch command {
	//TODO: file validation
	case "restore" :
		restoreCmd.Parse(os.Args[2:])
		core.Restore(*stagedFlag, restoreCmd.Args()[0])
	case "branch":
		branchCmd.Parse(os.Args[2:])
		if *deleteBranchFlag == "" {
			// new branch
			if len(branchCmd.Args()) > 0 {
				core.Branch(branchCmd.Args()[0], false)
			} else {
				core.Branch("", false)
			}
		} else {
			// delete branch
			core.Branch(*deleteBranchFlag, true)
		}
	case "merge":
		mergeCmd.Parse(os.Args[2:])
		core.Merge(mergeCmd.Args()[0], *mergeStrategyFlag)
	case "diff":
		if len(os.Args) < 4 {
			fmt.Println("Usage: go-git diff <file1> <file2> [output]")
			os.Exit(1)
		}
		var outputPath string
		if len(os.Args) >= 5 {
			outputPath = os.Args[4]
		}
		diffOutput, err := core.Diff(os.Args[2], os.Args[3], outputPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if outputPath == "" {
			fmt.Print(diffOutput)
		}
	default: 
		var NameFile string
		if len(os.Args) > 2 {
			NameFile = os.Args[2]
		}
		ReadArg(command, NameFile)
	}

}
