package core 

import (
	"fmt"
	"os"
	"path/filepath"
	//"strings"

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
	for id, object := range(objects) {
		relPath, _ := filepath.Rel(dir, object.Path)
		objects[id].Path = relPath
	}
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

func findObjectInSliceByPath(obj types.Object, slice []types.Object) int {
	for index, entry := range(slice) {
		if obj.Path == entry.Path {
			return index
		}
	}
	return  -1
}
func findObjectInSliceByHash(obj types.Object, slice []types.Object) int {
	for index, entry := range(slice) {
		if obj.Hash == entry.Hash{
			return index
		}
	}
	return  -1
}

// TODO: refactor and write comments
func Status() {
	
	_, curBranch := adapter.Branch()
	fmt.Println("On branch ", curBranch)

	index := adapter.Status() // list of types.Object
	if len(index) == 0 {
		fmt.Println("Nothing to commit")
		return 
	} 
	lastCommitObjects := adapter.GetObjectsFromCommits(1) 
	repoPath := filepath.Dir(adapter.SearchRepoMeta())
	workingTreeObjects := getWorkingTreeObjects(repoPath)


	fmt.Println("Changes to be committed:")
	fmt.Println("  Not implemented -> (use \"git restore --staged <file>...\" to unstage)")
	for _, object := range index {
		if id := findObjectInSliceByHash(object, lastCommitObjects); id ==  -1{
			fmt.Println("    " + object.Path)
		}
	}


	var objectsNotStaged []types.Object
	for _, object := range index{
		if id := findObjectInSliceByPath(object, workingTreeObjects); id !=  -1{
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
		for _, object := range objectsNotStaged{
			fmt.Println("    " + object.Path)
		}
	}

	var objectsNotInIndex []types.Object
	for _, object := range workingTreeObjects{
		if id := findObjectInSliceByPath(object, index); id ==  -1{
			objectsNotInIndex = append(objectsNotInIndex, object)
		}
	}

	if len(objectsNotInIndex) >0 {
		fmt.Println()
		fmt.Println("Untracked files:")
		fmt.Println("   (use \"git add <file>...\" to include in what will be committed)")	
		for _, file := range objectsNotInIndex{
			fmt.Println("    " + file.Path)
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

func Restore(staged bool, filename string){
	object := adapter.Restore(staged, "HEAD", filename) 
	if !staged {
		file, err := os.Create(filename)	
		if err != nil {panic(err)}
		defer file.Close()
		//content := strings.TrimRight(object.Content, "\n")
		_, err = file.WriteString(object.Content)
	}
	
}
