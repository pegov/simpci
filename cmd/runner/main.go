package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", "127.0.0.1:9095")
	if err != nil {
		log.Fatalf("failed to connect to server: %v", err)
	}

	go func() {
		<-signals
		cancel()
		conn.Close()
	}()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		for {
			time.Sleep(time.Second * 2)
			log.Println("writing ping")
			if _, err := conn.Write([]byte("ping")); err != nil {
				select {
				case <-ctx.Done():
					return nil
				default:
					return err
				}
			}
		}
	})

	if err := g.Wait(); err != nil {
		log.Println("g wait err:", err)
	}
}
