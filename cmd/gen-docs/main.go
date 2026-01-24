package main

import (
	"log"
	"os"

	"github.com/vgaro/yotocli/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	// We need a dummy rootCmd to generate docs from
	// Since cmd.Execute() runs the command, we need to access the rootCmd variable.
	// We'll export it or use a helper in the cmd package.
	
	dir := "./docs/commands"
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatal(err)
	}

	err := doc.GenMarkdownTree(cmd.RootCmd(), dir)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Documentation generated in %s\n", dir)
}

