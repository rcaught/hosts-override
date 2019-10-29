package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
)

func clearScreen() {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}

func waitUntilExit() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)

		done <- true
	}()

	<-done
}

func maybeIP(value *string) *string {
	v := net.ParseIP(*value)

	four := v.To4()
	six := v.To16()
	maybeIP := ""

	if four != nil {
		maybeIP = four.String()
	} else if six != nil {
		maybeIP = six.String()
	}

	return &maybeIP
}
