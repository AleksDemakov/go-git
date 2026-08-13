package storage

import (
    types "github.com/Clorter/go-git/types"
)

type Storage interface{
    Init() 
    Add(types.Object)
    Get() types.Object
    Commit(string)
	Status() []types.Object
	Branch() ([]string, string)
	SearchRepoMeta() string
}

