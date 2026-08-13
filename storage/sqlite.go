package storage 
import (
        "database/sql"
        "os"
		"strconv"
        "fmt"
        "path/filepath"
        _ "modernc.org/sqlite"
        "encoding/hex"
		"crypto/sha1"

        types "github.com/Clorter/go-git/types"
)

type Database_storage struct{
}

// SearchRepoMeta returns absolute path to meta file or directory including file or dir itself
func (ds Database_storage) SearchRepoMeta() string{
	absPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(absPath, "sqlite.db")); os.IsNotExist(err) {
			dir := filepath.Dir(absPath)
			if dir == absPath {
				fmt.Println("file not exist")
				panic(fmt.Sprintln("there is no go-git repository"))
			}
			absPath = dir
		} else {
			return filepath.Join(absPath, "sqlite.db")
		}
	}
}

func (ds Database_storage) Init(){
    if _, err := os.Create("sqlite.db"); err!=nil {
        panic(err)
    } 
    db, err := sql.Open("sqlite", "sqlite.db")
    if err != nil {
        panic(err) }
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
	parent_hash text,
    author text, 
    email text,
	message text,
	date text
    ) without rowid;
create table commit_objects(
    commit_hash text, 
    object_hash text  
    );
create table _index(
	hash text primary key
	) without rowid;
create table refs(
	name text,
	commit_hash text
);
create table head(
	name text
);
    `) ; err != nil {
        panic(err)
    }
	if _, err = db.Exec(`insert into refs (name, commit_hash) values ("main", NULL)`) ; err != nil {
		panic(err)
	}
	if _, err = db.Exec(`insert into head values ("main")`) ; err != nil {
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
	repometa := ds.SearchRepoMeta()
    db, err := sql.Open("sqlite", repometa)
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
	repometa := ds.SearchRepoMeta()
    db, err := sql.Open("sqlite", repometa)
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

func (ds Database_storage) Status() []types.Object {
	repometa := ds.SearchRepoMeta()
    db, err := sql.Open("sqlite", repometa)
    if err != nil {
        panic(err)
    }
	defer db.Close()
	// get objects from index
	rows, err := db.Query("select objects.hash, objects.path, objects.size, objects.content from _index LEFT JOIN objects on _index.hash == objects.hash")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var files []types.Object
	for rows.Next() {
		obj := getObjectFromRows(rows)
		files = append(files, obj)
		// insert index into commits
	}
	return files
}

func (ds Database_storage) Commit(message string) {
	repometa := ds.SearchRepoMeta()
    db, err := sql.Open("sqlite", repometa)
    if err != nil {
        panic(err)
    }
	defer db.Close()
	
	commitHash := sha1.Sum([]byte(message))
	commitHashString := hex.EncodeToString(commitHash[:])
	for _, file := range ds.Status(){
		// Insert row about commited files in link table
		if _, err = db.Exec(`
		insert into commit_objects (commit_hash, object_hash) values (?, ?);
		`, commitHashString, hex.EncodeToString(file.Hash[:])); err != nil {
			panic(err)
		}
	}
	
	
	// Insert commit row in Commit table
    if _, err = db.Exec(`
    insert into commits (hash, parent_hash, author, email, message, date) values (?,(select commit_hash from refs where name = "main"), ?, ?, ?, datetime("now"));
    `, commitHashString, "me", "me@mail.com", message); err != nil {
        panic(err)
    }

	// Move refs
    if _, err = db.Exec(`
    update refs set commit_hash = ? where name = "main";
    `, commitHashString); err != nil {
        panic(err)
    }

	// to Clear index or to not clear?
    //if _, err = db.Exec(`delete from _index;`); err != nil {
        //panic(err)
    //}
}

func (ds Database_storage) Branch() ([]string, string) {
	repometa := ds.SearchRepoMeta()
    db, err := sql.Open("sqlite", repometa)
	rows, err := db.Query("select name from refs;")
	if err != nil{
		panic(err)
	}
	var branches []string
	for rows.Next(){
		var name string 
		rows.Scan(&name)
		branches = append(branches, name)
	}
	row, err := db.Query("select name from head;")
	if err != nil{
		panic(err)
	}
	var curBranch string
	for row.Next(){
		row.Scan(&curBranch)
	}
	return branches, curBranch
}

func (ds Database_storage) Restore(isIndex bool, source string, filname string) types.object {
	repometa := ds.SearchRepoMeta()
    db, err := sql.Open("sqlite", repometa)
	if isIndex {
		// restore file in index from a source commit 
		rows, err := db.Query("select hash, size, path, content from objects where 
		()")
		if err != nil{
			panic(err)
		}
	} else {
		// restore file in working tree from the index	
	}
}
