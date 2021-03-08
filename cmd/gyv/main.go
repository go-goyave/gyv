package main

import "os"

type exitCode int

const (
	exitOK     exitCode = 0
	exitError  exitCode = 1
	exitCancel exitCode = 2
)

func main() {
	os.Exit(int(mainRun()))
}

func mainRun() exitCode {
	return exitOK
}
