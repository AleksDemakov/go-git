package main

import (
	"fmt"
	"os"
	"path/filepath"
    "math/rand"
    "time"
)

type object struct {
	hash    string
	path    string
	size    int32
	content string
}

func randomString(length int) string {
    rand.Seed(time.Now().UnixNano())
    const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    result := make([]byte, length)

    for i:=0;i<length;i++ {
        result[i] = charset[rand.Intn(len(charset))]
    }
    return string(result)
}

func newObject(filepath_ string) *object {
	content, err := os.ReadFile(filepath_)
	if err != nil {
		panic(err)
	}
	obj := object{randomString(8), filepath_, 42, string(content)}
	return &obj
}

func Add(filepath_ string) {
	workingDir, _ := os.Getwd()
	absPath := filepath.Join(workingDir, filepath_)

	// check if file exists
	if _, err := os.Stat(absPath); err != nil {
		panic(fmt.Sprintf("file %s does not exist\n", absPath))
	}
	// save "object" of file in .go-git/Objects
	gogit := SearchGit()
	obj := newObject(absPath)
	objects_file, _ := os.Create(filepath.Join(gogit, "objects", obj.hash))
    defer objects_file.Close()
	objects_file.Write([]byte(obj.content))

    index_file, err := os.OpenFile(filepath.Join(gogit, "index"), os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil{
        panic(err)
    }
    index_file.WriteString(obj.hash)

}
