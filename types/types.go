package types

import (
	"os"
	"crypto/sha1"
)

type Commit struct {
	Hash    [20]byte
	Author  string
	Email   string
	Message string
}

// a link struct between commits and commited files
type Commit_objects struct {
	Commit_hash [20]byte
	Objects     []byte
}

type Object struct {
	Hash    [20]byte 
	Path    string
	Size    int32
	Content string
}

func NewObject(filepath_ string) Object {
	content, err := os.ReadFile(filepath_)
	if err != nil {
		panic(err)
	}
    obj := Object{sha1.Sum([]byte(content)), filepath_, 42, string(content)}
	return obj
}
