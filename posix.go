package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// https://gobyexample.com/signals
func handlePosix() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	<-done
	fmt.Println("exiting")
}
