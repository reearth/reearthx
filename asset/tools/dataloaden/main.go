package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	fmt.Println("Generating data loaders...")

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	parentDir := filepath.Dir(filepath.Dir(wd))
	fmt.Printf("Will generate data loaders in: %s\n", parentDir)

	if err := os.Chdir(parentDir); err != nil {
		panic(fmt.Errorf("failed to change directory to %s: %v", parentDir, err))
	}

	generateAssetLoader()

	generateGroupLoader()

	fmt.Println("Data loaders generated successfully!")
}

func generateAssetLoader() {
	fmt.Println("Generating AssetLoader...")

	cmd := exec.Command(
		"go", "run", "github.com/vektah/dataloaden",
		"AssetLoader",
		"github.com/eukarya-inc/reearthx/asset.AssetID",
		"*github.com/eukarya-inc/reearthx/asset.Asset",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error generating AssetLoader: %v\n", err)
		os.Exit(1)
	}
}

func generateGroupLoader() {
	fmt.Println("Generating GroupLoader...")

	cmd := exec.Command(
		"go", "run", "github.com/vektah/dataloaden",
		"GroupLoader",
		"github.com/eukarya-inc/reearthx/asset.GroupID",
		"*github.com/eukarya-inc/reearthx/asset.Group",
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error generating GroupLoader: %v\n", err)
		os.Exit(1)
	}
}
