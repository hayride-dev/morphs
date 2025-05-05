package main

import (
	"fmt"
	"io"

	"net/http"

	"github.com/hayride-dev/bindings/go/imports/net/http/transport"
)

func main() {
	client := &http.Client{
		Transport: transport.NewWasiRoundTripper(),
	}

	resp, err := client.Get("https://postman-echo.com/get?foo1=bar1&foo2=bar2")
	if err != nil {
		fmt.Println("error making GET request:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("unexpected status code:", resp.StatusCode)
		return
	}
	fmt.Println("GET request successful:", resp.Status)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error reading response body:", err)
		return
	}

	fmt.Println(string(body))
}
