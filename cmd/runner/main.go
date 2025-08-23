package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"os/signal"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	go func() {
		<-signals
		cancel()
	}()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return dockerRun(ctx)
	})

	if err := g.Wait(); err != nil {
		log.Println("g wait err:", err)
	}
}

func dockerRun(ctx context.Context) error {
	cmd := exec.CommandContext(
		ctx,
		"docker",
		"run",
		"--rm",
		"--pull=never",
		"-v", "./docker/:/tmp/",
		"simpci-template",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("failed to run docker cmd: %v", err)
	}
	if !cmd.ProcessState.Success() {
		log.Fatalf("docker command's exit code != 0")
	}

	return nil
}
