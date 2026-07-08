package main

import (
	"fmt"
	"os"
	"path/filepath"

    types "github.com/Clorter/go-git/types"
    storage "github.com/Clorter/go-git/storage"
)


func newObject(filepath_ string) *types.Object {
	content, err := os.ReadFile(filepath_)
	if err != nil {
		panic(err)
	}
	obj := types.Object{"42", filepath_, 42, string(content)}
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
	gogit := SearchGit()
	fmt.Printf("gogit=%s\n", gogit)
	obj := newObject(absPath)
	f, err := os.Create(filepath.Join(gogit, "objects", obj.Hash))
	if err != nil {
		panic(err)
	}
	f.Write([]byte(obj.Content))
	f.Close()

}
