package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
)

// parse Command for FILTER arguments.
// i.e. +tag, -tag, created.after, created.before,
// modified.after, modified.before etc.
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

		noteId, aliasExists := aliases.Get(v)
		if aliasExists {
			filter.Notes = append(filter.Notes, noteId.String())
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

		// try Note ID
		if len(v) == 36 {
			if _, err := uuid.Parse(v); err != nil {
				Exit("Invalid UUID syntax: " + v)
			}

			n, err := loadNote(v)
			if err != nil {
				Exit(err.Error())
			}

			filter.Notes = append(filter.Notes, n.Id.String())
			continue
		}

		// try short Note ID
		if isUuidAbbr(v) {
			id, err := uuidByAbbr(v)
			if err != nil {
				Exit(`No such note: ` + v)
			}

			n, err := loadNote(id.String())
			if err != nil {
				Exit(err.Error())
			}
			filter.Notes = append(filter.Notes, n.Id.String())
			continue
		}

		rargs = args[k:]
		break
	}

	// check if note ids and other filters are supplied.
	// that makes not a lot of sense and should fail
	// so client does not get unexpected results.
	if len(filter.Notes) > 0 {
		// assume to include deleted notes, if specific
		// note ids are supplied
		filter.IncludeDeleted = true

		cmp := NoteFilter{
			Notes:          filter.Notes,
			IncludeDeleted: filter.IncludeDeleted,
		}
		if reflect.DeepEqual(filter, cmp) == false {
			Exit("Note IDs and filter terms supplied, but they are mutually exclusive")
		}
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
		//r := regexp.MustCompile(`^\+[\pL0-9]+$`)
		r := regexp.MustCompile(`^\+`)
		if r.MatchString(v) {
			if slices.Contains(posTags, v[1:]) {
				continue
			}
			posTags = append(posTags, v[1:])
			continue
		}

		// match -somestring
		//r = regexp.MustCompile(`^\-[\pL0-9]+$`)
		r = regexp.MustCompile(`^\-`)
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

// returns a list of slice of Notes matching the filter
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
		if err != nil {
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
			for kk, vv := range v {
				str += fmt.Sprintf("%-*s", maxLength[kk]+2, vv)
			}
			str += "\n"
		}

		fmt.Printf("%s", str)
	}

	return
}

func Exit(msg string) {
	fmt.Fprintf(os.Stderr, "%s\n", msg)
	runtime.Goexit()
}

// Convert long texts into a better human readable
// format. I.e. long lines will be auto broken into max 72 chars
// or the terminal width, whatever is smaller.
func Autobreak(x string) (ret string) {
	var lines []string
	var width int
	width = terminalWidth()

	if width > 72 {
		width = 72
	}

	//x = strings.Replace(x, "\t", "    ", -1)

	// keep the already defined new lines, therefore create
	// slice of lines.
	for _, l := range strings.Split(x, "\n") {
		// work = our line to work with
		work := l

		// if the line is prefixed by a space or tab, read this
		// prefix into a variable, so we can later prefix
		// every broken line with it.
		var prefix string
		for _, c := range work {
			if c == ' ' || c == '\t' {
				prefix += string(c)
				continue
			}
			break
		}

		work = work[len(prefix):]

		splitPos := width - len(prefix)

		for len(work) > splitPos {
			index := strings.LastIndex(work[0:splitPos], ` `)
			if index == -1 {
				lines = append(lines, prefix+work)
				work = ``
			} else {
				lines = append(lines, prefix+work[:index])
				work = work[index+1:]
			}

		}
		lines = append(lines, prefix+work)
		continue
	}

	for _, line := range lines {
		ret += line + "\n"
	}
	return
}

// moves temporary note from tempDir to specific note directory inside noteDir
func moveFile(id string, version string) (err error) {
	oldFile := filepath.Clean(fmt.Sprintf("%s/%s", notemanager.TempDir, id))
	newFile := filepath.Clean(fmt.Sprintf("%s/%s/%s", notemanager.NoteDir, id, version))

	os.Mkdir(filepath.Clean(fmt.Sprintf("%s/%s", notemanager.NoteDir, id)), notemanager.DirPermission)
	err = os.Rename(oldFile, newFile)

	return
}

func readMetadataFile(id string) (metadata Metadata, err error) {
	metadataRaw, err := os.ReadFile(filepath.Clean(notemanager.NoteDir + "/" + id + "/meta"))
	if err != nil {
		return
	}

	err = yaml.Unmarshal(metadataRaw, &metadata)
	return metadata, err
}

func noteId(file string) []byte {
	body, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	rUuid := regexp.MustCompile(`(?m)^\-\-\-$[\d\D]+^Id: (?P<uuid>[a-z0-9\-]+)$[\d\D]+^\-\-\-$`)
	rVersion := regexp.MustCompile(`(?m)^\-\-\-$[\d\D]+^Version: (?P<version>[0-9]+)$[\d\D]+^\-\-\-$`)
	if rUuid.Match(body) && rVersion.Match(body) {
		matches := rUuid.FindSubmatch(body)
		uuid := matches[rUuid.SubexpIndex("uuid")]
		matches = rVersion.FindSubmatch(body)
		version := matches[rVersion.SubexpIndex("version")]
		versionNew, err := strconv.Atoi(string(version))
		versionNew += 1
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Uuid: %s", uuid)
		fmt.Printf("Version: %s", version)
		fmt.Println("VersionNew: ", versionNew)
		datadir, _ := cfg.String("default", "datadir")
		dstDir := fmt.Sprintf("%s/%s", datadir, uuid)
		os.MkdirAll(dstDir, 0655)
		if _, err := os.Stat(dstDir); os.IsNotExist(err) {
			fmt.Println("directory does not exist")
		}
	}

	return body
}

// generate sha1 hash from file
func fileSha1(path string) (ret string, err error) {
	fh, err := os.Open(path)
	if err != nil {
		return
	}
	defer fh.Close()

	hash := sha1.New()
	_, err = io.Copy(hash, fh)
	if err != nil {
		return
	}
	hashInBytes := hash.Sum(nil)[:20]
	ret = hex.EncodeToString(hashInBytes)
	return
}

func FilterIsDefined(filter NoteFilter) bool {
	cmp := NoteFilter{
		Notes: filter.Notes,
	}

	return !reflect.DeepEqual(filter, cmp)
}
