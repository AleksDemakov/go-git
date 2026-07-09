package main 

import (
	"crypto/sha1"
	//"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

    types "github.com/Clorter/go-git/types"
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
    // ====== An object instance creation ======
	workingDir, _ := os.Getwd()
	absPath := filepath.Join(workingDir, filepath_)

	// check if file exists
	if _, err := os.Stat(absPath); err != nil {
		panic(fmt.Sprintf("file %s does not exist\n", absPath))
	}
	obj := newObject(absPath)

    // ====== Store the object ================
    writer.Add(*obj)
    obj1 := writer.Get()
    fmt.Println(obj1)
}
