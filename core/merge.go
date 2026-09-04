package core 

import (
	"fmt"
	"os"
    types "github.com/Clorter/go-git/types"
    //storage "github.com/Clorter/go-git/storage"
)


func Merge(theirsBranch string, strategy string) {
	workingDir, _ := os.Getwd()
	untrackedObjects := getUntrackedObjects()
	hashes := adapter.FindHashesForMerge(theirsBranch)
	baseHash := hashes["base"]
	theirsHash := hashes["theirs"]
	oursHash := hashes["ours"]

	conflictFlag := false
	theirsObjects := adapter.GetObjectsFromCommit(theirsHash)
	oursObjects := adapter.GetObjectsFromCommit(oursHash) 
	baseObjects := adapter.GetObjectsFromCommit(baseHash)

	if strategy == "ours" || strategy == "theirs" {
		workingTreeTransformer := workingTreeTransform{}
		if strategy == "ours"{
			workingTreeTransformer.Apply(workingDir, untrackedObjects, oursObjects)
		}else{
			workingTreeTransformer.Apply(workingDir, untrackedObjects, theirsObjects)
		}
		adapter.Merge(theirsBranch, strategy, []types.Object{})
		return
	}

	// =========== common strategy (recursive) =============
	var newObjects []types.Object

	// traverse files in base commit
	for _, baseObj := range baseObjects {
		idInTheirs := findObjectInSliceByPath(baseObj, theirsObjects)
		idInOurs := findObjectInSliceByPath(baseObj, oursObjects)

		// file in both target commits
		if idInTheirs != -1 && idInOurs != -1 {
			theirsObj := theirsObjects[idInTheirs]
			oursObj := oursObjects[idInOurs]

			// file not changed
			if baseObj.Hash == theirsObj.Hash && baseObj.Hash == oursObj.Hash {
				newObjects = append(newObjects, baseObj)
			} else if baseObj.Hash == oursObj.Hash && baseObj.Hash != theirsObj.Hash {
				// file is changed in only in theirs
				newObjects = append(newObjects, theirsObj)
			} else if baseObj.Hash == theirsObj.Hash && baseObj.Hash != oursObj.Hash {
				// file is changed in only in ours
				newObjects = append(newObjects, oursObj)
			} else {
				// file is changed in both commits. To compare line by line
				conflictFlag = true 
			}
			continue
		}

		// file is only in theirs. Deleted in ours
		if idInTheirs != -1 && idInOurs == -1 {
			theirsObj := theirsObjects[idInTheirs]
			if baseObj.Hash != theirsObj.Hash {
				conflictFlag = true // file is changed in theirs but deleted in ours. Conflict
			}
			// file not changed but deleted in ours. Then delete
			continue
		}

		// Файл есть только в ours (удален в theirs)
		if idInTheirs == -1 && idInOurs != -1 {
			oursObj := oursObjects[idInOurs]
			if baseObj.Hash != oursObj.Hash {
				conflictFlag = true // file is changed in theirs but deleted in ours. Conflict
			}
			// file not changed but deleted in ours. Then delete
			continue
		}

		// file is deleted in both 
		if idInTheirs == -1 && idInOurs == -1 {
			continue
		}
	}

	// traverse files in theirs commit but only those are not in base
	for _, obj := range theirsObjects {
		idInBase := findObjectInSliceByPath(obj, baseObjects)
		idInOurs := findObjectInSliceByPath(obj, oursObjects)

		// file is created in theirs
		if idInBase == -1 {
			if idInOurs != -1 {
				// file is created in both commits 
				oursObj := oursObjects[idInOurs]
				// different content. Conflict
				if obj.Hash != oursObj.Hash {
					conflictFlag = true 
				} else {
					newObjects = append(newObjects, obj) 
				}
			} else {
				// file created only in theirs
				newObjects = append(newObjects, obj)
			}
		}
	}

	// traverse files in ours commit but only those are not in base
	for _, obj := range oursObjects {
		idInBase := findObjectInSliceByPath(obj, baseObjects)
		idInTheirs := findObjectInSliceByPath(obj, theirsObjects)

		// file is created in ours 
		if idInBase == -1 {
			if idInTheirs == -1 {
				newObjects = append(newObjects, obj)
			}
			// If file is in theirs, it already proccessed in previous loop
		}
	}

	if conflictFlag {
		fmt.Println("Conflict")
	} else {
		workingTreeTransformer := workingTreeTransform{}
		workingTreeTransformer.Apply(workingDir, untrackedObjects, newObjects)
		adapter.Merge(theirsBranch, "common", newObjects)
	}

	// TODO: catch move, rename.
	// Дополнительно: Проверка на перемещения/переименования (если нужно)
	// Здесь можно добавить логику для отслеживания перемещений,
	// сравнивая содержимое файлов и их пути.

}
