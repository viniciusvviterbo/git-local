package main

import (
	"time"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

const daysInLastSixMonths = 183
const outOfRange = 99999

func getBegginingOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	startOfDay := time.Date(year, month, day, 0, 0, 0, 0, t.Location())
	return startOfDay
}

func countDaysSinceDate(date time.Time) int {
	days := 0
	now := getBegginingOfDay(time.Now())
	for date.Before(now) {
		date = date.Add(time.Hour * 24)
		days++
		if days > daysInLastSixMonths {
			return outOfRange
		}
	}
	return days
}

func fillCommits(email string, path string, commits map[int]int) map[int]int {
	repo, err := git.PlainOpen(path)
	if err != nil {
		panic(err)
	}

	ref, err := repo.HEAD
	if err != nil {
		panic(err)
	}

	iterator, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		panic(err)
	}

	offset := calcOffset()
	err = iterator.ForEach(func (c *object.Commit) error {
		daysAgo := countDaysSinceDate(c.Author.When) + offset

		if c.Author.Email != email {
			return nil
		}

		if daysAgo != outOfRange {
			commits[daysAgo]++
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	return commits
}

func processRepositories(email string) map[int]int {
	filePath := getDotFilePath()
	repos := parseFileLinesToSlice(filePath)
	daysInMap := daysInLastSixMonths
	
	commits := make(map[int]int, daysInMap)
	for i := daysInMap; i > 0; i-- {
		commits[i] = 0
	}

	for _, path := range repos {
		commits = fillCommits(email, path, commits)
	}

	return commits
}

func Stats(email string) {
	commits := processRepositories(email)
	printCommitsStats(commits)
}
