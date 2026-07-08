package types

type Commit struct {
	Hash    [20]byte
	Author  string
	Email   string
	Message string
}

type Object struct {
	Hash    string
	Path    string
	Size    int32
	Content string
}
