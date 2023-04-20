package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Faione/ServerExporter/collector"
)

func main() {
	if err := collector.New().Execute(); err != nil && err != http.ErrServerClosed {
		fmt.Println("err: ", err)
		os.Exit(1)
	}
}
