package main

import (
	"time"
	"fmt"
	"log"
	"errors"
	"regexp"
	"strings"
)

func parseFilter(args []string) (filter NoteFilter, err error) {
	for _, v := range args {
		if v[0] == '+' {
			filter.TagsInclude = append(filter.TagsInclude, v[1:])
			/* // copy last element to current
			args[k] = args[len(args)-1]
			// remove last element
			args = args[:len(args)-1]
			goto RESTART */
			continue
		}

		if v[0] == '-' {
			filter.TagsExclude = append(filter.TagsExclude, v[1:])
		/* 	args[k] = args[len(args)-1]
			args = args[:len(args)-1]
			goto RESTART */
			continue
		}

		if len(v) > 14 {
			if v[0:14] == "created.after:" {
				fmt.Println("Found created.after:", v)
				date, err  := parseTimestamp(v[14:])
				//tmp, err := time.Parse("2006-01-02 15:04:05", v[14:])
				//tmp = tmp.Local()
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println("date", date.Format("2006-01-02 15:04:05"))
				filter.CreatedAfter = time.Now()
				continue
			}
		}

		if len(v) > 15 {
			if v[0:15] == "created.before:" {
				fmt.Println("Found created.before:")
				filter.CreatedBefore = time.Now()
				continue
			}
		}

		if len(v) > 15 {
			if v[0:15] == "modified.after:" {
				fmt.Println("Found modified.after:")
				filter.ModifiedAfter = time.Now()
				continue
			}
		}

		if len(v) > 16 {
			if v[0:16] == "modified.before:" {
				fmt.Println("Found modified.before:")
				filter.ModifiedBefore = time.Now()
				continue
			}
		}
	}
	
	return
}

// parses the date format YYYY-DD-MM HH:MM:SS and some shorter
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