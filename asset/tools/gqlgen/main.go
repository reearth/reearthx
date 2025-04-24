package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen/config"
)

func main() {
	assetDir, err := findAssetDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := os.Chdir(assetDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error changing to asset directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generating GraphQL code in directory:", assetDir)

	cfg, err := config.LoadConfig("gqlgen.yml")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	ensureDir(filepath.Join(assetDir, "graph", "generated"))
	ensureDir(filepath.Join(assetDir, "graph", "model"))
	ensureDir(filepath.Join(assetDir, "graph", "resolver"))

	err = api.Generate(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating code: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("GraphQL code generated successfully!")
}

func findAssetDir() (string, error) {
	if _, err := os.Stat("asset.graphql"); err == nil {
		return ".", nil
	}

	assetGraphqlPath := "asset.graphql"
	if _, err := os.Stat(assetGraphqlPath); err == nil {
		dir, _ := os.Getwd()
		return dir, nil
	}

	assetDir := "asset"
	assetGraphqlPath = filepath.Join(assetDir, "asset.graphql")
	if _, err := os.Stat(assetGraphqlPath); err == nil {
		absAssetDir, err := filepath.Abs(assetDir)
		if err != nil {
			return "", err
		}
		return absAssetDir, nil
	}

	return "", fmt.Errorf("cannot find asset directory containing asset.graphql")
}

func ensureDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
}
