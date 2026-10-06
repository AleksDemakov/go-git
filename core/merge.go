package core

import (
	"fmt"
)



// Merge performs a merge operation between branches using the specified strategy
// Valid strategies: "ours", "theirs", "recursive" (default)
func Merge(theirsBranch string, strategy string) {
	ctx := newMergeContext(theirsBranch, strategy)

	// Validation
	if err := ctx.Validate(); err != nil {
		fmt.Println(err)
		return
	}

	// Initialization - find commits and load objects
	if err := ctx.Init(); err != nil {
		fmt.Printf("Failed to initialize merge context: %v\n", err)
		return
	}

	// Strategy handling
	if strategy == "ours" || strategy == "theirs" {
		ctx.HandleTrivialStrategy()
		return
	}

	// Conflict analysis for recursive strategy
	ctx.AnalyzeConflicts()

	// Conflict resolution if there any
	ctx.ResolveConflicts()

	if ctx.ConflictsResolved {
		return
	}

	// Handle conflict output
	if ctx.HasConflicts {
		for _, msg := range ctx.Messages {
			fmt.Println(msg)
		}
		fmt.Println("Automatic merge failed; fix conflicts and then commit.")
	}

	// Apply final result to working tree, index, and commit
	ctx.ApplyResult()
}

func AbortMerge () {
	if !adapter.IsConflictState() {
		fmt.Println("There is no merge to abort")
		return
	}
	// doing before checkout in order to pass conflicted merge validation
	curHash := adapter.GracefullyAbortMerge()
	Checkout(curHash)
	
}
