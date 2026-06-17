package main


import (
	"fmt"
    "os"
	"path/filepath"
)

// Looks for .go-git directory by walking upside toward root dir
func findLocalRepo(curDir string) string {

    if curDir == "/" {
        panic(fmt.Sprintln("there is no go-git repository"))
    }
    if _, err := os.Stat(filepath.Join(curDir,".go-git")); err != nil {
        curDir1 := filepath.Dir(curDir)
        result := findLocalRepo(curDir1)
        return result
    }else{
        return curDir+"/.go-git/"
    }

}

func ReadArg(command string, NameFile string) {
	if command == "add" {
        Add(NameFile)
		
	} else if command == "commit" {
        workingDir, _ := os.Getwd()
        gog := findLocalRepo(workingDir)
		fmt.Println(gog)
	} else if command == "push" {
		fmt.Println(3)
	} else {
		fmt.Println("Failed command")
	}
}

func main() {

	command := os.Args[1]

    var NameFile string
    if len(os.Args) > 2 {
        NameFile = os.Args[2]
    }

	ReadArg(command, NameFile)

}
