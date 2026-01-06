package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func CollectGOFiles() {
	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	outputFile := "all_go_files.txt"
	out, err := os.Create(outputFile)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Process only .go files
		if !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		header := fmt.Sprintf("// %s\n", filepath.ToSlash(relPath))

		_, err = out.WriteString(header)
		if err != nil {
			return err
		}

		_, err = out.Write(content)
		if err != nil {
			return err
		}

		// Add spacing between files
		_, err = out.WriteString("\n\n")
		if err != nil {
			return err
		}

		fmt.Println("Collected:", relPath)
		return nil
	})

	if err != nil {
		panic(err)
	}

	fmt.Println("\n✅ All .go files copied into:", outputFile)
}
