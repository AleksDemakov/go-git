package main

import (
	"encoding/hex"
	"os"
	"path/filepath"

	//"fmt"
	"crypto/sha1"
)

type commit struct {
	hash    [20]byte
	author  string
	email   string
	message string
}

// a link struct between commits and commited files
type commit_objects struct {
	commit_hash [20]byte
	objects     []byte
}

func new_commit_objects(commit_hash [20]byte) {
	gogit := SearchGit()

    // read Index for staged files to include in the commit
	index_bytes, err := os.ReadFile(filepath.Join(gogit, "index"))
	if err != nil {
		panic(err)
	}

	commit_objects_instance := commit_objects{commit_hash, index_bytes}

    f, err := os.Create(filepath.Join(gogit, "commits-objects", hex.EncodeToString(commit_hash[:])))
    defer f.Close()
    if err != nil {
       panic(err)
    }
    f.Write([]byte(commit_objects_instance.objects))
}

func Commit(message string) {
	commit_hash := sha1.Sum([]byte(message))
	commit := commit{commit_hash, "author", "email", message}

	gogit := SearchGit()
	f, err := os.Create(filepath.Join(gogit, "commits", hex.EncodeToString(commit.hash[:]))) 
	if err != nil {
		panic(err)
	}
	f.Write([]byte(commit.message))
	f.Close()

	new_commit_objects(commit_hash)
}
