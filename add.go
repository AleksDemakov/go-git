package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type object struct {
	hash    string
	path    string
	size    int32
	content string
}

func newObject(filepath_ string) *object {
	content, err := os.ReadFile(filepath_)
	if err != nil {
		panic(err)
	}
	obj := object{"42", filepath_, 42, string(content)}
	return &obj
}

func Add(filepath_ string) {
	workingDir, _ := os.Getwd()
	absPath := filepath.Join(workingDir, filepath_)

	// check if file exists
	if _, err := os.Stat(absPath); err != nil {
		panic(fmt.Sprintf("file %s does not exist\n", absPath))
	}
	// save "object" of file to .go-git
	gogit := findLocalRepo(absPath)
	fmt.Printf("gogit=%s\n", gogit)
	obj := newObject(absPath)
	f, err := os.Create(filepath.Join(gogit, "objects", obj.hash))
	if err != nil {
		panic(err)
	}
	f.Write([]byte(obj.content))
	f.Close()
}
