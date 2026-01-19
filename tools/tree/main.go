// tools/tree/main.go
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	maxDepth = flag.Int("depth", 8, "max depth")
	outFile  = flag.String("out", "folder_tree.txt", "output file name")
)

func main() {
	flag.Parse()

	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	out, err := os.Create(*outFile)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	fmt.Fprintf(out, "Root: %s\n\n", root)

	printTree(out, root, ".", 0)

	fmt.Println("✅ Folder tree written to:", *outFile)
}

func printTree(out *os.File, root, rel string, depth int) {
	if depth > *maxDepth {
		return
	}

	abs := filepath.Join(root, rel)

	entries, err := os.ReadDir(abs)
	if err != nil {
		return
	}

	// Sort stable: dirs first then files
	sort.Slice(entries, func(i, j int) bool {
		di := entries[i].IsDir()
		dj := entries[j].IsDir()
		if di != dj {
			return di
		}
		return entries[i].Name() < entries[j].Name()
	})

	for i, e := range entries {
		name := e.Name()

		if shouldSkip(name, e) {
			continue
		}

		isLast := i == len(entries)-1

		prefix := strings.Repeat("│   ", depth)
		if depth > 0 {
			if isLast {
				prefix = strings.Repeat("│   ", depth-1) + "└── "
			} else {
				prefix = strings.Repeat("│   ", depth-1) + "├── "
			}
		}

		fmt.Fprintf(out, "%s%s\n", prefix, name)

		if e.IsDir() {
			printTree(out, root, filepath.Join(rel, name), depth+1)
		}
	}
}

func shouldSkip(name string, e fs.DirEntry) bool {
	// skip hidden folders/files
	if strings.HasPrefix(name, ".") {
		return true
	}

	// common junk
	skip := map[string]bool{
		"node_modules": true,
		"vendor":       true,
		"dist":         true,
		"build":        true,
		"tmp":          true,
		"coverage":     true,
		"bin":          true,
		"out":          true,
	}

	// skip git stuff even if not hidden
	if name == ".git" {
		return true
	}
	if skip[name] {
		return true
	}

	return false
}
