package storage

import (
    types "github.com/Clorter/go-git/types"
)

type Storage interface{
    Init() 
    Add(types.Object)
    Get() types.Object
}

