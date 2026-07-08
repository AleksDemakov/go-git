package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func SearchGit() string {
	absPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(absPath, ".go-git")); os.IsNotExist(err) {
			dir := filepath.Dir(absPath)
			if dir == absPath {
				fmt.Println("file not exist")
				panic(fmt.Sprintln("there is no go-git repository"))
			}
			absPath = dir
		} else {
			return filepath.Join(absPath, ".go-git")
		}
	}
}
func ReadArg(command string, NameFile string) {
	if command == "add" {
		Add(NameFile)

	} else if command == "commit" {
        Commit(NameFile)
	} else if command == "push" {
		fmt.Println(3)
	} else if command == "init" {
		dirs := []string{".go-git", ".go-git/commits", ".go-git/commits-objects", ".go-git/objects"}
        files := []string{".go-git/index", ".go-git/config"}

		for _, name := range dirs {
            if err := os.Mkdir(name, 0777); err != nil {
				panic(err) // Останавливает программу, если ошибка есть
			}
		}
        for _, name := range files {
            if _, err := os.Create(name); err != nil {
                panic(err)
            }
        }
	} else {
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
