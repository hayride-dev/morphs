//go:build llama3

package main

import (
	"github.com/hayride-dev/bindings/go/hayride/ai/models"
	"github.com/hayride-dev/bindings/go/hayride/ai/models/export"
)

func constructor() (models.Format, error) {
	return &llama3{}, nil
}

func init() {
	export.Format(constructor)
}

func main() {}
