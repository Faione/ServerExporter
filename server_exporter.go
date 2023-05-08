package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Faione/ServerExporter/cmd"
)

func main() {
	if err := cmd.New().Execute(); err != nil && err != http.ErrServerClosed {
		fmt.Println("err: ", err)
		os.Exit(1)
	}
}
