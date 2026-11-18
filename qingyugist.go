package main

import (
	"github.com/LiuShen-Fork/Qingyu-Gist/internal/cli"
	"os"
)

func main() {
	if err := cli.App(); err != nil {
		os.Exit(1)
	}
}
