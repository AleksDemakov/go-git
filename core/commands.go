package core 

import (
	"fmt"
	"os"
	"path/filepath"

    types "github.com/Clorter/go-git/types"
    storage "github.com/Clorter/go-git/storage"
)

// TODO: read from config
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

// workingTreeObjects walks over all directories recursively collecting files 
func getWorkingTreeObjects( dir string) []types.Object {
	var objects []types.Object
	workingTreeObjectsRecursion(&objects, dir)
	return objects
}

func workingTreeObjectsRecursion(objects *[]types.Object, dir string) *[]types.Object{
    entries, err := os.ReadDir(dir)
    if err != nil {
        panic(err) 
    }
    for _, e := range entries {
		if e.Name()[0] == '.'{
			continue
		}	
		filename := filepath.Join(dir, e.Name())
		if e.IsDir() {
			workingTreeObjectsRecursion(objects, filename)
		} else { 
			*objects = append(*objects, types.NewObject(filename))
		}
    }
	return objects
}

func findObjectInSilce(obj types.Object, slice []types.Object) int {
	for index, entry := range(slice) {
		if obj.Path == entry.Path {
			return index
		}
	}
	return  -1
}

func Status() {
	index := adapter.Status() // list of types.Object
	if len(index) == 0 {
		fmt.Println("Nothing to commit")
		return 
	} 
	repoPath := filepath.Dir(adapter.SearchRepoMeta())
	workingTreeObjects := getWorkingTreeObjects(repoPath)

	fmt.Println("Changes to be committed:")
	fmt.Println("  Not implemented -> (use \"git restore --staged <file>...\" to unstage)")
	workingDir, _ := os.Getwd()
	for _, file := range index {
		fileRelPath, _ := filepath.Rel(workingDir, file.Path)
		fmt.Println("    " + fileRelPath)
	}


	var objectsNotStaged []types.Object
	for _, object := range index{
		if id := findObjectInSilce(object, workingTreeObjects); id !=  -1{
			if object.Hash != workingTreeObjects[id].Hash {
				objectsNotStaged = append(objectsNotStaged, workingTreeObjects[id])
			}
		}
	}

	if len(objectsNotStaged) >0 {
		fmt.Println()
		fmt.Println("Changes not stagged for commit:")
		fmt.Println("  (use \"git add <file>\" to update what will be committed)")	
		fmt.Println("  Not implemented -> (use \"git restore <file>...\" to discard changes in working directory)")	
		for _, file := range objectsNotStaged{
			fileRelPath, _ := filepath.Rel(workingDir, file.Path)
			fmt.Println("    " + fileRelPath)
		}
	}

	var objectsNotInIndex []types.Object
	for _, object := range workingTreeObjects{
		if id := findObjectInSilce(object, index); id ==  -1{
			objectsNotInIndex = append(objectsNotInIndex, object)
		}
	}

	if len(objectsNotInIndex) >0 {
		fmt.Println()
		fmt.Println("Untracked files:")
		fmt.Println("   (use \"git add <file>...\" to include in what will be committed)")	
		for _, file := range objectsNotInIndex{
			fileRelPath, _ := filepath.Rel(workingDir, file.Path)
			fmt.Println("    " + fileRelPath)
		}
	}
}

func Branch() {
	branches, curBranch := adapter.Branch()
	for _, branch := range(branches){
		if curBranch == branch{
			fmt.Println("* " + branch)
		}else{
			fmt.Println("  " + branch)
		}
	}
}
