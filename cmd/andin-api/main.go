package main

import (
	"fmt"

	"github.com/ubipo/andin-api/internal/api"
)

func main() {
	fmt.Println("Start andin api server")

	api.Serve()
}
