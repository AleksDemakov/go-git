package storage

import (
    "os"
	"encoding/hex"
    "path/filepath"
    "fmt"
	"crypto/sha1"

    types "github.com/Clorter/go-git/types"
)

type Disk_storage struct{}

// SearchRepoMeta returns absolute path to meta file or directory including file or dir itself
func (ds Disk_storage) SearchRepoMeta() string {
	absPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(absPath, ".go-git")); os.IsNotExist(err) {
			dir := filepath.Dir(absPath)
			if dir == absPath {
				fmt.Println("file not exist")
				panic(fmt.Sprintln("there is no go-git repository"))
			}
			absPath = dir
		} else {
			return filepath.Join(absPath, ".go-git")
		}
	}
}

func (ds Disk_storage) Init(){
		dirs := []string{".go-git", ".go-git/commits", ".go-git/commits-objects", ".go-git/objects"}
        files := []string{".go-git/index", ".go-git/config"}

		for _, name := range dirs {
            if err := os.Mkdir(name, 0777); err != nil {
				panic(err) // Останавливает программу, если ошибка есть
			}
		}
        for _, name := range files {
            if _, err := os.Create(name); err != nil {
                panic(err)
            }
        }
}

func (ds Disk_storage) Add(obj types.Object){
    dotgogit := ds.SearchRepoMeta()
	objects_file, err := os.Create(filepath.Join(dotgogit, "objects", hex.EncodeToString(obj.Hash[:]))) // hex.EncodeToString(obj.hash[:])  instead of obj.hash
	if err != nil {
		panic(err)
	} 
    defer objects_file.Close()
    
	objects_file.Write([]byte(obj.Content))

	index_file, err := os.OpenFile(filepath.Join(dotgogit, "index"), os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
    defer index_file.Close()
    
	index_file.WriteString(hex.EncodeToString(obj.Hash[:]) + "\n") 
}

func (ds Disk_storage) Get() types.Object{
    hash, _ := hex.DecodeString("da39a3ee5e6b4b0d3255bfef95601890afd80709")
    return types.Object{[20]byte(hash) , "da", 64.0, "da"} 
}

func (ds Disk_storage) new_commit_objects(commit_hash [20]byte) {
	gogit := ds.SearchRepoMeta()

    // read Index for staged files to include in the commit
	index_bytes, err := os.ReadFile(filepath.Join(gogit, "index"))
	if err != nil {
		panic(err)
	}

	commit_objects_instance := types.Commit_objects{commit_hash, index_bytes}

    f, err := os.Create(filepath.Join(gogit, "commits-objects", hex.EncodeToString(commit_hash[:])))
    if err != nil {
       panic(err)
    }
    defer f.Close()
    f.Write([]byte(commit_objects_instance.Objects))
}

func (ds Disk_storage) Commit(message string){
	commit_hash := sha1.Sum([]byte(message))
	// TODO: insert parentHash as second argument! 
	commit := types.Commit{commit_hash, commit_hash, "author", "email", message, "01-01-2026 00:00:00"}
	ds.new_commit_objects(commit_hash)

    dotgogit := ds.SearchRepoMeta()
	f, err := os.Create(filepath.Join(dotgogit, "commits", hex.EncodeToString(commit.Hash[:]))) 
	if err != nil {
		panic(err)
	}
	f.Write([]byte(commit.Message))
	f.Close()
}

func Status(){}

func Branch()([] string, string){return []string{}, ""}
