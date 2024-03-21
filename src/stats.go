package main

import (
	"fmt"
	"sort"
	"time"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

const daysInLastSixMonths = 183
const weeksInLastSixMonths = daysInLastSixMonths / 7
const outOfRange = 99999

type column = []int

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

func calcOffset() int {
	var offset int
	weekday := time.Now().Weekday()

	switch weekday {
	case time.Sunday:
		offset = 7
	case time.Monday:
		offset = 6
	case time.Tuesday:
		offset = 5
	case time.Wednesday:
		offset = 4
	case time.Thursday:
		offset = 3
	case time.Friday:
		offset = 2
	case time.Saturday:
		offset = 1
	}

	return offset
}

func fillCommits(email string, path string, commits map[int]int) map[int]int {
	repo, err := git.PlainOpen(path)
	if err != nil {
		panic(err)
	}

	ref, err := repo.Head()
	if err != nil {
		panic(err)
	}

	iterator, err := repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		panic(err)
	}

	offset := calcOffset()
	err = iterator.ForEach(func(commit *object.Commit) error {
		daysAgo := countDaysSinceDate(commit.Author.When) + offset

		if commit.Author.Email != email {
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

func sortMapIntoSlices(m map[int]int) []int {
	var keys []int
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	return keys
}

func buildCols(keys []int, commits map[int]int) map[int]column {
	cols := make(map[int]column)
	col := column{}

	for _, k := range keys {
		week := int(k / 7)
		dayInWeek := k % 7

		if dayInWeek == 0 {
			col = column{}
		}

		col = append(col, commits[k])

		if dayInWeek == 6 {
			cols[week] = col
		}
	}

	return cols
}

func printMonths() {
	week := getBegginingOfDay(time.Now()).Add(-(daysInLastSixMonths * time.Hour * 24))
	month := week.Month()

	fmt.Printf("         ")

	for {
		if week.Month() != month {
			fmt.Printf("%s ", week.Month().String()[:3])
			month = week.Month()
		} else {
			fmt.Printf("    ")
		}

		week = week.Add(7 * time.Hour * 24)
		if week.After(time.Now()) {
			break
		}
	}
	fmt.Printf("\n")
}

func printDayCol(day int) {
	out := "     "

	switch day {
	case 5:
		out = " Mon "
	case 3:
		out = " Wed "
	case 1:
		out = " Fri"
	}

	fmt.Printf(out)
}

func printCell(val int, today bool) {
	escape := "\033[0;37;30m"

	switch {
	case val > 0 && val < 5:
		escape = "\033[1;30;47m"
	case val >= 5 && val < 10:
		escape = "\033[1;30;43m"
	case val >= 10:
		escape = "\033[1;30;42m"
	}

	if today {
		escape = "\033[1;37;45m"
	}

	if val == 0 {
		fmt.Printf(escape + "  - " + "\033[0m")
		return
	}

	str := "  %d "
	switch {
	case val >= 10:
		str = " %d "
	case val >= 100:
		str = "%d "
	}

	fmt.Printf(escape+str+"\033[0m", val)
}

func printCells(cols map[int]column) {
	printMonths()
	for j := 6; j >= 0; j-- {
		for i := weeksInLastSixMonths + 1; i >= 0; i-- {
			if i == weeksInLastSixMonths+1 {
				printDayCol(j)
			}
			if col, ok := cols[i]; ok {
				//special case today
				if i == 0 && j == calcOffset()-1 {
					printCell(col[j], true)
					continue
				} else {
					if len(col) > j {
						printCell(col[j], false)
						continue
					}
				}
			}
			printCell(0, false)
		}
		fmt.Printf("\n")
	}
}





// 	printMonths()

//   fmt.Println(cols)

// 	for i := 6; i >= 0; i-- {
// 		for j := weeksInLastSixMonths + 1; j >= 0; j-- {
// 			if j == weeksInLastSixMonths+1 {
// 				printDayCol(i)
// 			}
// 			if col, ok := cols[i]; ok {
// 				// special case today
// 				if j == 0 && i == calcOffset()-1 {
// 					printCell(col[i], true)
// 					continue
// 				} else {
// 					if len(col) > i {
// 						printCell(col[i], false)
// 						continue
// 					}
// 				}
// 			}
// 			printCell(0, false)
// 		}
// 		fmt.Printf("\n")
// 	}
// }

func printCommitsStats(commits map[int]int) {
	keys := sortMapIntoSlices(commits)
	cols := buildCols(keys, commits)
	printCells(cols)
}

func stats(email string) {
	commits := processRepositories(email)
	printCommitsStats(commits)
}
