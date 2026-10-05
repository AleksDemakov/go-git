package core

import (
	"fmt"
	"os"
	"path/filepath"

	types "github.com/Clorter/go-git/types"
)

// MergeContext encapsulates all state and data needed for a merge operation
type MergeContext struct {
	// Input parameters
	TheirsBranch string
	Strategy     string

	// Working directory
	WorkingDir string

	// Commit hashes
	BaseHash   string
	TheirsHash string
	OursHash   string

	// Objects from commits
	BaseObjects     []types.Object
	TheirsObjects   []types.Object
	OursObjects     []types.Object
	UntrackedObjects []types.Object

	// Results
	MergeResultObjects []types.Object
	ConflictObjects    []types.ConflictFile

	// State
	HasConflicts       bool
	ConflictsResolved  bool
	Error              error
	Messages           []string
}

// newMergeContext creates a new merge context with basic initialization
func newMergeContext(theirsBranch, strategy string) *MergeContext {
	workingDir, _ := os.Getwd()
	return &MergeContext{
		TheirsBranch: theirsBranch,
		Strategy:     strategy,
		WorkingDir:   workingDir,
		Messages:     make([]string, 0),
	}
}

// NewMergeContextForTest creates a new merge context for testing purposes
// This is exported so tests can access it
func NewMergeContextForTest(theirsBranch, strategy string) *MergeContext {
	return newMergeContext(theirsBranch, strategy)
}

// Init performs the initial setup: finds commit hashes and loads objects
func (ctx *MergeContext) Init() error {
	// Get commit hashes for merge
	hashes := adapter.FindCommitsForMerge(ctx.TheirsBranch)
	ctx.BaseHash = hashes["base"]
	ctx.TheirsHash = hashes["theirs"]
	ctx.OursHash = hashes["ours"]

	// Load objects from all commits
	ctx.BaseObjects = adapter.GetObjectsFromCommit(ctx.BaseHash)
	ctx.TheirsObjects = adapter.GetObjectsFromCommit(ctx.TheirsHash)
	ctx.OursObjects = adapter.GetObjectsFromCommit(ctx.OursHash)

	// Get untracked objects from working directory
	repoPath := filepath.Dir(adapter.SearchRepoMeta())
	ctx.UntrackedObjects = getWorkingTreeObjects(repoPath)

	return nil
}

// validate checks if merge can proceed
// Validate checks if merge can proceed
func (ctx *MergeContext) Validate() error {
	if adapter.IsConflictState() {
		return fmt.Errorf("Conflict State. Resolve conflict first")
	}
	return nil
}

// handleTrivialStrategy handles "ours" and "theirs" merge strategies
// HandleTrivialStrategy handles "ours" and "theirs" merge strategies
func (ctx *MergeContext) HandleTrivialStrategy() {
	workingTreeTransformer := workingTreeTransform{}
	if ctx.Strategy == "ours" {
		workingTreeTransformer.Apply(ctx.WorkingDir, ctx.UntrackedObjects, ctx.OursObjects)
	} else {
		workingTreeTransformer.Apply(ctx.WorkingDir, ctx.UntrackedObjects, ctx.TheirsObjects)
	}
	adapter.Merge(ctx.TheirsBranch, ctx.Strategy, []types.Object{}, []types.ConflictFile{})
}

// analyzeConflicts performs conflict analysis between the three sets of objects
// AnalyzeConflicts performs conflict analysis between the three sets of objects
func (ctx *MergeContext) AnalyzeConflicts() {
	// Reset state
	ctx.MergeResultObjects = make([]types.Object, 0)
	ctx.ConflictObjects = make([]types.ConflictFile, 0)
	for _, baseObj := range ctx.BaseObjects {
		idInTheirs := findObjectInSliceByPath(baseObj, ctx.TheirsObjects)
		idInOurs := findObjectInSliceByPath(baseObj, ctx.OursObjects)

		// file in both target commits
		if idInTheirs != -1 && idInOurs != -1 {
			theirsObj := ctx.TheirsObjects[idInTheirs]
			oursObj := ctx.OursObjects[idInOurs]

			// file not changed
			if baseObj.Hash == theirsObj.Hash && baseObj.Hash == oursObj.Hash {
				baseObj.Flag = "CE_VALID"
				ctx.MergeResultObjects = append(ctx.MergeResultObjects, baseObj)
			} else if baseObj.Hash == oursObj.Hash && baseObj.Hash != theirsObj.Hash {
				// file is changed in only in theirs
				theirsObj.Flag = "CE_VALID"
				ctx.MergeResultObjects = append(ctx.MergeResultObjects, theirsObj)
			} else if baseObj.Hash == theirsObj.Hash && baseObj.Hash != oursObj.Hash {
				// file is changed only in ours
				oursObj.Flag = "CE_VALID"
				ctx.MergeResultObjects = append(ctx.MergeResultObjects, oursObj)
			} else {
				// file is changed in both commits. need to compare line by line
				baseObj.Flag = "CE_CONFLICTED_0"
				oursObj.Flag = "CE_CONFLICTED_1"
				theirsObj.Flag = "CE_CONFLICTED_2"
				ctx.ConflictObjects = append(ctx.ConflictObjects,
					types.ConflictFile{baseObj, oursObj, theirsObj, types.BothModified})
			}
			continue
		}

		// file is only in theirs. Deleted in ours
		if idInTheirs != -1 && idInOurs == -1 {
			theirsObj := ctx.TheirsObjects[idInTheirs]
			if baseObj.Hash != theirsObj.Hash {
				baseObj.Flag = "CE_CONFLICTED_0"
				theirsObj.Flag = "CE_CONFLICTED_2"
				ctx.ConflictObjects = append(ctx.ConflictObjects,
					types.ConflictFile{baseObj, theirsObj, theirsObj, types.OurDeleted})
			}
			// file not changed but deleted in ours. Then delete
			continue
		}

		// file is only in ours. Deleted in theirs 
		if idInTheirs == -1 && idInOurs != -1 {
			oursObj := ctx.OursObjects[idInOurs]
			if baseObj.Hash != oursObj.Hash {
				baseObj.Flag = "CE_CONFLICTED_0"
				oursObj.Flag = "CE_CONFLICTED"
				ctx.ConflictObjects = append(ctx.ConflictObjects, types.ConflictFile{baseObj, oursObj, oursObj, types.TheirDeleted})
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
	for _, theirObj := range ctx.TheirsObjects {
		idInBase := findObjectInSliceByPath(theirObj, ctx.BaseObjects)
		idInOurs := findObjectInSliceByPath(theirObj, ctx.OursObjects)

		// file is created in theirs
		if idInBase == -1 {
			if idInOurs != -1 {
				// file is created in both commits 
				oursObj := ctx.OursObjects[idInOurs]
				// different content. Conflict
				if theirObj.Hash != oursObj.Hash {
					oursObj.Flag = "CE_CONFLICTED"
					theirObj.Flag = "CE_CONFLICTED"
					ctx.ConflictObjects = append(ctx.ConflictObjects,
						types.ConflictFile{
							ctx.BaseObjects[idInBase],
							oursObj, 
							theirObj, types.BothModified})
				} else {
					ctx.MergeResultObjects = append(ctx.MergeResultObjects, theirObj)
				}
			} else {
				// file created only in theirs
				ctx.MergeResultObjects = append(ctx.MergeResultObjects, theirObj)
			}
		}
	}

	// traverse files in ours commit but only those are not in base
	for _, ourObj := range ctx.OursObjects {
		idInBase := findObjectInSliceByPath(ourObj, ctx.BaseObjects)
		idInTheirs := findObjectInSliceByPath(ourObj, ctx.TheirsObjects)

		// file is created in ours 
		if idInBase == -1 {
			if idInTheirs == -1 {
				ctx.MergeResultObjects = append(ctx.MergeResultObjects, ourObj)
			}
			// If file is in theirs, it already processed in previous loop
		}
	}
	
	// Set conflict state based on results
	ctx.HasConflicts = len(ctx.ConflictObjects) > 0
}

// resolveConflicts handles conflict resolution with git markers
// ResolveConflicts handles conflict resolution with git markers
func (ctx *MergeContext) ResolveConflicts() {
	// Clear messages
	ctx.Messages = make([]string, 0)
	
	if len(ctx.ConflictObjects) == 0 {
		return
	}

	if adapter.IsConflictResolved(ctx.ConflictObjects) {
		adapter.ContinueMerge(ctx.TheirsBranch)
		ctx.ConflictsResolved = true
		return
	}

	// Prepare files for conflict resolving with git-style markers
	for _, conflictObject := range ctx.ConflictObjects {
		if conflictObject.Conflict == types.BothModified {
			// Generate conflict content with git-style markers
			oursContent := conflictObject.Ours.Content
			theirsContent := conflictObject.Theirs.Content
			conflictContent := GenerateDiff(
				oursContent,
				theirsContent,
				GitConflictFormat,
				"HEAD",
				ctx.TheirsBranch,
			)
			// Create new object with conflict content
			conflictObj := types.NewObjectFromContent(
				conflictObject.Ours.Path,
				conflictContent,
			)
			conflictObj.Flag = "CE_CONFLICTED"
			ctx.MergeResultObjects = append(ctx.MergeResultObjects, conflictObj)
			ctx.Messages = append(ctx.Messages, fmt.Sprintf("CONFLICT (content): Merge conflict in %s", conflictObject.Ours.Path))
		}
		if conflictObject.Conflict == types.TheirDeleted {
			ctx.MergeResultObjects = append(ctx.MergeResultObjects, conflictObject.Ours)
			ctx.Messages = append(ctx.Messages, fmt.Sprintf("CONFLICT (delete/modify):  %s deleted in theirs and modified in ours.", conflictObject.Ours.Path))
		}
		if conflictObject.Conflict == types.OurDeleted {
			ctx.MergeResultObjects = append(ctx.MergeResultObjects, conflictObject.Theirs)
			ctx.Messages = append(ctx.Messages, fmt.Sprintf("CONFLICT (modify/delete):  %s deleted in ours and modified in theirs.", conflictObject.Ours.Path))
		}
	}
}

// ApplyResult applies the merge result to working tree and database
func (ctx *MergeContext) ApplyResult() {
	workingTreeTransformer := workingTreeTransform{}
	var objectsWithMarkers []types.Object
	
	// For conflicted objects, create conflict markers in working tree
	for _, conflictObject := range ctx.ConflictObjects {
		if conflictObject.Conflict == types.BothModified {
			// Generate conflict content with git-style markers
			oursContent := conflictObject.Ours.Content
			theirsContent := conflictObject.Theirs.Content
			conflictContent := GenerateDiff(
				oursContent,
				theirsContent,
				GitConflictFormat,
				"HEAD",
				ctx.TheirsBranch,
			)
			// Create new object with conflict content
			conflictObj := types.NewObjectFromContent(
				conflictObject.Ours.Path,
				conflictContent,
			)
			conflictObj.Flag = "CE_CONFLICTED"
			objectsWithMarkers = append(objectsWithMarkers, conflictObj)
		}
	}
	
	workingTreeTransformer.Apply(
		ctx.WorkingDir,
		ctx.UntrackedObjects,
		append(objectsWithMarkers, ctx.MergeResultObjects...),
	)
	adapter.Merge(
		ctx.TheirsBranch,
		ctx.Strategy,
		ctx.MergeResultObjects,
		ctx.ConflictObjects,
	)
}
