package main

import (
	"crypto/sha1"
	"encoding/hex"
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
    obj := types.Object{sha1.Sum([]byte(content)), filepath_, 42, string(content)}
	return &obj
}

func Add(filepath_ string) {
	workingDir, _ := os.Getwd()
	absPath := filepath.Join(workingDir, filepath_)

	// check if file exists
	if _, err := os.Stat(absPath); err != nil {
		panic(fmt.Sprintf("file %s does not exist\n", absPath))
	}
	// save "object" of file in .go-git/Objects
	gogit := SearchGit()
	obj := newObject(absPath)
	objects_file, _ := os.Create(filepath.Join(gogit, "objects", hex.EncodeToString(obj.Hash[:]))) // hex.EncodeToString(obj.hash[:])  instead of obj.hash
	defer objects_file.Close()
	objects_file.Write([]byte(obj.Content))

	index_file, err := os.OpenFile(filepath.Join(gogit, "index"), os.O_WRONLY|os.O_APPEND, 0644)
    defer index_file.Close()
	if err != nil {
		panic(err)
	}
	index_file.WriteString(hex.EncodeToString(obj.Hash[:]) + "\n") 

    // database test
    //storage.Add(*obj)
    //obj1 := storage.Get()
    //fmt.Println(obj1)


}
