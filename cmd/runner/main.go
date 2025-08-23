package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"

	"golang.org/x/sync/errgroup"

	"github.com/pegov/simpci/internal/runner"
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

	target := "test"
	state := runner.NewState()
	state.AddTarget(target, "./docker/template.sh")

	path, ok := state.Script(target)
	if !ok {
		fmt.Println("target not found:", target)
		os.Exit(1)
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := dockerBuild(ctx); err != nil {
			return fmt.Errorf("docker build: %w", err)
		}
		if err := dockerRun(ctx, path); err != nil {
			return fmt.Errorf("docker run: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		log.Println("g wait err:", err)
	}
}

type State struct {
	M map[string]string
}

func dockerBuild(ctx context.Context) error {
	cmd := exec.CommandContext(
		ctx,
		"docker",
		"build",
		"-t", "simpci-template",
		"-f", "./docker/template.Dockerfile",
		".",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("failed to build docker template: %v", err)
	}
	if !cmd.ProcessState.Success() {
		log.Fatalf("docker command's exit code != 0")
	}

	return nil
}

func dockerRun(ctx context.Context, entrypointPath string) error {
	cmd := exec.CommandContext(
		ctx,
		"docker",
		"run",
		"--rm",
		"--pull=never",
		"-v", fmt.Sprintf("%s:/tmp/entrypoint.sh", entrypointPath),
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
