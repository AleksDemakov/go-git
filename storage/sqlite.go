package storage 
import (
        "database/sql"
        "os"
		"strconv"
        //"fmt"
        //"path/filepath"
        _ "modernc.org/sqlite"
        "encoding/hex"
		"crypto/sha1"

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
    hash text primary key, 
    path text, 
    size real, 
    content text
    ) without rowid;
create table commits(
    hash text primary key, 
    author text, 
    email text,
	message text
    ) without rowid;
create table commit_objects(
    commit_hash text, 
    object_hash text  
    );
create table _index(
	hash text primary key
	) without rowid;
    `) ; err != nil {
        panic(err)
    }
}

func getObjectFromRows(rows *sql.Rows) types.Object{
	var obj types.Object
	var size_str string
	var hash_str string
	if err := rows.Scan(&hash_str, &obj.Path, &size_str, &obj.Content); err != nil {
		panic(err)
	}
	hash_slice, _ := hex.DecodeString(hash_str)
	copy(obj.Hash[:], hash_slice)
	size_64, _ := strconv.ParseInt(size_str, 10, 32)
	obj.Size = int32(size_64)
	return obj
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
	defer rows.Close()

	var files []types.Object
    for rows.Next() {
		files = append(files, getObjectFromRows(rows))
    }
	return files
}

func (ds Database_storage) Get() types.Object {
	return ds.FilesAll()[0]
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
	if _, err = db.Exec(`insert into _index (hash) values (?)`, hex.EncodeToString(obj.Hash[:])) ; err != nil {
		panic(err)
	}
}


func (ds Database_storage) Commit(message string){
    db, err := sql.Open("sqlite", "sqlite.db")
    if err != nil {
        panic(err)
    }
	defer db.Close()
	// get objects from index
	rows, err := db.Query("select objects.hash, objects.path, objects.size, objects.content from _index LEFT JOIN objects on _index.hash == objects.hash")
	if err != nil{
		panic(err)
	}
	defer rows.Close()

	commit_hash := sha1.Sum([]byte(message))
	var files []types.Object
	for rows.Next() {
		obj := getObjectFromRows(rows)
		files = append(files, obj)
		// insert index into commits
	}
	for _, file := range files{
		if _, err = db.Exec(`
		insert into commit_objects (commit_hash, object_hash) values (?, ?);
		`, hex.EncodeToString(commit_hash[:]), hex.EncodeToString(file.Hash[:])); err != nil {
			panic(err)
		}
	}
	

    if _, err = db.Exec(`
    insert into commits (hash, author, email, message) values (?, ?, ?, ?);
    `, hex.EncodeToString(commit_hash[:]), "me", "me@mail.com", message); err != nil {
        panic(err)
    }
    
}


