# go-git

A simple Git implementation written in Go. This project provides a command-line interface that mimics core Git functionality, including repository initialization, file staging, committing, branching, merging, and conflict detection.

## Features

- **Repository Management**: Initialize new Git repositories
- **File Tracking**: Add and stage files for commit
- **Commits**: Create commits with messages
- **Branching**: Create, list, and delete branches
- **Checkout**: Switch between branches
- **Merging**: Merge branches with conflict detection
  - Supports `ours`, `theirs`, and recursive (default) merge strategies
  - Detects conflicts when the same file is modified in both branches
- **Restore**: Recover files from HEAD (staged or unstaged)
- **Status**: View working tree status
- **Storage**: SQLite-based persistent storage for commits, objects, and references

## Installation

### Prerequisites

- Go 1.26 or later

### Build from Source

```bash
# Clone the repository
git clone https://github.com/Clorter/go-git.git
cd go-git

# Build the binary
go build -o go-git .

# Verify installation
./go-git --help
```

## Usage

```
go-git - A simple git implementation

Usage: go-git [--help] <command> [<args>]

Commands:
  init                    Initialize a git repository
  add <file>             Stage a file for commit
  commit <message>       Commit staged changes
  push                   Push commits to remote (not implemented)
  status                 Show working tree status
  checkout <branch>      Switch branches
  branch [-d <name>]     List branches, create new branch, or delete with -d
  merge [-s <strategy>]  Merge a branch (strategy optional)
  restore [--staged] <file>  Restore file from HEAD

Options:
  --help, -h             Show this help message
```


## Project Structure

```
go-git/
├── main.go              # CLI entry point and command routing
├── go.mod               # Go module definition
├── go.sum               # Go dependencies checksum
├── README.md            # This file
├── core/
│   ├── commands.go      # Core command implementations
│   ├── merge.go         # Merge logic and conflict detection
│   └── working_tree.go  # Working tree transformation
├── storage/
│   ├── storage.go       # Storage interface
│   ├── sqlite.go        # SQLite storage implementation
│   └── disk.go          # Disk storage implementation
├── types/
│   └── types.go         # Type definitions (Object, Commit, ConflictFile)
└── tests/
    ├── merge_test.go    # Merge and conflict detection tests
    ├── pipeline_test.go # Integration tests
    ├── db_test.go       # Database tests
    └── disk_test.go     # Disk storage tests
```


## License

This project is open source software.
