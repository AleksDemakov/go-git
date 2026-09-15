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

func (ds Database_storage) openConnection() (*sql.DB, error) {
	repometa := ds.SearchRepoMeta()
    db, err := sql.Open("sqlite", repometa)
    if err != nil {
        panic(err)
    }
	return db, err
}
func (ds Database_storage) closeConnection (db *sql.DB) {
	db.Close()
}

// SearchRepoMeta returns absolute path to meta file or directory including file or dir itself including "sqlite.db"
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
    content text,
	flag text
    ) without rowid;
create table commits(
    hash text primary key, 
	parent_hash text,
	parent2_hash text,
    author text, 
    email text,
	message text,
	date text
    ) without rowid;
create table commit_object(
    commit_hash text, 
    object_hash text  
    );
create table _index(
	hash text primary key
	) without rowid;
create table refs(
	name text primary key,
	commit_hash text
);
create table head(
	name text
);
    `) ; err != nil {
        panic(err)
    }
	if _, err = db.Exec(`
	INSERT INTO refs 
				(name, commit_hash) 
	VALUES 		("main", NULL);`) ; err != nil {
		panic(err)
	}
	if _, err = db.Exec(`
	INSERT INTO head 
	VALUES 		("main");`) ; err != nil {
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
	//var parent2HashNullable sql.NullString
	var hash string
	// TODO: convert date from string into smt?
	if err := rows.Scan(
		&hash, 
		&parentHashNullable, 
	//	&parent2HashNullable, 
		&commit.Author, 
		&commit.Email, 
		&commit.Message, 
		&commit.Date ); err != nil {
		panic(err)
	}
	
	hashSlice, _ := hex.DecodeString(hash) 
	copy(commit.Hash[:], hashSlice)

	var parentHash string
	if parentHashNullable.Valid {
		parentHash = parentHashNullable.String 
	}else{
		parentHash = "" 
	}
	hashSlice1, _ := hex.DecodeString(parentHash) 
	copy(commit.ParentHash[:], hashSlice1)

//	if parent2HashNullable.Valid {
//		parentHash = parent2HashNullable.String 
//	}else{
//		parentHash = "" 
//	}
//	hashSlice1, _ = hex.DecodeString(parentHash) 
//	copy(commit.Parent2Hash[:], hashSlice1)
	return commit 
}

// TODO: remove this function?
func (ds Database_storage) FilesAll() []types.Object{
	db, err := ds.openConnection()
    defer ds.closeConnection(db)

    rows, err := db.Query(`
	SELECT hash, path, size, content 
	FROM object;
	`)
    if err != nil {panic(err)}
	defer rows.Close()

	var files []types.Object
    for rows.Next() {
		files = append(files, getObjectFromRows(rows))
    }
	return files
}

// TODO: Remove this?
func (ds Database_storage) Get() types.Object {
	return ds.FilesAll()[0]
}

func (ds Database_storage) AddDeletedFile(filepath_ string) {
	db, err := ds.openConnection()
    defer ds.closeConnection(db)
	
	if _, err = db.Exec(`
		DELETE FROM _index 
		WHERE hash = (select hash from objects where path = ?);
	`, filepath_); err != nil {
		panic(err)
	}
}

func (ds Database_storage) Add(obj types.Object){
	db, err := ds.openConnection()
    defer ds.closeConnection(db)


	repometa := ds.SearchRepoMeta()
	relPath, _ := filepath.Rel(filepath.Dir(repometa), obj.Path)
    if _, err = db.Exec(`
    INSERT INTO objects 
				(hash, path, size, content) 
	VALUES 		(?, ?, ?, ?);`, hex.EncodeToString(obj.Hash[:]), relPath, obj.Size, obj.Content); err != nil {
        panic(err)
    }
	q := `
	UPDATE _index 
	SET	   hash = ? 
	WHERE  hash = 
	(SELECT _index.hash 
	FROM  	_index 
	INNER JOIN objects
	ON 	objects.path = ? 
	AND _index.hash = objects.hash);
	`
	if _, err = db.Exec(q, hex.EncodeToString(obj.Hash[:]), relPath) ; err != nil {
		panic(err)
	}
	q1 := `
	INSERT INTO _index 
				(hash) 
	VALUES 		(?) 
	ON CONFLICT (hash) 
	DO NOTHING;	`
	if _, err = db.Exec(q1, hex.EncodeToString(obj.Hash[:])) ; err != nil {
		panic(err)
	}
}

func (ds Database_storage) Index() []types.Object {
	db, err := ds.openConnection()
    defer ds.closeConnection(db)

	// get objects from index
	rows, err := db.Query(`
	SELECT  	obj.hash, obj.path, obj.size, obj.content 
	FROM		_index 
	INNER JOIN  objects AS obj 
	ON 			_index.hash = obj.hash;
	`)
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
	db, err := ds.openConnection()
    defer ds.closeConnection(db)

	rows, err := db.Query(`
	SELECT hash, parent_hash, author, email, message, date 
	FROM   commits
	WHERE  hash = 
	(SELECT commit_hash FROM head INNER JOIN refs ON head.name = refs.name);`)
	if err != nil {
		panic(err)
	}
	var commits []types.Commit
	if !rows.Next() {
		// TODO: check why always prints
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
func (ds Database_storage) GetObjectsFromLastCommits(limit int) []types.Object {
	db, err := ds.openConnection()
    defer ds.closeConnection(db)

	rows, err := db.Query(`
	SELECT objects.hash, objects.path, objects.size, objects.content 
	FROM   commit_object INNER JOIN objects
	ON	   commit_object.object_hash = objects.hash AND commit_object.commit_hash = 
	(SELECT commit_hash 
	FROM   head INNER JOIN refs ON head.name = refs.name);`)
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
// TODO: actually same as GetObjectsFromCommits(int). mb to abstract?
func (ds Database_storage) GetObjectsFromCommit(commit string) []types.Object {
	db, err := ds.openConnection()
    defer ds.closeConnection(db)
	rows, err := db.Query(`
	SELECT objects.hash, objects.path, objects.size, objects.content 
	FROM   commit_object INNER JOIN objects
	ON	   commit_object.object_hash = objects.hash AND commit_object.commit_hash = ?;`, commit)
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

func findObjectInSliceByPath(obj types.Object, slice []types.Object) int {
	for index, entry := range(slice) {
		if obj.Path == entry.Path {
			return index
		}
	}
	return  -1
}

// FindHashesForMerge returns {"base":baseHash, "theirs":theirsHash, "ours":oursHash}
func (ds Database_storage) FindHashesForMerge(theirsBranchName string) map[string]string {
	db, err := ds.openConnection()
    defer ds.closeConnection(db)
	oursHash := hex.EncodeToString(ds.Log(1)[0].Hash[:])
	var theirsHash string 
	rows, err := db.Query(`
	select commit_hash from refs where name = ?;
	`, theirsBranchName)
	for rows.Next(){
		if err = rows.Scan(&theirsHash);err != nil {panic(err)}
	}
	// find LCA (least common ancestor)
	rows, _ = db.Query(`
	WITH RECURSIVE commits_graph(hash, parent_hash) AS (
		SELECT hash, parent_hash FROM commits
	),
	hit(hash, depth) AS (
		SELECT ?, 0  
		UNION ALL
		SELECT ?, 0 
		UNION ALL
		SELECT commits_graph.parent_hash, hit.depth + 1
		FROM hit
		JOIN commits_graph ON hit.hash = commits_graph.hash
		WHERE commits_graph.parent_hash IS NOT NULL
	)
	SELECT hit.hash
	FROM hit
	WHERE hit.hash IN (
		SELECT hash FROM hit GROUP BY hash HAVING COUNT(*) = 2  
	)
	ORDER BY hit.depth DESC  
	LIMIT 1;  
	`, oursHash, theirsHash)
//	rows, _ = db.Query(`
//	with recursive commits_graph(hash, parent_hash) as (
//		select hash, parent_hash from commits
//	),
// 	hit(hash, cnt) as (
//		select ?, 1
//		union all
//		select ?, 1
//		union all 
//		select commits_graph.parent_hash, hit.cnt from hit join commits_graph on hit.hash = commits_graph.hash 
//	)
//	select hit.hash as cnt from hit where hit.hash is not null group by hit.hash order by cnt desc limit 1;
//	`, oursHash, theirsHash)
	var baseHash string 
	for rows.Next(){
		if err = rows.Scan(&baseHash); err != nil {panic(err)}
	}
	return map[string]string {"base": baseHash, "theirs": theirsHash, "ours": oursHash}
}

// TODO: write purpose of parameters
// TODO: use references in conflictedFile type
func (ds Database_storage) Merge(branchName string, strategy string, commitObjects []types.Object, conflictedObjects []types.ConflictFile) {
	db, err := ds.openConnection()
    defer ds.closeConnection(db)


	// Transform the index
	// TODO: optimize. Many requests, mb unite into transaction or smth
	if _, err := db.Exec(`
	DROP TABLE _index;
	CREATE TABLE _index(
		hash text primary key
	) without rowid;
	`); err != nil {panic(err)}

	// insert objects into index 
	for _, object := range commitObjects {
		if _, err = db.Exec(`
		INSERT INTO _index (hash) 
		VALUES 		(?);`,  hex.EncodeToString(object.Hash[:])); err != nil {panic(err)}
	}
	
	for _, conflictObject := range conflictedObjects {
		if conflictObject.Conflict == types.BothModified {
			if _, err = db.Exec(`
			INSERT INTO _index (hash) 
			VALUES 		(?) ON CONFLICT DO NOTHING;`,  hex.EncodeToString(conflictObject.Ours.Hash[:])); err != nil {panic(err)}
			if _, err = db.Exec(`
				UPDATE objects 
				SET flag = 'CE_CONFLICT'
				WHERE hash = ?;
			`, hex.EncodeToString(conflictObject.Ours.Hash[:])); err != nil {panic(err)}

			if _, err = db.Exec(`
				UPDATE objects 
				SET flag = 'CE_CONFLICT'
				WHERE hash = ?;
			`, hex.EncodeToString(conflictObject.Theirs.Hash[:])); err != nil {panic(err)}

		}
		if conflictObject.Conflict == types.TheirDeleted {
			if _, err = db.Exec(`
			INSERT INTO _index (hash) 
			VALUES 		(?) ON CONFLICT DO NOTHING;`,  hex.EncodeToString(conflictObject.Ours.Hash[:])); err != nil {panic(err)}
			if _, err = db.Exec(`
				UPDATE objects 
				SET flag = 'CE_CONFLICT'
				WHERE hash = ?;
			`, hex.EncodeToString(conflictObject.Ours.Hash[:])); err != nil {panic(err)}
		}
		if conflictObject.Conflict == types.OurDeleted {
			if _, err = db.Exec(`
			INSERT INTO _index (hash) 
			VALUES 		(?) ON CONFLICT DO NOTHING;`,  hex.EncodeToString(conflictObject.Theirs.Hash[:])); err != nil {panic(err)}
			if _, err = db.Exec(`
				UPDATE objects 
				SET flag = 'CE_CONFLICT'
				WHERE hash = ?;
			`, hex.EncodeToString(conflictObject.Theirs.Hash[:])); err != nil {panic(err)}
		}

	}

	// Stop if there is a conflict
	if len(conflictedObjects) > 0 {
		return
	}

	message := fmt.Sprintf("Merge %s to %s", branchName, "HEAD")
	hashes := ds.FindHashesForMerge(branchName)
	//baseHash := hashes["base"]
	theirsHash := hashes["theirs"]
	oursHash := hashes["ours"]
	var parent1Hash, parent2Hash string
	parent1Hash = oursHash
	parent2Hash = theirsHash
	if strategy == "theirs"{
		parent1Hash, parent2Hash = parent2Hash, parent1Hash  //swap
	}
	
	var joinedObjectsHash string
	// combine all info about objects and commit for hash calculation
	if strategy == "ours" || strategy == "theirs" {
		for _, object := range ds.GetObjectsFromCommit(parent1Hash){
			joinedObjectsHash += hex.EncodeToString(object.Hash[:])
		}
	} else {
		for _, object := range commitObjects{
			joinedObjectsHash += hex.EncodeToString(object.Hash[:])
		}
	}
	bytesForHash := fmt.Sprintf("parent %s, parent2 %s, author %s, email %s, message %s, objects %s", parent1Hash, parent2Hash, "me", "me@mail.com", message, joinedObjectsHash)
	commitHashSum := sha1.Sum([]byte(bytesForHash))
	commitHash := hex.EncodeToString(commitHashSum[:])

	// insert objects in link table
	if strategy == "ours" || strategy == "theirs" {
		// copy from parent commit
		if _, err = db.Exec(`
		INSERT INTO commit_object (commit_hash, object_hash) 
		SELECT 	 	?, object_hash 
		FROM 	    commit_object 
		WHERE 		commit_hash = ?;`, commitHash, parent1Hash); err != nil {panic(err)}
	} else {
		// TODO: optimize. Many requests, mb unite into transaction or smth
		// insert objects into link table
		for _, object := range commitObjects {
			if _, err = db.Exec(`
			INSERT INTO commit_object (commit_hash, object_hash) 
			VALUES 		(?, ?);`, commitHash, hex.EncodeToString(object.Hash[:])); err != nil {panic(err)}
		}
	}

	// Create merge commit
	if _, err = db.Exec(`
	INSERT INTO commits (hash, parent_hash, parent2_hash, author, email, message, date) 
	VALUES 				(?,?,  ?, ?, ?, ?, datetime("now"));
	`, commitHash, parent1Hash, parent2Hash, "me", "me@mail.com", message); err != nil {
		panic(err)
	}
	// Move refs
	if _, err = db.Exec(`
	UPDATE  refs
	SET 	commit_hash = ? 
	WHERE   name = (select name from head);
	`, commitHash); err != nil {
		panic(err)
	}


}

//TODO: validate if all changes from workuing tree are added
func (ds Database_storage) Commit(message string) {
	db, err := ds.openConnection()
    defer ds.closeConnection(db)

	parent := ds.Log(1)
	var parentHash string
	if parent == nil {
		parentHash = ""
	} else {
		parentHash = hex.EncodeToString(parent[0].Hash[:])
	}

	index := ds.Index()
	var mergedObjectsHash string
	for _, object := range index {
		mergedObjectsHash += hex.EncodeToString(object.Hash[:])
	}
	//TODO: insert email
	bytesForHash := fmt.Sprintf("parent %s, author %s, message %s, objects %s", parentHash, "me", message, mergedObjectsHash)
	
	commitHash := sha1.Sum([]byte(bytesForHash))
	commitHashString := hex.EncodeToString(commitHash[:])
	// TODO: return commitHashString
	// fmt.Println(commitHashString)
	for _, file := range index {
		// Insert row about commited files in link table
		if _, err = db.Exec(`
		INSERT INTO commit_object (commit_hash, object_hash) VALUES (?, ?);
		`, commitHashString, hex.EncodeToString(file.Hash[:])); err != nil {
			panic(err)
		}
	}
	
	// Insert commit row in Commit table
    if _, err = db.Exec(`
    INSERT INTO commits (hash, parent_hash, parent2_hash, author, email, message, date) 
	VALUES 				(?,(select refs.commit_hash from head join refs on head.name = refs.name), null, ?, ?, ?, datetime("now"));
    `, commitHashString, "me", "me@mail.com", message); err != nil {
        panic(err)
    }

	// Move refs
    if _, err = db.Exec(`
    UPDATE 	refs 
	SET 	commit_hash = ? 
	WHERE   name = (select name from head);
    `, commitHashString); err != nil {
        panic(err)
    }

}

func (ds Database_storage) MoveHeadTo(name string){
	db, err := ds.openConnection()
    defer ds.closeConnection(db)
    if err != nil {panic(err)}
	if _, err := db.Exec(`
	UPDATE head
	SET name = ?;
	`, name); err != nil {
		panic(err)
	}
}
// TODO: validation for unsaved items 
func (ds Database_storage) MoveIndexTo(name string) {
	db, err := ds.openConnection()
    defer ds.closeConnection(db)
    if err != nil {panic(err)}
	
	if _, err := db.Exec(`
	DROP TABLE _index;
	CREATE TABLE _index(
		hash text primary key
	) without rowid;


	INSERT INTO _index (hash)
	-- get all object hashes from <name> branch and insert in _index table
	SELECT objects.hash
	FROM   commit_object INNER JOIN objects
	ON	   commit_object.object_hash = objects.hash AND commit_object.commit_hash = 
	(SELECT commit_hash 
	FROM   refs 
	WHERE name = ?);`, name); err != nil {
		panic(err)
	}
}
func (ds Database_storage) NewBranch(name string){
	db, err := ds.openConnection()
    defer ds.closeConnection(db)
    if err != nil {panic(err)}
	if _, err := db.Exec(`
	INSERT INTO refs (name, commit_hash)
	VALUES (?, (select refs.commit_hash from head inner join refs on head.name = refs.name));
	`, name); err != nil {
		panic(err)
	}
}

func (ds Database_storage) DeleteBranch(name string){
	db, err := ds.openConnection()
    defer ds.closeConnection(db)
    if err != nil {panic(err)}
	if _, err := db.Exec(`
	DELETE FROM refs WHERE name = ?;`, name); err != nil {
		panic(err)
	}
}

func (ds Database_storage) Branch() ([]string, string) {
	db, err := ds.openConnection()
    defer ds.closeConnection(db)

	var branches []string
	rows, err := db.Query("SELECT name FROM refs;")
    if err != nil {panic(err)}
	for rows.Next() {
		var name string 
		rows.Scan(&name)
		branches = append(branches, name)
	}
	row, err := db.Query("SELECT name FROM head;")
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
	db, err := ds.openConnection()
    defer ds.closeConnection(db)
	if err != nil {panic(err)}

	var files []types.Object
	if staged {
		// restore file in the index from a source 
		var rows *sql.Rows
		var err error

		if source == "HEAD" {
			rows, err = db.Query(`
			SELECT	   hash, size, path, content 
			FROM  	   object AS obj 
			INNER JOIN commit_object AS co 
			ON 	       co.commit_hash = 
			(SELECT 	refs.commit_hash 
			FROM 		head 
			INNER JOIN  refs 
			ON 			head.name = refs.name) 
			AND 		obj.path = ?;`,
			 filename)
		} else {
			rows, err = db.Query(`
			SELECT   	hash, size, path, content 
			FROM  	    object AS obj 
			INNER JOIN  commit_object AS co 
			ON 			co.commit_hash = ? 
			AND 		obj.path = ?;`,
			source,  filename)
		}
		if err != nil{panic(err)}
		for rows.Next() {
			obj := getObjectFromRows(rows)
			files = append(files, obj)
		}
		if _, err = db.Exec(`
		UPDATE   _index 
		SET		 hash = ? 
		WHERE 	 hash = 
		(SELECT _index.hash 
		FROM _index INNER JOIN object
		ON _index.hash = object.hash AND object.path = ? );`,hex.EncodeToString(files[0].Hash[:]), filename ) ; err != nil {
			panic(err)
		}
	} else {
		// restore file in working tree from the index	
		rows, err := db.Query(`
		SELECT object.hash, object.path, object.size, object.content 
		FROM _index INNER JOIN objects 
		ON _index.hash = object.hash AND object.path = ?;`, filename)
		if err != nil {panic(err)}
		for rows.Next() {
			obj := getObjectFromRows(rows)
			files = append(files, obj)
		}
	}
	return files[0]
}
