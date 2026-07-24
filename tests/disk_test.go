package tests 

import (
	"path/filepath"
	"testing"
	"encoding/hex"
	"os"
    types "github.com/Clorter/go-git/types"
    storage "github.com/Clorter/go-git/storage"
)

func TestDiskPipeline(t *testing.T) {
    writer := storage.Disk_storage{}
	writer.Init()
	defer os.RemoveAll(".go-git/")
	obj := types.NewObject("./disk_test.go")
	writer.Add(obj)
	if _, err := os.Stat(filepath.Join("./.go-git/objects", hex.EncodeToString(obj.Hash[:]))); os.IsNotExist(err) {
		t.Errorf("Object is not created. Add failed")	
	}
	
}
