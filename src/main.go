package main

import (
	"flag"
)

func main() {
	var folder string 
	var email string 

	flag.StringVar(&folder, "add", "", "Add a new folder to scan for git repositories")
	flag.StringVar(&email, "email", "your@email.com", "Sets the email address to scan for")
	flag.Parse()

	if folder != "" {
		Scan(folder)
		return
	}

	Stats(email)
}
