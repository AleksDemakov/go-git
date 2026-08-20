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


func getObjectFromRows(rows *sql.Rows) types.Object {
	var obj types.Object
	var size_str string
	var hash_str string
	if err := rows.Scan(&hash_str, &obj.Path, &size_str, &obj.Content); err != nil {
		panic(err)
	}
	hash_slice, err := hex.DecodeString(hash_str)
	if err != nil {panic(err)}
	copy(obj.Hash[:], hash_slice)
	size_64, err := strconv.ParseInt(size_str, 10, 32)
	if err != nil {panic(err)}
	obj.Size = int32(size_64)
	return obj
}

func getCommitFromRows(rows *sql.Rows) types.Commit {
	var commit types.Commit
	var parentHashNullable sql.NullString
	var hash string
	// TODO: convert date from string into smt?
	if err := rows.Scan(&hash, &parentHashNullable, &commit.Author, &commit.Email, &commit.Message, &commit.Date); err != nil {
		panic(err)
	}
	var parentHash string
	if parentHashNullable.Valid {
		parentHash = parentHashNullable.String 
	}else{
		parentHash = "" 
	}
	hashSlice, _ := hex.DecodeString(hash) 
	copy(commit.Hash[:], hashSlice)
	hashSlice1, _ := hex.DecodeString(parentHash) 
	copy(commit.ParentHash[:], hashSlice1)
	
	return commit 
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

	relPath, _ := filepath.Rel(filepath.Dir(repometa), obj.Path)
    if _, err = db.Exec(`
    insert into objects (hash, path, size, content) values (?, ?, ?, ?);
    `, hex.EncodeToString(obj.Hash[:]), relPath, obj.Size, obj.Content); err != nil {
        panic(err)
    }
	q := `
	update _index set hash = ? where hash = (select _index.hash from _index inner join objects on objects.path = ? and _index.hash = objects.hash);
	`
	if _, err = db.Exec(q, hex.EncodeToString(obj.Hash[:]), relPath) ; err != nil {
		panic(err)
	}
	q1 := `	insert into _index (hash) values (?) on conflict (hash) do nothing;	`
	if _, err = db.Exec(q1, hex.EncodeToString(obj.Hash[:])) ; err != nil {
		panic(err)
	}
}

// TODO: rename to Index? because is is just returns index
func (ds Database_storage) Status() []types.Object {
	repometa := ds.SearchRepoMeta()
    db, err := sql.Open("sqlite", repometa)
	if err != nil {panic(err)}
	defer db.Close()

	// get objects from index
	rows, err := db.Query("select objects.hash, objects.path, objects.size, objects.content from _index INNER JOIN objects on _index.hash = objects.hash;")
	if err != nil {panic(err)}
	defer rows.Close()

	var files []types.Object
	for rows.Next() {
		obj := getObjectFromRows(rows)
		files = append(files, obj)
	}
	return files
}

// TODO: actualy use limit. Now it returns only 1 commit
func (ds Database_storage) Log(limit int) []types.Commit {
	repometa := ds.SearchRepoMeta()
    db, err := sql.Open("sqlite", repometa)
	if err != nil {panic(err)}
	defer db.Close()

	rows, err := db.Query("select hash, parent_hash, author, email, message, date from commits where hash = (select commit_hash from head inner join refs on head.name = refs.name);")
	if err != nil {
		panic(err)
	}
	var commits []types.Commit
	if !rows.Next() {
		fmt.Println("no rows found")
		return commits 
	}
	for  {
		commit := getCommitFromRows(rows)
		commits = append(commits, commit)
		if !rows.Next() {
			break	
		}
	}
	return commits	
}

// TODO: actualy use limit. Now it returns only 1 set 
func (ds Database_storage) GetObjectsFromCommits(limit int) []types.Object {
	repometa := ds.SearchRepoMeta()
    db, err := sql.Open("sqlite", repometa)
	if err != nil {panic(err)}
	defer db.Close()

	rows, err := db.Query("SELECT objects.hash, objects.path, objects.size, objects.content FROM commit_objects INNER JOIN objects ON commit_objects.object_hash = objects.hash and commit_objects.commit_hash =  (select commit_hash from head inner join refs on head.name = refs.name);")
	if err != nil {
		panic(err)
	}
	var objects []types.Object
	for rows.Next() {
		object := getObjectFromRows(rows)
		objects = append(objects, object)
	}
	return objects
}

func (ds Database_storage) Commit(message string) {
	repometa := ds.SearchRepoMeta()
    db, err := sql.Open("sqlite", repometa)
    if err != nil {
        panic(err)
    }
	defer db.Close()

	parent := ds.Log(1)
	var parentHash string
	if parent == nil {
		parentHash = ""
	} else {
		parentHash = hex.EncodeToString(parent[0].Hash[:])
	}

	index := ds.Status()
	var mergedObjectsHash string
	for _, object := range index {
		mergedObjectsHash += hex.EncodeToString(object.Hash[:])
	}
	//TODO: insert email
	bytesForHash := fmt.Sprintf("parent %s, author %s, message %s, objects %s", parentHash, "me", message, mergedObjectsHash)
	
	commitHash := sha1.Sum([]byte(bytesForHash))
	commitHashString := hex.EncodeToString(commitHash[:])
	fmt.Println(commitHashString)
	for _, file := range index {
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
	
    defer db.Close()

	var branches []string
	rows, err := db.Query("select name from refs;")
    if err != nil {
        panic(err) 
	}
	for rows.Next() {
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

func (ds Database_storage) Restore(staged bool, source string, filename string) types.Object {
	repometa := ds.SearchRepoMeta()
    db, err := sql.Open("sqlite", repometa)
    if err != nil {panic(err)}
    defer db.Close()

	var files []types.Object
	if staged {
		// restore file in the index from a source 
		var rows *sql.Rows
		var err error

		if source == "HEAD" {
			rows, err = db.Query(`SELECT hash, size, path, content FROM  
			objects as obj INNER JOIN commit_objects as co ON co.commit_hash = 
			(select refs.commit_hash from head INNER JOIN refs ON head.name == refs.name) 
			and obj.path = ?;`,
			 filename)
		} else {
			rows, err = db.Query(`SELECT hash, size, path, content FROM  
			objects as obj INNER JOIN commit_objects as co ON co.commit_hash = ? 
			and obj.path = ?;`,
			source,  filename)
		}
		if err != nil{panic(err)}
		for rows.Next() {
			obj := getObjectFromRows(rows)
			files = append(files, obj)
		}
		if _, err = db.Exec(`update _index set hash = ? where hash = 
		(select _index.hash from _index inner join objects on _index.hash = objects.hash and objects.path = ? );`,hex.EncodeToString(files[0].Hash[:]), filename ) ; err != nil {
			panic(err)
		}
	} else {
		// restore file in working tree from the index	
		rows, err := db.Query("select objects.hash, objects.path, objects.size, objects.content from _index INNER JOIN objects on _index.hash = objects.hash and  objects.path = ?;", filename)
		if err != nil {panic(err)}
		for rows.Next() {
			obj := getObjectFromRows(rows)
			files = append(files, obj)
		}
	}
	return files[0]
}
