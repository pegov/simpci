package main

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	g, gCtx := errgroup.WithContext(ctx)
	l, err := net.Listen("tcp", "127.0.0.1:9095")
	tcpL := l.(*net.TCPListener)
	if err != nil {
		log.Fatalf("net listen: %v", err)
	}
	defer l.Close()
	g.Go(func() error {
		return server(gCtx, tcpL)
	})

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	go func() {
		<-signals
		cancel()
		l.Close()
	}()

	if err := g.Wait(); err != nil {
		log.Println(err)
	}
}

func server(ctx context.Context, l *net.TCPListener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				log.Println("listener shutdown")
				return nil
			default:
				log.Println("listener error:", err)
				return errors.New("tcp error")
			}
		}
		defer conn.Close()
	}
}
