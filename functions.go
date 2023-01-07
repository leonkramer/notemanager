package main

import (
	"time"
	"fmt"
	"log"
	"errors"
	"regexp"
	"strings"
	"os/exec"
	"os"
	"sort"
	"golang.org/x/exp/slices"
	"github.com/google/uuid"
)

// parse Command for FILTER arguments.
// i.e. +tag, -tag, created.after, created.before, 
// modified.after, modified.before
func parseFilter(args []string) (filter NoteFilter, rargs []string, err error) {
	for k, v := range args {
		// match +somestring tag
		r := regexp.MustCompile(`^\+[\pL0-9]+$`)
		if r.MatchString(v) {
			if slices.Contains(filter.TagsInclude, v[1:]) {
				continue
			}
			filter.TagsInclude = append(filter.TagsInclude, v[1:])
			continue
		}

		// match -somestring tag
		r = regexp.MustCompile(`^\-[\pL0-9]+$`)
		if r.MatchString(v) {
			if slices.Contains(filter.TagsExclude, v[1:]) {
				continue
			}
			filter.TagsExclude = append(filter.TagsExclude, v[1:])
			continue
		}

		// created.after:
		if len(v) > 14 {
			if v[0:14] == "created.after:" {
				ts := v[14:]
				filter.CreatedAfter, err = parseTimestamp(ts)
				if err != nil {
					log.Fatalf("Timestamp parsing failed: %s", ts)
				}
				continue
			}
		}

		// created.before:
		if len(v) > 15 {
			if v[0:15] == "created.before:" {
				ts := v[15:]
				filter.CreatedBefore, err = parseTimestamp(ts)
				if err != nil {
					log.Fatalf("Timestamp parsing failed: %s", ts)
				}
				continue
			}
		}

		// modified.after:
		if len(v) > 15 {
			if v[0:15] == "modified.after:" {
				ts := v[15:]
				filter.ModifiedAfter, err = parseTimestamp(ts)
				if err != nil {
					log.Fatalf("Timestamp parsing failed: %s", ts)
				}
				continue
			}
		}

		// modified.before:
		if len(v) > 16 {
			if v[0:16] == "modified.before:" {
				ts := v[16:]
				filter.ModifiedBefore, err = parseTimestamp(ts)
				if err != nil {
					log.Fatalf("Timestamp parsing failed: %s", ts)
				}
				continue
			}
		}
		rargs = args[k:]
		return
	}
	
	return
}

// parses a string with format YYYY-DD-MM HH:MM:SS and some shorter
// forms of it into a time.Time variable.
// At least a date must be given, if a time is given, at least HH:MM
// must be provided.
func parseTimestamp(str string) (ts time.Time, err error) {
	// explode date in two parts, date and time
	explode := strings.Split(str, " ")

	var format string
	// expect at least 1 part, but at most 2
	if len(explode) < 1 || len(explode) > 2 {
		err = errors.New("Invalid date syntax. Expecting YYYY-MM-DD HH:MM:SS")
	}

	for k, v := range explode {
		// date parsing. Allowed YYYY, YYYY-MM, YYYY-MM-DD with separators / or -
		if k == 0 {
			r := regexp.MustCompile(`^[0-9]{8}$`)
			if r.MatchString(v) {
				//2006-01-02
				format = "20060102"
			}
		
			r = regexp.MustCompile(`^[0-9]{4}/[0-9]{2}/[0-9]{2}$`)
			if r.MatchString(v) {
				format = "2006/01/02"
			}

			r = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}$`)
			if r.MatchString(v) {
				format = "2006-01-02"
			}
		}

		if k == 1 {
			r := regexp.MustCompile(`^[0-9]{6}$`)
			if r.MatchString(v) {
				// 15:04:05
				format += " 150405"
			}

			r = regexp.MustCompile(`^[0-9]{4}$`)
			if r.MatchString(v) {
				format += " 1504"
			}

			r = regexp.MustCompile(`^[0-9]{2}:[0-9]{2}:[0-9]{2}$`)
			if r.MatchString(v) {
				format += " 15:04:05"
			}

			r = regexp.MustCompile(`^[0-9]{2}:[0-9]{2}$`)
			if r.MatchString(v) {
				format += " 15:04"
			}
		}
	}
    
	ts, err = time.ParseInLocation(format, str, time.Now().Location())
	return
}

// parse command for supplied tag modifiers, i.e. +something -something etc
func parseTagModifiers(args []string) (posTags []string, negTags []string, rargs []string, err error) {
	for k, v := range args {
		// match +somestring
		r := regexp.MustCompile(`^\+[\pL0-9]+$`)
		if r.MatchString(v) {
			if slices.Contains(posTags, v[1:]) {
				continue
			}
			posTags = append(posTags, v[1:])
			continue
		}

		// match -somestring
		r = regexp.MustCompile(`^\-[\pL0-9]+$`)
		if r.MatchString(v) {
			if slices.Contains(negTags, v[1:]) {
				continue
			}
			negTags = append(negTags, v[1:])
			continue
		}

		// argument does not match
		// we are done
		rargs = args[k:]
		break
	}
	
	return
}

// Open the Editor and edit file filepath
func runEditor(filepath string) (err error) {
	cmd := exec.Command(notemanager.Editor, filepath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	
	err = cmd.Run()	
	return
}

func notes(filter NoteFilter) (notes []Note, err error) {
	files, err := os.ReadDir(notemanager.NoteDir)
	if err != nil {
        log.Fatal(err)
    }

	for _, file := range files {
		if file.IsDir() == false {
			continue
		}

		noteId := uuid.MustParse(file.Name())

		note, err := loadNote(noteId.String())
		if (err != nil) {
			log.Fatal(err)
		}

		matches, err := note.MatchesFilter(filter)
		if err != nil {
			log.Fatal(err)
		}
		if matches {
			notes = append(notes, note)
		}
    }
	
	return
}

func sortNotes(notes []Note) (ret []Note, err error) {
	// sort notes by DateCreated ASC
	sort.Slice(notes, func(a int, b int) bool {
		return notes[a].DateCreated.String() < notes[b].DateCreated.String()
	})

	fields := []string{
		"id",
		"tags",
		"title",
		"created",
	}

	var output [][]string
	maxLength := make([]int, len(fields))

	if len(notes) > 0 {
		// build 2-dimensional slice of notes.
		// one index per note and another per field.
		// also obtain the maximum length of entry to get a clean table output.
		for _, n := range notes {
			// m = 
			row := make([]string, 0)
			var s string
			for j, field := range fields {	
				switch field {
					case "id":
						s = n.ShortId()
					
					case "tags":
						s = strings.Join(n.Tags, ",")

					case "title":
						s = n.Title

					case "created":
						s = n.DateCreated.Local().Format(notemanager.OutputTimeFormatShort)
				}
				row = append(row, s)

				// update max length slice
				if maxLength[j] < len(s) {
					maxLength[j] = len(s)
				}
			}
			output = append(output, row)
		}


		var str string
		for k, v := range fields {
			str += fmt.Sprintf("%-*s", maxLength[k]+2, v)
		}
		str += "\n"
		for k, _ := range fields {
			str += fmt.Sprintf("%-*s", maxLength[k]+2, "--")
		}
		str += "\n"
		for _, v := range output {
			//fmt.Println(v)
			for kk, vv:= range v {
				str += fmt.Sprintf("%-*s", maxLength[kk]+2, vv)
			}
			str += "\n"
		}

		fmt.Printf("%s", str)
	}

	return
}