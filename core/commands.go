package core 

import (
	"fmt"
	"os"
	"path/filepath"

    types "github.com/Clorter/go-git/types"
    storage "github.com/Clorter/go-git/storage"
)

var adapter storage.Storage = storage.Database_storage{}

func Init(){
	adapter.Init()
}

func Commit(message string) {
    adapter.Commit(message)
}

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
    adapter.Add(obj)
}

func Status() {
	files := adapter.Status()
	if len(files) == 0 {
		fmt.Println("Nothing to commit")
	} else {
		fmt.Println("Changes staged for commit:")
		workingDir, _ := os.Getwd()
		for _, file := range files {
			fileRelativePath, _ := filepath.Rel(workingDir, file.Path)
			fmt.Println(fileRelativePath)
		}
	}
 	// TODO: add untracked but changed files listing
}

//func Branch() {
	//branches := adapter.Branch()
	//fmt.Println(branches)
//}
