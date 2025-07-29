package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"time"

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
	if err != nil {
		log.Fatalf("net listen: %v", err)
	}
	tcpL := l.(*net.TCPListener)
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

	log.Println("wait")
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

		log.Println("handling conn")
		// TODO: wait for it
		go func() {
			if err := handleConn(ctx, conn); err != nil {
				log.Println("handle conn err:", err)
			}
		}()
	}
}

func handleConn(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
			return fmt.Errorf("set read deadline: %w", err)
		}

		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				log.Println("conn was closed")
				break
			}
			if !errors.Is(err, io.EOF) {
				select {
				case <-ctx.Done():
					return nil
				default:
					return fmt.Errorf("conn read: %w", err)
				}
			}
		}

		log.Println("buf:", string(buf[:n]))
	}

	return nil
}
