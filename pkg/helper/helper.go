package helper

import (
	"flag"
	"fmt"
)

func LogTags() {
	fmt.Println()
	fmt.Println("Parameters:")
	flag.VisitAll(func(f *flag.Flag) {
		fmt.Printf("-%s: %s\n", f.Name, f.Value)
	})
	fmt.Println()
}
