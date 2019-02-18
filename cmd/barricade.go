// Package cmd implements the command line interface of barricade
package cmd

import (
	"flag"
	"log"
)

func main() {
	constrainsFile := flag.String("c", "", "path of the constraints file")
	flag.Parse()

	if *constrainsFile == "" {
		log.Fatal("You need to specify a constraint file, please use -c flag")
	}
}
