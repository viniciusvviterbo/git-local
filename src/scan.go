package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"strings"
)

func folderMatchesIgnoredFolders(folder string, ignoredFolders []string) bool {
    for _, substr := range ignoredFolders {
        if strings.Contains(folder, substr) {
            return true
        }
    }
    return false
}

// Recursively searches `folder` for the parent folders of `.git` directories
func scanGitFolders(folders []string, folder string, ignoreFolders []string) []string {
  if folderMatchesIgnoredFolders(folder, ignoreFolders) {
    return folders
  }

	folder = strings.TrimSuffix(folder, "/")

	f, err := os.Open(folder)
	if err != nil {
		log.Fatal(err)
	}
  files, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		log.Fatal(err)
	}

	var path string

	for _, file := range files {
		if file.IsDir() {
			path = folder + "/" + file.Name()
			if file.Name() == ".git" {
				path = strings.TrimSuffix(path, "/.git")
				fmt.Println("\t" + path)
        folders = append(folders, path)
				continue
			}
			if file.Name() == "vendor" || file.Name() == "node_modules" {
				continue
			}
			folders = scanGitFolders(folders, path, ignoreFolders)
		}
	}

	return folders
}

// Returns the path to the user's `.gitlocalstats` file
func getDotFilePath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	dotFile := usr.HomeDir + "/.gitlocalstats"

	return dotFile
}

// Returns the path to the user's `.gitlocalstatsignore` file
func getDotFileIgnorePath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	dotFileIgnore := usr.HomeDir + "/.gitlocalstatsignore"

	return dotFileIgnore
}

// Opens the file at `filePath`. Creates it if not existing
func openFile(filePath string) *os.File {
	f, err := os.OpenFile(filePath, os.O_APPEND, 0755)
	if err != nil {
		if os.IsNotExist(err) {
			_, err = os.Create(filePath)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err)
		}
	}
	return f
}

// Given file at `filePath`, returns a slice of strings in wich each line is an instance in the slice
func parseFileLinesToSlice(filePath string) []string {
	f := openFile(filePath)
	defer f.Close()

	var lines []string
	johnLennon := bufio.NewScanner(f)
	for johnLennon.Scan() {
		lines = append(lines, johnLennon.Text())
	}
	if err := johnLennon.Err(); err != nil {
		if err != io.EOF {
			panic(err)
		}
	}

	return lines
}

// Returns `true` if `slice` contains `value`
func sliceContains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// Adds elements from `new` into `existing` if not already there
func joinSlices(new []string, existing []string) []string {
	for _, i := range new {
		if !sliceContains(existing, i) {
			existing = append(existing, i)
		}
	}
	return existing
}

// Writes content from `repos` into file at `filePath`. Existing content is overwritten
func dumpStringsSliceToFile(repos []string, filePath string) {
	content := strings.Join(repos, "\n")
	os.WriteFile(filePath, []byte(content), 0755)
	ioutil.WriteFile(filePath, []byte(content), 0755)
}

// Stores `newRepos` lines into file at `filePath`
func addNewSliceElementsToFile(filePath string, newRepos []string) {
	existingRepos := parseFileLinesToSlice(filePath)
	repos := joinSlices(newRepos, existingRepos)
	dumpStringsSliceToFile(repos, filePath)
}

// Starts the recursive search of git repositories
func recursiveScanFolder(folder string) []string {
  fileIgnorePath := getDotFileIgnorePath()
	ignoreFolders  := parseFileLinesToSlice(fileIgnorePath)
  fmt.Println(ignoreFolders)
	return scanGitFolders(make([]string, 0), folder, ignoreFolders)
}

// Scans for a new folder in git repositories
func scan(folder string) {
	fmt.Printf("Found folders:\n")
	var repositories []string = recursiveScanFolder(folder)
	filePath := getDotFilePath()
	addNewSliceElementsToFile(filePath, repositories)
	fmt.Printf("\n\nSuccessfully added\n\n")
}
