package tests 

import (
	"os" 
	"testing"
	//"fmt"
    types "github.com/Clorter/go-git/types"
    storage "github.com/Clorter/go-git/storage"
)

func TestDatabasePipeline(t *testing.T) {
    // ====== Add a file ======
    writer := storage.Database_storage{}
	writer.Init() 
	defer os.RemoveAll("sqlite.db")
	obj := types.NewObject("./db_test.go")
	writer.Add(obj)

    // ====== Retrieve and check ======
	files := writer.FilesAll()
	if files[0] != obj {
		t.Errorf("Object is not created. Add failed")	
	}
	
}
