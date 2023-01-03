package main

import (
	"time"
	"fmt"
	"log"
	"errors"
	"regexp"
	"strings"
	"golang.org/x/exp/slices"
)

// parse Command for FILTER arguments.
// i.e. +tag, -tag, created.after, created.before, 
// modified.after, modified.before
func parseFilter(args []string) (filter NoteFilter, rargs []string, err error) {
	for k, v := range args {
		// +tag
		if v[0] == '+' {
			filter.TagsInclude = append(filter.TagsInclude, v[1:])
			continue
		}

		// -tag
		if v[0] == '-' {
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
		r := regexp.MustCompile(`^\+\pL+$`)
		if r.MatchString(v) {
			if slices.Contains(posTags, v[1:]) {
				continue
			}
			posTags = append(posTags, v[1:])
			continue
		}

		// match -somestring
		r = regexp.MustCompile(`^\-\pL+$`)
		if r.MatchString(v) {
			if slices.Contains(negTags, v[1:]) {
				continue
			}
			negTags = append(negTags, v[1:])
			continue
		}

		rargs = args[k:]
		break
	}
	
	return
}