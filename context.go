//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("done")
				return
			default:
				fmt.Println("running...")
				time.Sleep(time.Second)
			}
		}
	}()
	time.Sleep(time.Second * 3)
	cancel()
}
