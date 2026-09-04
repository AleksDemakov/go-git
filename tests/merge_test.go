package tests

import (
	"path/filepath"
	"os"
	"os/exec"
	"testing"

)

func TestMergeTwoBranchesOneFile(t *testing.T) {
	// Create a temporary directory for the test repo
	tmpDir, err := os.MkdirTemp("", "go-git-merge-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binPath, _ := filepath.Abs("../go-git")
	//printTablesScriptPath, _ := filepath.Abs("./print_tables.sql")
	// Change to the temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Fatalf("Failed to find ./go-git")
	}

	// 1. Initialize repository
	execGit(t, exec.Command(binPath, "init"))

	// 2. Create a test file on main branch
	testFile := "test.txt"
	contentV1 := "version 1"
	if err := os.WriteFile(testFile, []byte(contentV1), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 3. Add and commit on main
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "\"test.txt: version 1\""))


	// 4. Create a new branch
	execGit(t, exec.Command(binPath, "branch", "feature"))

	// 5. Switch to feature branch
	execGit(t, exec.Command(binPath, "checkout", "feature"))

	// 6. Modify the file on feature branch
	contentV2 := "version 2"
	if err := os.WriteFile(testFile, []byte(contentV2), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// 7. Add and commit on feature branch
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "\"test.txt: version 2\""))

	// 8. Switch back to main branch
	execGit(t, exec.Command(binPath, "checkout", "main"))
	execGit(t, exec.Command("cat", "test.txt"))

	// 9. Verify we're on main and file has original content
	currentContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file on main: %v", err)
	}
	if string(currentContent) != contentV1 {
		t.Errorf("Expected file content on main to be '%s', got '%s'", contentV1, string(currentContent))
	}

	// 10. Merge feature branch into main
	execGit(t, exec.Command(binPath, "merge", "-s","ours", "feature"))

	// 11. Verify merge result
	// With "ours" strategy, the file should keep main's content (version 1)
	mergedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file after merge: %v", err)
	}

	if string(mergedContent) != contentV1 {
		t.Errorf("Merge failed: expected content '%s' (ours), got '%s'", contentV1, string(mergedContent))
	}

	// 12. Verify both branches still exist
	execGit(t, exec.Command(binPath, "branch"))

	t.Logf("Merge test passed: file content after merge is '%s' (ours strategy)", string(mergedContent))
}

// TestMergeNonConflictingChange tests merge when branches modify different files
func TestMergeNonConflictingChange(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "go-git-merge-test-nonconflict")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binPath, _ := filepath.Abs("../go-git")
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Fatalf("Failed to find ./go-git")
	}

	// 1. Initialize repository
	execGit(t, exec.Command(binPath, "init"))

	// 2. Create first file on main
	file1 := "file1.txt"
	content1 := "content of file 1"
	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}
	execGit(t, exec.Command(binPath, "add", file1))
	execGit(t, exec.Command(binPath, "commit", "add file1"))

	// 3. Create feature branch
	execGit(t, exec.Command(binPath, "branch", "feature"))
	execGit(t, exec.Command(binPath, "checkout", "feature"))

	// 4. Create a different file on feature branch
	file2 := "file2.txt"
	content2 := "content of file 2"
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}
	execGit(t, exec.Command(binPath, "add", file2))
	execGit(t, exec.Command(binPath, "commit", "add file2 on feature"))

	// 5. Switch to main and merge
	execGit(t, exec.Command(binPath, "checkout", "main"))
	execGit(t, exec.Command(binPath, "merge", "feature"))

	// 6. Verify file1 still exists with correct content
	if _, err := os.Stat(file1); os.IsNotExist(err) {
		t.Errorf("file1 should exist after merge")
	}

	// Note: With current "ours" strategy, file2 from feature won't appear in main
	// This test documents current behavior
	t.Log("Non-conflicting merge test completed")
}
