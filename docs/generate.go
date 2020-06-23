package main

import (
	"github.com/nuts-foundation/nuts-go-core/docs"
	"github.com/nuts-foundation/nuts-registry/engine"
)

func main() {
	docs.GenerateConfigOptionsDocs("README_options.rst", engine.NewRegistryEngine().FlagSet)
}
