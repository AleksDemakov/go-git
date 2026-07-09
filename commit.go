package main

import (
	"encoding/hex"
	"os"
	"path/filepath"

	//"fmt"
	"crypto/sha1"
    types "github.com/Clorter/go-git/types"
    storage "github.com/Clorter/go-git/storage"
)



func new_commit_objects(commit_hash [20]byte) {
	gogit := storage.SearchGit()

    // read Index for staged files to include in the commit
	index_bytes, err := os.ReadFile(filepath.Join(gogit, "index"))
	if err != nil {
		panic(err)
	}

	commit_objects_instance := types.Commit_objects{commit_hash, index_bytes}

    f, err := os.Create(filepath.Join(gogit, "commits-objects", hex.EncodeToString(commit_hash[:])))
    defer f.Close()
    if err != nil {
       panic(err)
    }
    f.Write([]byte(commit_objects_instance.Objects))
}

func Commit(message string) {
	commit_hash := sha1.Sum([]byte(message))
	commit := types.Commit{commit_hash, "author", "email", message}

	gogit := storage.SearchGit()
	f, err := os.Create(filepath.Join(gogit, "commits", hex.EncodeToString(commit.Hash[:]))) 
	if err != nil {
		panic(err)
	}
	f.Write([]byte(commit.Message))
	f.Close()

	new_commit_objects(commit_hash)
}
