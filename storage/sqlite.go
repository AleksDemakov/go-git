package storage 
import (
        "database/sql"
        "os"
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

func (ds Database_storage) Get() types.Object{
    db, err := sql.Open("sqlite", "sqlite.db")
    if err != nil {
        panic(err)
    }
    defer db.Close()
    rows, err := db.Query("select hash, path, size, content from objects;")
    if err != nil {
            panic(err)
    }

    for rows.Next() {
        var obj types.Object
        var test string
        var hash string
        if err = rows.Scan(&hash, &obj.Path, &test, &obj.Content); err != nil {
            panic(err)
        }

       return obj
    }

    if err = rows.Err(); err != nil {
            panic(err)
    }
    hash, _ := hex.DecodeString("da39a3ee5e6b4b0d3255bfef95601890afd80709")
    return types.Object{[20]byte(hash) , "da", 64.0, "da"} 
}

func (ds Database_storage) Add(obj types.Object){
    db, err := sql.Open("sqlite", "sqlite.db")
    if err != nil {
        panic(err)
    }
    defer db.Close()

    //fmt.Println(obj.Hash)
    if _, err = db.Exec(`
    insert into objects (hash, path, size, content) values (?, ?, ?, ?);
    `, hex.EncodeToString(obj.Hash[:]), obj.Path, obj.Size, obj.Content); err != nil {
        panic(err)
    }
}


