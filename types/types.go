package types

import (
	"os"
	"crypto/sha1"
)

type Commit struct {
	Hash    [20]byte
	ParentHash [20]byte
	Parent2Hash [20]byte
	Author  string
	Email   string
	Message string
	Date string
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
	Flag 	string
}

type ConflictType int
const ( 
	BothModified = iota
	OurDeleted
	TheirDeleted
)

type ConflictFile struct {
	Base 	 Object
    Ours 	 Object
    Theirs 	 Object
	Conflict ConflictType
}

func NewObject(filepath_ string) Object {
	content, err := os.ReadFile(filepath_)
	if err != nil {
		panic(err)
	}
    obj := Object{sha1.Sum([]byte(content)), filepath_, 42, string(content), "CE_VALID"}
	return obj
}

// NewObjectFromContent creates an Object from content string directly.
// Useful for generating conflict files and other synthetic objects.
func NewObjectFromContent(path string, content string) Object {
	hash := sha1.Sum([]byte(content))
	return Object{
		Hash:    hash,
		Path:    path,
		Size:    int32(len(content)),
		Content: content,
		Flag:    "CE_VALID",
	}
}
