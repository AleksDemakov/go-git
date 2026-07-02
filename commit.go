package main

import (
	"encoding/hex"
	"os"
	"path/filepath"

	//"fmt"
	"crypto/sha256"
)

type commit struct {
	hash    [32]byte
	author  string
	email   string
	message string
}

type commit_objects struct {
	commit_hash [32]byte
	objects     []byte
}

func new_commit_objects(commit_hash [32]byte) {
	gogit := SearchGit()

	index_bytes, err := os.ReadFile(filepath.Join(gogit, "index"))
	if err != nil {
		panic(err)
	}

	commit_objects_instance := commit_objects{commit_hash, index_bytes}

	f, err := os.Create(filepath.Join(gogit, "commits-objects", hex.EncodeToString(commit_hash[:])))
	if err != nil {
		panic(err)
	}
	f.Write([]byte(commit_objects_instance.objects))
	f.Close()

}

func Commit(message string) {
	commit_hash := sha256.Sum256([]byte(message))
	commit := commit{commit_hash, "author", "email", message}

	gogit := SearchGit()
	f, err := os.Create(filepath.Join(gogit, "commits", hex.EncodeToString(commit.hash[:]))) // hex.EncodeToString(obj.hash[:]))
	if err != nil {
		panic(err)
	}
	f.Write([]byte(commit.message))
	f.Close()

	new_commit_objects(commit_hash)
}
