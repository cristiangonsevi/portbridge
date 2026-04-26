package main

import (
	"portbridge/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
