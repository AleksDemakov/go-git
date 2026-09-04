package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestBranchAndMergePipeline reproduces the exact pipeline described in the requirements:
// ./go-git init
// cat file.txt -> "version 1"
// modify file.txt to "version 2"
// ./go-git add file.txt
// ./go-git commit "init"
// ./go-git branch feature
// ./go-git checkout feature
// modify file.txt to "version 2" 
// ./go-git add file.txt
// ./go-git commit "v2"
// ./go-git checkout main
// cat file.txt -> "version 1"
// ./go-git merge feature -s ours
// Currently outputs "Conflict" but should merge successfully with ours strategy
func TestBranchAndMergePipeline(t *testing.T) {
	// Create a temporary directory for the test repo
	tmpDir, err := os.MkdirTemp("", "go-git-pipeline-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binPath, _ := filepath.Abs("../go-git")
	
	// Change to the temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Fatalf("Failed to find ./go-git")
	}

	// 1. Initialize repository
	execGit(t, exec.Command(binPath, "init"))

	// 2. Create file.txt with "version 1" on main branch
	testFile := "file.txt"
	contentV1 := "version 1"
	if err := os.WriteFile(testFile, []byte(contentV1), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 3. Add and commit on main with message "init"
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "init"))

	// 4. Create feature branch
	execGit(t, exec.Command(binPath, "branch", "feature"))

	// 5. Switch to feature branch
	execGit(t, exec.Command(binPath, "checkout", "feature"))

	// 6. Modify file.txt to "version 2" on feature branch
	contentV2 := "version 2"
	if err := os.WriteFile(testFile, []byte(contentV2), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// 7. Add and commit on feature branch with message "v2"
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "v2"))

	// 8. Switch back to main branch
	execGit(t, exec.Command(binPath, "checkout", "main"))

	// 9. Verify we're on main and file.txt has original content "version 1"
	currentContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file on main: %v", err)
	}
	if string(currentContent) != contentV1 {
		t.Errorf("Expected file content on main to be '%s', got '%s'", contentV1, string(currentContent))
	}

	// 10. Merge feature branch into main with ours strategy
	// This currently outputs "Conflict" but should succeed
	execGit(t, exec.Command(binPath, "merge", "-s", "ours", "feature"))

	// 11. Verify merge result - with "ours" strategy, file should keep main's content (version 1)
	mergedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file after merge: %v", err)
	}

	if string(mergedContent) != contentV1 {
		t.Errorf("Merge with -s ours failed: expected content '%s' (main's version), got '%s'", contentV1, string(mergedContent))
	}

	// 12. Verify both branches still exist
	execGit(t, exec.Command(binPath, "branch"))

	t.Logf("Pipeline test passed: file content after merge is '%s' (ours strategy preserved main's version)", string(mergedContent))
}

// TestBranchAndMergePipelineTheirs tests the same pipeline with "theirs" strategy
// With "theirs" strategy, the merge should preserve the feature branch's content (version 2)
func TestBranchAndMergePipelineTheirs(t *testing.T) {
	// Create a temporary directory for the test repo
	tmpDir, err := os.MkdirTemp("", "go-git-pipeline-theirs-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binPath, _ := filepath.Abs("../go-git")
	
	// Change to the temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Fatalf("Failed to find ./go-git")
	}

	// 1. Initialize repository
	execGit(t, exec.Command(binPath, "init"))

	// 2. Create file.txt with "version 1" on main branch
	testFile := "file.txt"
	contentV1 := "version 1"
	if err := os.WriteFile(testFile, []byte(contentV1), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 3. Add and commit on main
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "init"))

	// 4. Create feature branch
	execGit(t, exec.Command(binPath, "branch", "feature"))

	// 5. Switch to feature branch
	execGit(t, exec.Command(binPath, "checkout", "feature"))

	// 6. Modify file.txt to "version 2" on feature branch
	contentV2 := "version 2"
	if err := os.WriteFile(testFile, []byte(contentV2), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// 7. Add and commit on feature branch
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "v2"))

	// 8. Switch back to main branch
	execGit(t, exec.Command(binPath, "checkout", "main"))

	// 9. Merge feature branch into main with theirs strategy
	// With "theirs" strategy, the merge should preserve feature's content (version 2)
	execGit(t, exec.Command(binPath, "merge", "-s", "theirs", "feature"))

	// 10. Verify merge result - with "theirs" strategy, file should have feature's content (version 2)
	mergedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file after merge: %v", err)
	}

	if string(mergedContent) != contentV2 {
		t.Errorf("Merge with -s theirs failed: expected content '%s' (feature's version), got '%s'", contentV2, string(mergedContent))
	}

	t.Logf("Pipeline test passed with theirs strategy: file content after merge is '%s'", string(mergedContent))
}

// TestBranchAndMergePipelineRecursive tests the same pipeline with "recursive" strategy
// With "recursive" strategy, the merge should auto-merge non-conflicting changes
func TestBranchAndMergePipelineRecursive(t *testing.T) {
	// Create a temporary directory for the test repo
	tmpDir, err := os.MkdirTemp("", "go-git-pipeline-recursive-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binPath, _ := filepath.Abs("../go-git")
	
	// Change to the temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		t.Fatalf("Failed to find ./go-git")
	}

	// 1. Initialize repository
	execGit(t, exec.Command(binPath, "init"))

	// 2. Create file.txt with "version 1" on main branch
	testFile := "file.txt"
	contentV1 := "version 1"
	if err := os.WriteFile(testFile, []byte(contentV1), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 3. Add and commit on main
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "init"))

	// 4. Create feature branch
	execGit(t, exec.Command(binPath, "branch", "feature"))

	// 5. Switch to feature branch
	execGit(t, exec.Command(binPath, "checkout", "feature"))

	// 6. Modify file.txt to "version 2" on feature branch
	contentV2 := "version 2"
	if err := os.WriteFile(testFile, []byte(contentV2), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// 7. Add and commit on feature branch
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "v2"))

	// 8. Switch back to main branch
	execGit(t, exec.Command(binPath, "checkout", "main"))

	// 9. Merge feature branch into main with recursive strategy
	// With "recursive" strategy, the merge should handle the conflict appropriately
	execGit(t, exec.Command(binPath, "merge", "-s", "recursive", "feature"))

	// 10. Verify merge result - with "recursive" strategy
	// For this test, we just check that the merge completes without panicking
	// The actual content will depend on how recursive strategy is implemented
	mergedContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read file after merge: %v", err)
	}

	// With recursive strategy, it might keep main's version or handle conflict
	// This test documents the current behavior
	t.Logf("Pipeline test with recursive strategy: file content after merge is '%s'", string(mergedContent))
}