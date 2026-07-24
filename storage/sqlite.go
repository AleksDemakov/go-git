package storage 
import (
        "database/sql"
        "os"
		"strconv"
        //"fmt"
        //"path/filepath"
        _ "modernc.org/sqlite"
        "encoding/hex"

        types "github.com/Clorter/go-git/types"
)

type Database_storage struct{
}

func (ds Database_storage) Init(){
    if _, err := os.Create("sqlite.db"); err!=nil {
        panic(err)
    } 
    db, err := sql.Open("sqlite", "sqlite.db")
    if err != nil {
        panic(err)
    }
    defer db.Close()
    if _, err = db.Exec(`
create table objects(
    hash blob primary key, 
    path text, 
    size real, 
    content text
    ) without rowid;
create table commit_(
    hash blob primary key, 
    parent_hash blob, 
    author text, 
    email text
    ) without rowid;
create table commit_objects(
    commit_hash blob, 
    object_hash blob
    );
    `) ; err != nil {
        panic(err)
    }
}

func (ds Database_storage) FilesAll() []types.Object{
    db, err := sql.Open("sqlite", "sqlite.db")
    if err != nil {
        panic(err)
    }
    defer db.Close()
    rows, err := db.Query("select hash, path, size, content from objects;")
    if err != nil {
		panic(err)
    }

	var files []types.Object
    for rows.Next() {
        var obj types.Object
        var size_str string
        var hash_str string
        if err = rows.Scan(&hash_str, &obj.Path, &size_str, &obj.Content); err != nil {
            panic(err)
        }
		hash_slice, _ := hex.DecodeString(hash_str)
		copy(obj.Hash[:], hash_slice)
		size_64, _ := strconv.ParseInt(size_str, 10, 32)
		obj.Size = int32(size_64)
		files = append(files, obj)
    }
	return files
}

func (ds Database_storage) Add(obj types.Object){
    db, err := sql.Open("sqlite", "sqlite.db")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    //fmt.Printf("sqlite.go:84: %s", hex.EncodeToString(obj.Hash[:]))
    if _, err = db.Exec(`
    insert into objects (hash, path, size, content) values (?, ?, ?, ?);
    `, hex.EncodeToString(obj.Hash[:]), obj.Path, obj.Size, obj.Content); err != nil {
        panic(err)
    }
}


