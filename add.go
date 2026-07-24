package main 

import (
	//"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

    types "github.com/Clorter/go-git/types"
)


func Add(filepath_ string) {
    // ====== An object instance creation ======
	workingDir, _ := os.Getwd()
	absPath := filepath.Join(workingDir, filepath_)

	// check if file exists
	if _, err := os.Stat(absPath); err != nil {
		panic(fmt.Sprintf("file %s does not exist\n", absPath))
	}
	obj := types.NewObject(absPath)

    // ====== Store the object ================
    writer.Add(obj)
    obj1 := writer.Get()
    fmt.Println(obj1)
}
