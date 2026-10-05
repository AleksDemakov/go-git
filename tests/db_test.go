package tests 

import (
	"os" 
	"os/exec" 
	"testing"
)

func execGit(t *testing.T, cmd *exec.Cmd ) {
	res, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Error: %v\n", err)
	}
	// mark function as a helper, show Logs only in verbose mode
	t.Helper()
	t.Log("running <"+cmd.String()+"> cmd: "+string(res))
}

func TestDatabasePipeline(t *testing.T) {
	binPath := "/var/home/alex_lab/wsp/go-git/go-git"
	execGit(t, exec.Command(binPath, "init"))
	defer os.RemoveAll("sqlite.db")
	execGit(t, exec.Command(binPath, "add", "db_test.go"))
	execGit(t, exec.Command(binPath, "status"))
	execGit(t, exec.Command(binPath, "commit" ,"\"new message\""))
	execGit(t, exec.Command(binPath, "status"))
	
}
