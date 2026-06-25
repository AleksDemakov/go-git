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
		fmt.Println(2)
	} else if command == "push" {
		fmt.Println(3)
	} else if command == "init" {
		dir := []string{".go-git", ".go-git/Config", ".go-git/Commits", ".go-git/CommitsObjects", ".go-git/Objects"}

		for _, name := range dir {
			err := os.Mkdir(name, 0777) // создает директороию, если ее не существует
			if err != nil {
				panic(err) // Останавливает программу, если ошибка есть
			}
		}
	} else {
		fmt.Println("Faild command")
	}
}

func main() {

	command := os.Args[1] //Только ему всегда надо указывать файл како-либо похоже

	NameFile := os.Args[2]

	ReadArg(command, NameFile)
	SearchGit()

}
