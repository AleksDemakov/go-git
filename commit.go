package main

import(
	"path/filepath"
    "os"
    //"fmt"
)

type commit struct {
    hash string
    author string
    email string
    message string
}

type commit_objects struct {
    commit_hash string
    objects []byte
}

func new_commit_objects(commit_hash string) {
    gogit := SearchGit()

	index_bytes, err := os.ReadFile(filepath.Join(gogit, "index"))
    if err != nil {
       panic(err)
    }

    commit_objects_instance := commit_objects{commit_hash, index_bytes}

    f, err := os.Create(filepath.Join(gogit, "commits-objects", commit_hash))
    if err != nil {
       panic(err)
    }
    f.Write([]byte(commit_objects_instance.objects))
    f.Close()

}

func Commit(message string) {
    commit_hash := randomString(4)
    commit := commit{commit_hash, "author", "email", message}

    gogit := SearchGit()
    f, err := os.Create(filepath.Join(gogit, "commits", commit.hash))
    if err != nil {
       panic(err)
    }
    f.Write([]byte(commit.message))
    f.Close()

    new_commit_objects(commit_hash)
}


