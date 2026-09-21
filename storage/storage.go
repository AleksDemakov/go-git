package storage

import (
    types "github.com/Clorter/go-git/types"
)

type Storage interface{
    Init() 
    Add(types.Object)
    AddDeletedFile(string)
    Get() types.Object
    Commit(string)
	Index() []types.Object
	Branch() ([]string, string)
	NewBranch(string)
	DeleteBranch(string)
	SearchRepoMeta() string
	Restore(bool, string, string) types.Object
	GetObjectsFromLastCommits(int) []types.Object
	GetObjectsFromCommit(string) []types.Object
	FindHashesForMerge(string) map[string]string
	Log(int) []types.Commit
	MoveHeadTo(string)
	MoveIndexTo(string)
	Merge(string, string, []types.Object, []types.ConflictFile)
	isConflictState() bool
	Rm(string)

}

