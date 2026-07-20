package types

type Commit struct {
	Hash    [20]byte
	Author  string
	Email   string
	Message string
}

// a link struct between commits and commited files
type Commit_objects struct {
	Commit_hash [20]byte
	Objects     []byte
}

type Object struct {
	Hash    [20]byte 
	Path    string
	Size    int32
	Content string
}
