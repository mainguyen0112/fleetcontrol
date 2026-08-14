package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mainguyen0112/fleetcontrol/api/gen"
)

func main() {
	client, err := gen.NewClientWithResponses("http://localhost:8080")
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to create client:", err)
		os.Exit(1)
	}

	resp, err := client.GetHealthWithResponse(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "health check failed:", err)
		os.Exit(1)
	}

	fmt.Printf("status code: %d\n", resp.StatusCode())
	fmt.Println(string(resp.Body))
}