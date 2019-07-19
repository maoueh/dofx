package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var fitIDRuneCount = 9

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	if len(os.Args) <= 2 {
		fmt.Println("usage: dofx (stats|clean) <file1> <file2> ...")
		os.Exit(1)
	}

	command := os.Args[1]
	if command != "stats" && command != "clean" {
		fmt.Println("usage: dofx (stats|clean) <file1> <file2> ...")
		os.Exit(1)
	}

	for _, ofxFilepath := range os.Args[2:] {
		inDir := filepath.Dir(ofxFilepath)
		inBase := filepath.Base(ofxFilepath)
		inBaseExt := filepath.Ext(inBase)

		// Hopefully, the extension is not in the filename, ouch otherwise!
		inBaseNoExt := strings.Replace(inBase, inBaseExt, "", 1)

		inFile, err := os.Open(ofxFilepath)
		check(err)

		defer inFile.Close()

		reader := transform.NewReader(inFile, charmap.ISO8859_1.NewDecoder())

		if command == "clean" {
			outFilepath := filepath.Join(inDir, inBaseNoExt+"_cleaned"+inBaseExt)
			outFile, err := os.Create(outFilepath)
			check(err)

			defer outFile.Close()

			writer := transform.NewWriter(outFile, charmap.ISO8859_1.NewEncoder())

			dedup_fitid(bufio.NewScanner(reader), writer)
		} else if command == "stats" {
			print_stats(bufio.NewScanner(reader))
		}

		// update_fitid_in_range("20190301", "20190325", bufio.NewScanner(reader), writer)
	}
}

func print_stats(scanner *bufio.Scanner) {
	lowest := time.Time{}
	highest := time.Time{}
	fitids := map[string]bool{}
	fitidsToDate := map[string]time.Time{}
	fitDupIdsCount := map[string]int{}

	activeDatePosted := lowest
	for scanner.Scan() {
		line := scanner.Text()
		check(scanner.Err())

		if strings.HasPrefix(line, "<DTPOSTED>") {
			datePosted := extractTime(line[10:])
			if lowest.IsZero() || datePosted.Before(lowest) {
				lowest = datePosted
			}

			if highest.IsZero() || datePosted.After(highest) {
				highest = datePosted
			}

			activeDatePosted = datePosted
		}

		if strings.HasPrefix(line, "<FITID>") {
			id := line[7:]
			exists := fitids[id]

			if exists {
				fitDupIdsCount[id] = fitDupIdsCount[id] + 1
			}

			fitids[id] = true
			fitidsToDate[id] = activeDatePosted
		}
	}

	fmt.Printf("Oldest: %s\n", lowest)
	fmt.Printf("Newest %s\n", highest)

	if len(fitDupIdsCount) > 0 {
		fmt.Println()
		for id, count := range fitDupIdsCount {
			datePosted := fitidsToDate[id]

			fmt.Printf("Duplicate fitid: %s(%d) @ %s\n", id, count+1, datePosted)
		}
	}
}

func dedup_fitid(scanner *bufio.Scanner, writer io.Writer) {
	fmt.Println("Changing transaction with same id into different ones ...")
	fitids := map[string]bool{}

	for scanner.Scan() {
		line := scanner.Text()
		check(scanner.Err())

		if strings.HasPrefix(line, "<FITID>") {
			actual := line[7:]
			exists := fitids[actual]

			if exists {
				new := randomFitID()
				fmt.Printf("Performing transform %q -> %q\n", actual, new)
				line = "<FITID>" + new
			}

			fitids[actual] = true
		}

		fmt.Fprintln(writer, line)
	}
}

// update_fitid_in_range: date range is all inclusive (both ends).
func update_fitid_in_range(dateRangeStart, dateRangeEnd string, scanner *bufio.Scanner, writer io.Writer) {
	dateStart := extractTime(dateRangeStart)
	dateEnd := extractTime(dateRangeEnd)
	replaceNextFitID := false
	replacements := map[string]string{}

	for scanner.Scan() {
		line := scanner.Text()
		check(scanner.Err())

		if strings.HasPrefix(line, "<DTPOSTED>") {
			datePosted := extractTime(line[10:])
			if inRange(dateStart, dateEnd, datePosted) {
				fmt.Printf("Flagging next fitid to be replaced for date posted %q\n", datePosted)
				replaceNextFitID = true
			}
		}

		if strings.HasPrefix(line, "<FITID>") {
			id := line[7:]
			if replaceNextFitID {
				new, ok := replacements[id]
				if !ok {
					new = randomFitID()
				}

				replacements[id] = new

				fmt.Printf("Performing transform %q -> %q due to flag\n", id, new)
				line = "<FITID>" + new
			}

			replaceNextFitID = false
		}

		fmt.Fprintln(writer, line)
	}
}

func randomFitID() string {
	b := make([]rune, fitIDRuneCount)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func inRange(start, end, check time.Time) bool {
	if start.Equal(check) || end.Equal(check) {
		return true
	}

	return check.After(start) && check.Before(end)
}

func extractTime(dtposted string) time.Time {
	year, _ := strconv.ParseInt(dtposted[0:4], 10, 32)
	month, _ := strconv.ParseInt(dtposted[4:6], 10, 32)
	day, _ := strconv.ParseInt(dtposted[6:8], 10, 32)
	location, _ := time.LoadLocation("Local")

	return time.Date(int(year), time.Month(month), int(day), 0, 0, 0, 0, location)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
