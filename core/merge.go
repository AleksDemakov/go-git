package core 

import (
	"fmt"
	"os"
    types "github.com/Clorter/go-git/types"
    //storage "github.com/Clorter/go-git/storage"
)


func Merge(theirsBranch string, strategy string) {
	if adapter.IsConflictState() {
		fmt.Println("Conflict State. Resolve conflict first")
		return
	} 
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
		adapter.Merge(theirsBranch, strategy, []types.Object{}, []types.ConflictFile{})
		return
	}

	// =========== common strategy (recursive) =============
	var mergeResultObjects []types.Object
	var	conflictedObjects  []types.ConflictFile

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
				//fmt.Println("file not changed")
				baseObj.Flag = "CE_VALID"
				mergeResultObjects = append(mergeResultObjects , baseObj)
			} else if baseObj.Hash == oursObj.Hash && baseObj.Hash != theirsObj.Hash {
				// file is changed in only in theirs
				//fmt.Println("file is changed in only in theirs")
				theirsObj.Flag = "CE_VALID"
				mergeResultObjects = append(mergeResultObjects, theirsObj)
			} else if baseObj.Hash == theirsObj.Hash && baseObj.Hash != oursObj.Hash {
				// file is changed only in ours
				//fmt.Println("file is changed only in ours")
				oursObj.Flag = "CE_VALID"
				mergeResultObjects = append(mergeResultObjects, oursObj)
			} else {
				// file is changed in both commits. need to compare line by line
				//fmt.Println("file is changed in both. need to compare line by line")
				baseObj.Flag = "CE_CONFLICTED_0"
				oursObj.Flag = "CE_CONFLICTED_1"
				theirsObj.Flag = "CE_CONFLICTED_2"
				conflictedObjects = append(conflictedObjects, 
					types.ConflictFile{baseObj, oursObj, theirsObj, types.BothModified})
				conflictFlag = true 
			}
			continue
		}

		// file is only in theirs. Deleted in ours
		if idInTheirs != -1 && idInOurs == -1 {
			theirsObj := theirsObjects[idInTheirs]
			if baseObj.Hash != theirsObj.Hash {
				baseObj.Flag = "CE_CONFLICTED_0"
				theirsObj.Flag = "CE_CONFLICTED_2"
				//fmt.Println("file is changed in theirs but deleted in ours. Conflict")
				conflictedObjects = append(conflictedObjects, 
					types.ConflictFile{baseObj, theirsObj, theirsObj, types.OurDeleted})
				conflictFlag = true // file is changed in theirs but deleted in ours. Conflict
			}
			// file not changed but deleted in ours. Then delete
			continue
		}

		// file is only in ours. Deleted in theirs 
		if idInTheirs == -1 && idInOurs != -1 {
			oursObj := oursObjects[idInOurs]
			if baseObj.Hash != oursObj.Hash {
				baseObj.Flag = "CE_CONFLICTED_0"
				oursObj.Flag = "CE_CONFLICTED"
				//fmt.Println("file is changed in ours but deleted in theris. Conflict")
				conflictedObjects = append(conflictedObjects, types.ConflictFile{baseObj, oursObj, oursObj, types.TheirDeleted})
				conflictFlag = true // file is changed in theirs but deleted in ours. Conflict
			}
			// file not changed but deleted in ours. Then delete
			continue
		}
		// file is deleted in both 
		//fmt.Println("file is deleted in both ")
		if idInTheirs == -1 && idInOurs == -1 {
			continue
		}
	}

	// traverse files in theirs commit but only those are not in base
	for _, theirObj := range theirsObjects {
		idInBase := findObjectInSliceByPath(theirObj, baseObjects)
		idInOurs := findObjectInSliceByPath(theirObj, oursObjects)

		// file is created in theirs
		if idInBase == -1 {
			if idInOurs != -1 {
				// file is created in both commits 
				oursObj := oursObjects[idInOurs]
				// different content. Conflict
				if theirObj.Hash != oursObj.Hash {
					oursObj.Flag = "CE_CONFLICTED"
					theirObj.Flag = "CE_CONFLICTED"
					conflictedObjects = append(conflictedObjects, 
					types.ConflictFile{
						baseObjects[idInBase],
						oursObj, 
						theirObj, types.BothModified})
					conflictFlag = true 
				} else {
					mergeResultObjects = append(mergeResultObjects, theirObj) 
				}
			} else {
				// file created only in theirs
				mergeResultObjects = append(mergeResultObjects, theirObj)
			}
		}
	}

	// traverse files in ours commit but only those are not in base
	for _, ourObj := range oursObjects {
		idInBase := findObjectInSliceByPath(ourObj, baseObjects)
		idInTheirs := findObjectInSliceByPath(ourObj, theirsObjects)

		// file is created in ours 
		if idInBase == -1 {
			if idInTheirs == -1 {
				mergeResultObjects = append(mergeResultObjects, ourObj)
			}
			// If file is in theirs, it already proccessed in previous loop
		}
	}

	workingTreeTransformer := workingTreeTransform{}
	var objectsWithMarkers []types.Object
	if conflictFlag {
		if adapter.IsConflictResolved(conflictedObjects) {
			adapter.ContinueMerge(theirsBranch)	
			return
		}
	
		// Prepare files for conflict resolving with git-style markers
		// add conflict files in working tree 
		for _, conflictObject := range conflictedObjects {
			if conflictObject.Conflict == types.BothModified {
				// TODO: if changes do not intersect then there is no confilct
				// Generate conflict content with git-style markers
				oursContent := conflictObject.Ours.Content
				theirsContent := conflictObject.Theirs.Content
				conflictContent := GenerateDiff(
					oursContent,
					theirsContent,
					GitConflictFormat,
					"HEAD",
					theirsBranch,
				)
				// Create new object with conflict content
				conflictObj := types.NewObjectFromContent(
					conflictObject.Ours.Path,
					conflictContent,
				)
				conflictObj.Flag = "CE_CONFLICTED"
				//mergeResultObjects = append(mergeResultObjects, conflictObj)
				objectsWithMarkers = append(objectsWithMarkers, conflictObj)
				fmt.Println("CONFLICT (content): Merge conflict in", conflictObject.Ours.Path)
			}
			if conflictObject.Conflict == types.TheirDeleted {
				mergeResultObjects = append(mergeResultObjects, conflictObject.Ours)
				fmt.Println("CONFLICT (delete/modify): ",  conflictObject.Ours.Path," deleted in theirs and modified in ours." )
			}
			if conflictObject.Conflict == types.OurDeleted {
				mergeResultObjects = append(mergeResultObjects, conflictObject.Theirs)
				fmt.Println("CONFLICT (modify/delete): ", conflictObject.Ours.Path, " deleted in ours and modified in theirs." )
			}
		}
		fmt.Println("Automatic merge failed; fix conflicts and then commit.")

	} else {
		//workingTreeTransformer.Apply(workingDir, untrackedObjects, newObjects)
	}
	workingTreeTransformer.Apply(workingDir, untrackedObjects, append(objectsWithMarkers, mergeResultObjects...))
	adapter.Merge(theirsBranch, "recursive", mergeResultObjects, conflictedObjects)


	// TODO: catch move, rename.
	// Дополнительно: Проверка на перемещения/переименования (если нужно)
	// Здесь можно добавить логику для отслеживания перемещений,
	// сравнивая содержимое файлов и их пути.

}
