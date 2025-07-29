package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("3 args")
	}

	dataDir := os.Args[1]
	repoDir := os.Args[2]

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoDir
	var stdoutBuffer bytes.Buffer
	cmd.Stdout = &stdoutBuffer
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("failed to run git cmd: %v", err)
	}
	if !cmd.ProcessState.Success() {
		log.Fatalf("git command's exit code != 0")
	}

	if err := os.MkdirAll(dataDir, 0777); err != nil {
		log.Fatalf("failed to mkdir all data dir: %v", err)
	}

	commitPath := path.Join(dataDir, ".commit")
	if err := os.WriteFile(commitPath, stdoutBuffer.Bytes(), 0777); err != nil {
		log.Fatalf("failed to write to .commit file: %v", err)
	}
}
