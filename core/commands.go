package core 

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

    types "github.com/Clorter/go-git/types"
    storage "github.com/Clorter/go-git/storage"
)

// TODO: read from config
// TODO: use types.AppConfig.AdapterType to initialize the appropriate adapter
var adapter storage.Storage = storage.Database_storage{}



func Init(){
	adapter.Init()
}

func Commit(message string) {
// TODO: check IsConflictResolved and ContinueMerge. Because classic git require commit after conflicted merge
    adapter.Commit(message)
}

// TODO: add Rename type of change (move is equal rename)
func Add(filepath_ string) {
    // ====== An object instance creation ======
	workingDir, _ := os.Getwd()
	absPath := filepath.Join(workingDir, filepath_)
	

	// check if file exists
	if _, err := os.Stat(absPath); err != nil {
		// if file was in index than file is removed
		for _, object := range adapter.Index() {
			if object.Path == filepath_ {
				adapter.AddDeletedFile(filepath_)
				return 
			}
		}
		panic(fmt.Sprintf("file %s does not exist\n", absPath))
	}
	obj := types.NewObject(absPath)

    // ====== Store the object ================
    adapter.Add(obj)
}

func Log(limit int, branch string){
	commits := adapter.Log(limit, branch)
	for _, commit := range commits {
		fmt.Println(commit.RichString())
	}
}

func Rm(cacheFlag bool, filepath_ string) {
	// Check if file exists in working directory or index
	workingDir, _ := os.Getwd()
	absPath := filepath.Join(workingDir, filepath_)
	
	// Check if file exists in working directory
	if _, err := os.Stat(absPath); err != nil {
		// Check if file is in index
		fileInIndex := false
		for _, object := range adapter.Index() {
			if object.Path == filepath_ {
				fileInIndex = true
				break
			}
		}
		if !fileInIndex {
			panic(fmt.Sprintf("pathspec '%s' did not match any files", filepath_))
		}
	}
	
	adapter.Rm(cacheFlag, filepath_)
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
		if e.IsDir() &&  e.Name()[0] == '.' {
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

// TODO: abstract these 3 functions
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
func findObjectInSliceByPathSuffix(path string, slice []types.Object) int {
	for index, entry := range(slice) {
		if strings.HasSuffix(entry.Path, path) {
			return index
		}
	}
	return  -1
}

func getUntrackedObjects() []types.Object {
	index := adapter.Index() // list of types.Object
	repoPath := filepath.Dir(adapter.SearchRepoMeta())
	workingTreeObjects := getWorkingTreeObjects(repoPath)
	var objectsNotInIndex []types.Object
	for _, object := range workingTreeObjects{
		if id := findObjectInSliceByPath(object, index); id ==  -1{
			objectsNotInIndex = append(objectsNotInIndex, object)
		}
	}
	return objectsNotInIndex
}

// TODO: print: 
//Unmerged paths:
//  (use "git add/rm <file>..." as appropriate to mark resolution)
//
//    both modified:   file1.txt
//    both modified:   file2.txt
//    deleted by us:   file3.txt
//    deleted by them: file4.txt
func Status() {
	
	if adapter.IsConflictState() {
		adapter.PrintUnmergedPaths()
		return 
	}
	
	_, curBranch := adapter.Branch()
	fmt.Println("On branch ", curBranch)

	index := adapter.Index() // list of types.Object
	if len(index) == 0 {
		fmt.Println("Nothing to commit")
		return 
	} 
	lastCommitObjects := adapter.GetObjectsFromLastCommits(1) 
	repoPath := filepath.Dir(adapter.SearchRepoMeta())
	workingTreeObjects := getWorkingTreeObjects(repoPath)


	var tobeCommited []string 
	for _, object := range index {
		// if object from index is not in las commit
		if id := findObjectInSliceByHash(object, lastCommitObjects); id ==  -1{
			tobeCommited = append(tobeCommited, object.Path)
		}
	}
	
	if len(tobeCommited) > 0 {
		fmt.Println("Changes to be committed:")
		fmt.Println("  Not implemented -> (use \"git restore --staged <file>...\" to unstage)")
		for _, path := range tobeCommited{
			fmt.Println("    " + path)
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
		fmt.Println("  (use \"git restore <file>...\" to discard changes in working directory)")	
		for _, object := range objectsNotStaged{
			fmt.Println("    " + object.Path)
		}
	}

	objectsNotInIndex := getUntrackedObjects()
	if len(objectsNotInIndex) > 0 {
		fmt.Println()
		fmt.Println("Untracked files:")
		fmt.Println("   (use \"git add <file>...\" to include in what will be committed)")	
		for _, file := range objectsNotInIndex{
			fmt.Println("    " + file.Path)
		}
	}
}

// TODO: name equals commit hash case. also in table head should be possible to use commit hash as a name 
// TODO: working tree validation for unsaved items 
func Checkout(name string) {
	adapter.MoveHeadTo(name)
	adapter.MoveIndexTo(name)

	workingDir, _ := os.Getwd()
	objects := adapter.GetObjectsFromLastCommits(1)	
	untrackedObjects := getUntrackedObjects()
	workingTreeTransformer := workingTreeTransform{}
	workingTreeTransformer.Apply(workingDir, untrackedObjects, objects)
}

func Branch(name string, isDelete bool) {
	if isDelete && name != "main"{
		adapter.DeleteBranch(name)
	}
	if !isDelete && name != "" {
		adapter.NewBranch(name)
	}	
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

// Diff generates a conflict-style diff between two files using LCS algorithm.
// Lines only in file1 are prefixed with '+', lines only in file2 are prefixed with '-',
// and common lines (from LCS) are written unchanged.
// If outputPath is empty, the diff is returned as a string; otherwise written to file.
func Diff(file1Path string, file2Path string, outputPath string) (string, error) {
	// Read file contents
	content1, err := os.ReadFile(file1Path)
	if err != nil {
		return "", err
	}
	content2, err := os.ReadFile(file2Path)
	if err != nil {
		return "", err
	}

	// Generate diff content using SimpleDiffFormat (+/- prefixes)
	diffContent := GenerateDiff(string(content1), string(content2), SimpleDiffFormat, "", "")

	if outputPath == "" {
		return diffContent, nil
	}

	// Write to output file
	return "", os.WriteFile(outputPath, []byte(diffContent), 0644)
}

