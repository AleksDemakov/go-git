package core;

import(
	"os"
	"path/filepath"
	"io/fs"
	"fmt"
    types "github.com/Clorter/go-git/types"
)

type workingTreeTransform struct {}

func (tr workingTreeTransform) Apply(rootDir string, untrackedFiles []types.Object, newObjects []types.Object){
	// 1. delete all files except untracked
	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		relPath, _ := filepath.Rel(rootDir, path)
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", relPath, err)
			return err
		}
		if info.IsDir() {
			if info.Name()[0] == '.' {
				return filepath.SkipDir
			}
			return nil
		}

		if findObjectInSliceByPathSuffix(relPath, untrackedFiles) != -1 {
			return nil 
		}
		if err = os.Remove(path); err != nil {panic(err)}
		return nil
	})
	if err != nil {panic(err)}	
	// 2. recreate all files from newObjects
	for _, object := range newObjects {
		if filepath.Dir(object.Path)!="." {
			if err = os.Mkdir(filepath.Dir(object.Path), 0777); err != nil {
				panic(err)
			}
		}
		file, err := os.Create(filepath.Base(object.Path))
		defer file.Close()
		if err != nil {panic(err)}
		file.WriteString(object.Content)
	}
}
