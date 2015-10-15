package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`([\w]+)\s+([\d\w-]+)\s+([-\*])`)

var module = flag.String("m", "default", "module to clean up versions for")

func main() {
	flag.Parse()
	log.Printf("Listing versions for module %q...", *module)
	out, err := exec.Command("gcloud", "preview", "app", "modules", "list", *module).Output()
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range re.FindAllStringSubmatch(string(out), -1) {
		version := m[2]

		if strings.HasPrefix(version, "ah-builtin") {
			log.Printf("Skipping magic version %q.", version)
			continue
		}

		if m[3] == "*" {
			log.Printf("Skipping default version %q.", version)
			continue
		}

		log.Printf("Deleting version %q...", version)
		out, err := exec.Command("gcloud", "preview", "app", "modules", "delete", *module, "--version", version, "--quiet").CombinedOutput()
		if err == nil {
			continue
		}
		log.Print(err, " output follows:")
		fmt.Println(string(out))
	}
}
