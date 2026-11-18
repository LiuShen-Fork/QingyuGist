package main

import (
	"github.com/LiuShen-Fork/QingyuGist/internal/cli"
	"os"
)

func main() {
	if err := cli.App(); err != nil {
		os.Exit(1)
	}
}
