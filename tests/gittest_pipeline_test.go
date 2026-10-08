package tests

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestBothChangedAdd reproduces: bothchanged_add.sh
// Tests merge conflict when same file is modified in both branches, then resolved with add
func TestBothChangedAdd(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "go-git-bothchanged-add-test")
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

	// Initialize repository
	execGit(t, exec.Command(binPath, "init"))

	// Create initial file with content
	testFile := "file.txt"
	initialContent := "line 1\nline 2\nline 3"
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add and commit initial version on main
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "init"))

	// Create feature branch
	execGit(t, exec.Command(binPath, "branch", "feature"))

	// Modify file on main branch
	mainContent := "line 1\nmodified in main  \nline 3"
	if err := os.WriteFile(testFile, []byte(mainContent), 0644); err != nil {
		t.Fatalf("Failed to modify test file on main: %v", err)
	}

	// Add and commit modification on main
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "modify in main"))

	// Switch to feature branch
	execGit(t, exec.Command(binPath, "checkout", "feature"))

	// Modify same file on feature branch
	featureContent := "line 1\nmodified in feature \nline 3"
	if err := os.WriteFile(testFile, []byte(featureContent), 0644); err != nil {
		t.Fatalf("Failed to modify test file on feature: %v", err)
	}

	// Add and commit modification on feature
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "modify in feature"))

	// Switch back to main
	execGit(t, exec.Command(binPath, "checkout", "main"))

	// First merge attempt - should trigger conflict
	cmd := exec.Command(binPath, "merge", "feature")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	t.Logf("First merge output: %s", outputStr)

	// Should detect conflict when same file is modified in both branches
	if !strings.Contains(outputStr, "CONFLICT") && !strings.Contains(outputStr, "Conflict") {
		t.Logf("Warning: Expected conflict message, got: %s", outputStr)
	}

	// Use add to resolve conflict
	execGit(t, exec.Command(binPath, "add", testFile))

	// Second merge attempt - should complete without error
	cmd2 := exec.Command(binPath, "merge", "feature")
	output2, err2 := cmd2.CombinedOutput()
	outputStr2 := string(output2)
	t.Logf("Second merge output: %s", outputStr2)

	// Validate that second merge DOES NOT contain conflict messages (conflict should be resolved)
	if strings.Contains(outputStr2, "CONFLICT") || strings.Contains(outputStr2, "Conflict") {
		t.Errorf("Second merge should not contain conflict messages, got: %s", outputStr2)
	}
	if err2 != nil {
		t.Errorf("Second merge should succeed without error, got: %v", err2)
	}

	// Verify final state with log - should contain merge commit
	logCmd := exec.Command(binPath, "log")
	logOutput, logErr := logCmd.CombinedOutput()
	logStr := string(logOutput)
	t.Logf("Final log: %s", logStr)

	if logErr != nil {
		t.Errorf("Failed to get log: %v", logErr)
	}

	// Validate that merge commit appears in log
	if !strings.Contains(logStr, "Merge feature branch in main") {
		t.Errorf("Expected merge commit in log, got: %s", logStr)
	}

	t.Log("TestBothChangedAdd completed")
}

// TestBothChangedRm reproduces: bothchanged_rm.sh
// Tests merge conflict when same file is modified in both branches, then resolved with rm
func TestBothChangedRm(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "go-git-bothchanged-rm-test")
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

	// Initialize repository
	execGit(t, exec.Command(binPath, "init"))

	// Create initial file with content
	testFile := "file.txt"
	initialContent := "line 1\nline 2\nline 3"
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add and commit initial version on main
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "init"))

	// Create feature branch
	execGit(t, exec.Command(binPath, "branch", "feature"))

	// Modify file on main branch
	mainContent := "line 1\nmodified in main  \nline 3"
	if err := os.WriteFile(testFile, []byte(mainContent), 0644); err != nil {
		t.Fatalf("Failed to modify test file on main: %v", err)
	}

	// Add and commit modification on main
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "modify in main"))

	// Switch to feature branch
	execGit(t, exec.Command(binPath, "checkout", "feature"))

	// Modify same file on feature branch
	featureContent := "line 1\nmodified in feature \nline 3"
	if err := os.WriteFile(testFile, []byte(featureContent), 0644); err != nil {
		t.Fatalf("Failed to modify test file on feature: %v", err)
	}

	// Add and commit modification on feature
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "modified in feature"))

	// Switch back to main
	execGit(t, exec.Command(binPath, "checkout", "main"))

	// First merge attempt - should trigger conflict
	cmd := exec.Command(binPath, "merge", "feature")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	t.Logf("First merge output: %s", outputStr)

	// Should detect conflict when same file is modified in both branches
	if !strings.Contains(outputStr, "CONFLICT") && !strings.Contains(outputStr, "Conflict") {
		t.Logf("Warning: Expected conflict message, got: %s", outputStr)
	}

	// Use rm to resolve conflict
	execGit(t, exec.Command(binPath, "rm", testFile))

	// Second merge attempt - should complete without error
	cmd2 := exec.Command(binPath, "merge", "feature")
	output2, err2 := cmd2.CombinedOutput()
	outputStr2 := string(output2)
	t.Logf("Second merge output: %s", outputStr2)

	// Validate that second merge DOES NOT contain conflict messages (conflict should be resolved)
	if strings.Contains(outputStr2, "CONFLICT") || strings.Contains(outputStr2, "Conflict") {
		t.Errorf("Second merge should not contain conflict messages, got: %s", outputStr2)
	}
	if err2 != nil {
		t.Errorf("Second merge should succeed without error, got: %v", err2)
	}

	// Verify final state with log - should contain merge commit
	logCmd := exec.Command(binPath, "log")
	logOutput, logErr := logCmd.CombinedOutput()
	logStr := string(logOutput)
	t.Logf("Final log: %s", logStr)

	if logErr != nil {
		t.Errorf("Failed to get log: %v", logErr)
	}

	// Validate that merge commit appears in log
	if !strings.Contains(logStr, "Merge feature branch in main") {
		t.Errorf("Expected merge commit in log, got: %s", logStr)
	}

	t.Log("TestBothChangedRm completed")
}

// TestDeletedInFeatureAdd reproduces: deletedinfeature_add.sh
// Tests merge conflict when file is modified in main but deleted in feature branch, resolved with add
func TestDeletedInFeatureAdd(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "go-git-deletedinfeature-add-test")
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

	// Initialize repository
	execGit(t, exec.Command(binPath, "init"))

	// Create initial file with content
	testFile := "file.txt"
	initialContent := "line 1\nline 2\nline 3"
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add and commit initial version on main
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "init"))

	// Create feature branch
	execGit(t, exec.Command(binPath, "branch", "feature"))

	// Switch to feature branch
	execGit(t, exec.Command(binPath, "checkout", "feature"))

	// Modify file on feature branch
	featureContent := "line 1\nmodified in feature \nline 3"
	if err := os.WriteFile(testFile, []byte(featureContent), 0644); err != nil {
		t.Fatalf("Failed to modify test file on feature: %v", err)
	}

	// Add and commit modification on feature (note: original script had "modify in main" message)
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "modify in main"))

	// Switch back to main
	execGit(t, exec.Command(binPath, "checkout", "main"))

	// Delete file on main branch
	if err := os.Remove(testFile); err != nil {
		t.Fatalf("Failed to delete test file: %v", err)
	}

	// Add and commit deletion on main
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "delete"))

	// First merge attempt - should trigger conflict (deleted in feature, modified in main)
	cmd := exec.Command(binPath, "merge", "feature")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	t.Logf("First merge output: %s", outputStr)

	// Should detect conflict
	if !strings.Contains(outputStr, "CONFLICT") && !strings.Contains(outputStr, "Conflict") {
		t.Logf("Warning: Expected conflict message, got: %s", outputStr)
	}

	// Use add to resolve conflict
	execGit(t, exec.Command(binPath, "add", testFile))

	// Second merge attempt - should complete without error
	cmd2 := exec.Command(binPath, "merge", "feature")
	output2, err2 := cmd2.CombinedOutput()
	outputStr2 := string(output2)
	t.Logf("Second merge output: %s", outputStr2)

	// Validate that second merge DOES NOT contain conflict messages (conflict should be resolved)
	if strings.Contains(outputStr2, "CONFLICT") || strings.Contains(outputStr2, "Conflict") {
		t.Errorf("Second merge should not contain conflict messages, got: %s", outputStr2)
	}
	if err2 != nil {
		t.Errorf("Second merge should succeed without error, got: %v", err2)
	}

	// Verify final state with log - should contain merge commit
	logCmd := exec.Command(binPath, "log")
	logOutput, logErr := logCmd.CombinedOutput()
	logStr := string(logOutput)
	t.Logf("Final log: %s", logStr)

	if logErr != nil {
		t.Errorf("Failed to get log: %v", logErr)
	}

	// Validate that merge commit appears in log
	if !strings.Contains(logStr, "Merge feature branch in main") {
		t.Errorf("Expected merge commit in log, got: %s", logStr)
	}

	t.Log("TestDeletedInFeatureAdd completed")
}

// TestDeletedInFeatureRm reproduces: deletedinfeature_rm.sh
// Tests merge conflict when file is modified in main but deleted in feature branch, resolved with rm
func TestDeletedInFeatureRm(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "go-git-deletedinfeature-rm-test")
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

	// Initialize repository
	execGit(t, exec.Command(binPath, "init"))

	// Create initial file with content
	testFile := "file.txt"
	initialContent := "line 1\nline 2\nline 3"
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add and commit initial version on main
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "init"))

	// Create feature branch
	execGit(t, exec.Command(binPath, "branch", "feature"))

	// Switch to feature branch
	execGit(t, exec.Command(binPath, "checkout", "feature"))

	// Modify file on feature branch
	featureContent := "line 1\nmodified in feature \nline 3"
	if err := os.WriteFile(testFile, []byte(featureContent), 0644); err != nil {
		t.Fatalf("Failed to modify test file on feature: %v", err)
	}

	// Add and commit modification on feature
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "modify in main"))

	// Switch back to main
	execGit(t, exec.Command(binPath, "checkout", "main"))

	// Delete file on main branch
	if err := os.Remove(testFile); err != nil {
		t.Fatalf("Failed to delete test file: %v", err)
	}

	// Add and commit deletion on main
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "delete"))

	// First merge attempt - should trigger conflict
	cmd := exec.Command(binPath, "merge", "feature")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	t.Logf("First merge output: %s", outputStr)

	// Should detect conflict
	if !strings.Contains(outputStr, "CONFLICT") && !strings.Contains(outputStr, "Conflict") {
		t.Logf("Warning: Expected conflict message, got: %s", outputStr)
	}

	// Use rm to resolve conflict
	execGit(t, exec.Command(binPath, "rm", testFile))

	// Second merge attempt - should complete without error
	cmd2 := exec.Command(binPath, "merge", "feature")
	output2, err2 := cmd2.CombinedOutput()
	outputStr2 := string(output2)
	t.Logf("Second merge output: %s", outputStr2)

	// Validate that second merge DOES NOT contain conflict messages (conflict should be resolved)
	if strings.Contains(outputStr2, "CONFLICT") || strings.Contains(outputStr2, "Conflict") {
		t.Errorf("Second merge should not contain conflict messages, got: %s", outputStr2)
	}
	if err2 != nil {
		t.Errorf("Second merge should succeed without error, got: %v", err2)
	}

	// Verify final state with log - should contain merge commit
	logCmd := exec.Command(binPath, "log")
	logOutput, logErr := logCmd.CombinedOutput()
	logStr := string(logOutput)
	t.Logf("Final log: %s", logStr)

	if logErr != nil {
		t.Errorf("Failed to get log: %v", logErr)
	}

	// Validate that merge commit appears in log
	if !strings.Contains(logStr, "Merge feature branch in main") {
		t.Errorf("Expected merge commit in log, got: %s", logStr)
	}

	t.Log("TestDeletedInFeatureRm completed")
}

// TestDeletedInMainAdd reproduces: deletedinmain_add.sh
// Tests merge conflict when file is deleted in main but modified in feature branch, resolved with add
func TestDeletedInMainAdd(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "go-git-deletedinmain-add-test")
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

	// Initialize repository
	execGit(t, exec.Command(binPath, "init"))

	// Create initial file with content
	testFile := "file.txt"
	initialContent := "line 1\nline 2\nline 3"
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add and commit initial version on main
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "init"))

	// Create feature branch
	execGit(t, exec.Command(binPath, "branch", "feature"))

	// Delete file on main branch
	if err := os.Remove(testFile); err != nil {
		t.Fatalf("Failed to delete test file: %v", err)
	}

	// Add and commit deletion on main
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "delete"))

	// Switch to feature branch
	execGit(t, exec.Command(binPath, "checkout", "feature"))

	// Modify same file on feature branch
	featureContent := "line 1\nmodified in feature \nline 3"
	if err := os.WriteFile(testFile, []byte(featureContent), 0644); err != nil {
		t.Fatalf("Failed to modify test file on feature: %v", err)
	}

	// Add and commit modification on feature
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "modify"))

	// Switch back to main
	execGit(t, exec.Command(binPath, "checkout", "main"))

	// First merge attempt - should trigger conflict
	cmd := exec.Command(binPath, "merge", "feature")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	t.Logf("First merge output: %s", outputStr)

	// Should detect conflict
	if !strings.Contains(outputStr, "CONFLICT") && !strings.Contains(outputStr, "Conflict") {
		t.Logf("Warning: Expected conflict message, got: %s", outputStr)
	}

	// Use add to resolve conflict
	execGit(t, exec.Command(binPath, "add", testFile))

	// Second merge attempt - should complete without error
	cmd2 := exec.Command(binPath, "merge", "feature")
	output2, err2 := cmd2.CombinedOutput()
	outputStr2 := string(output2)
	t.Logf("Second merge output: %s", outputStr2)

	// Validate that second merge DOES NOT contain conflict messages (conflict should be resolved)
	if strings.Contains(outputStr2, "CONFLICT") || strings.Contains(outputStr2, "Conflict") {
		t.Errorf("Second merge should not contain conflict messages, got: %s", outputStr2)
	}
	if err2 != nil {
		t.Errorf("Second merge should succeed without error, got: %v", err2)
	}

	// Verify final state with log - should contain merge commit
	logCmd := exec.Command(binPath, "log")
	logOutput, logErr := logCmd.CombinedOutput()
	logStr := string(logOutput)
	t.Logf("Final log: %s", logStr)

	if logErr != nil {
		t.Errorf("Failed to get log: %v", logErr)
	}

	// Validate that merge commit appears in log
	if !strings.Contains(logStr, "Merge feature branch in main") {
		t.Errorf("Expected merge commit in log, got: %s", logStr)
	}

	t.Log("TestDeletedInMainAdd completed")
}

// TestDeletedInMainRm reproduces: deletedinmain_rm.sh
// Tests merge conflict when file is deleted in main but modified in feature branch, resolved with rm
func TestDeletedInMainRm(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "go-git-deletedinmain-rm-test")
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

	// Initialize repository
	execGit(t, exec.Command(binPath, "init"))

	// Create initial file with content
	testFile := "file.txt"
	initialContent := "line 1\nline 2\nline 3"
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add and commit initial version on main
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "init"))

	// Create feature branch
	execGit(t, exec.Command(binPath, "branch", "feature"))

	// Delete file on main branch
	if err := os.Remove(testFile); err != nil {
		t.Fatalf("Failed to delete test file: %v", err)
	}

	// Add and commit deletion on main
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "delete"))

	// Switch to feature branch
	execGit(t, exec.Command(binPath, "checkout", "feature"))

	// Modify same file on feature branch
	featureContent := "line 1\nmodified in feature \nline 3"
	if err := os.WriteFile(testFile, []byte(featureContent), 0644); err != nil {
		t.Fatalf("Failed to modify test file on feature: %v", err)
	}

	// Add and commit modification on feature
	execGit(t, exec.Command(binPath, "add", testFile))
	execGit(t, exec.Command(binPath, "commit", "modify"))

	// Switch back to main
	execGit(t, exec.Command(binPath, "checkout", "main"))

	// First merge attempt - should trigger conflict
	cmd := exec.Command(binPath, "merge", "feature")
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	t.Logf("First merge output: %s", outputStr)

	// Should detect conflict
	if !strings.Contains(outputStr, "CONFLICT") && !strings.Contains(outputStr, "Conflict") {
		t.Logf("Warning: Expected conflict message, got: %s", outputStr)
	}

	// Use rm to resolve conflict
	execGit(t, exec.Command(binPath, "rm", testFile))

	// Second merge attempt - should complete without error
	cmd2 := exec.Command(binPath, "merge", "feature")
	output2, err2 := cmd2.CombinedOutput()
	outputStr2 := string(output2)
	t.Logf("Second merge output: %s", outputStr2)

	// Validate that second merge DOES NOT contain conflict messages (conflict should be resolved)
	if strings.Contains(outputStr2, "CONFLICT") || strings.Contains(outputStr2, "Conflict") {
		t.Errorf("Second merge should not contain conflict messages, got: %s", outputStr2)
	}
	if err2 != nil {
		t.Errorf("Second merge should succeed without error, got: %v", err2)
	}

	// Verify final state with log - should contain merge commit
	logCmd := exec.Command(binPath, "log")
	logOutput, logErr := logCmd.CombinedOutput()
	logStr := string(logOutput)
	t.Logf("Final log: %s", logStr)

	if logErr != nil {
		t.Errorf("Failed to get log: %v", logErr)
	}

	// Validate that merge commit appears in log
	if !strings.Contains(logStr, "Merge feature branch in main") {
		t.Errorf("Expected merge commit in log, got: %s", logStr)
	}

	t.Log("TestDeletedInMainRm completed")
}