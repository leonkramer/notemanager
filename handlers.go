package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Command Handler: note ID file add FILE [..]
// args is all arguments after: note ID file add
func noteFileAddHandler(note Note, args []string) (err error) {
	if len(args) == 0 {
		err = errors.New("Missing file")
		return
	}

MAIN:
	for _, file := range args[0:] {
		basename := filepath.Base(file)
		sha1, err := fileSha1(file)
		if err != nil {
			sha1 = "failed"
			fmt.Println(err)
		}
		for _, a := range note.Attachments {
			if sha1 == a.Sha1 {
				fmt.Printf("File with same checksum already attached: %s.\n", a.Filename)
				continue MAIN
			}
		}

		dstDir := filepath.Clean(notemanager.NoteDir + "/" + note.Id.String() + "/" + "attachments")
		err = os.MkdirAll(dstDir, os.FileMode(notemanager.DirPermission))
		if err != nil {
			log.Fatal(err)
		}
		err = copyRegularFile(file, filepath.Clean(dstDir+"/"+basename))
		if err != nil {
			if err.Error() == "File already exists. Aborting." {
				fmt.Printf("File already attached. Use another name: %s.\n", basename)
			}
			fmt.Println(err)
			continue
		}
		note.Attachments = append(note.Attachments, Attachment{
			Filename:    basename,
			Sha1:        sha1,
			DateCreated: time.Now().UTC(),
		})
		note.WriteData()
		fmt.Printf("%s: Attached file %s.\n", note.ShortId(), file)
	}

	return
}

func noteFileHandler(note Note, args []string) (err error) {
	var optHelp bool
	fs := flag.NewFlagSet("note file", flag.ContinueOnError)
	fs.Usage = func() { helpNoteFile() }
	fs.BoolVar(&optHelp, "h", false, "Display usage")
	fs.BoolVar(&optHelp, "help", false, "Display usage")
	if err = fs.Parse(args); err != nil {
		return
	}

	if optHelp {
		helpNoteFile()
	}

	var action string
	action = "list"
	if len(args) > 0 {
		action = args[0]
	}

	switch action {
	case "list":
		fmt.Println("list files")

	case "add":
		noteFileAddHandler(note, args[1:])

	case "browse":
		noteFileBrowseHandler(note)

	case "delete":
		fmt.Println("delete file")

	case "purge":
		fmt.Println("purge file")

	default:
		helpNoteList()
	}

	return
}

func noteFileBrowseHandler(note Note) {
	path := filepath.Clean(fmt.Sprintf("%s/%s/attachments", notemanager.NoteDir, note.Id.String()))
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		fmt.Printf("%s: Note does not have attachments\n", note.ShortId())
		return
	}
	runFileManager(path)
}

// Modify title or tag of note, or multiple tags of multiple notes
// Modification of title of multiple notes is not possible.
// CMD: note FILTER modify ARGS
func modifyHandler(filter NoteFilter, notes []Note, args []string) (err error) {
	_, _, rargs, err := parseTagModifiers(args)
	if err != nil {
		log.Fatal(err)
	}
	if len(notes) > 1 && len(rargs) > 0 {
		Exit("Cannot rename multiple notes")
	}
	for _, n := range notes {
		err = noteModifyHandler(n, args)
		if err != nil {
			Exit(err.Error())
		}
	}
	return
}

// Modify tag or title of single note.
func noteModifyHandler(n Note, args []string) (err error) {
	if len(args) == 0 {
		Exit("Not enough parameters")
		return
	}
	addTags, delTags, rargs, err := parseTagModifiers(args)
	if err != nil {
		log.Fatal(err)
	}

	err = n.AddTags(addTags)
	if err != nil {
		Exit(err.Error())
		return
	}
	n.RemoveTags(delTags)
	if len(rargs) > 0 {
		n.Title = strings.Join(rargs, " ")
	}

	err = n.WriteData()
	if err == nil {
		fmt.Println(n.ShortId() + ": Updated note.")
	}
	return
}

// CMD: note UUID edit
func noteEditHandler(n Note) (err error) {
	in, err := os.ReadFile(n.Path() + `/` + n.LatestVersion())
	if err != nil {
		log.Fatal(err)
	}

	// Create a temporary file in Notemanager tmp dir.
	// Once the note editor has been closed check if checksum
	// differs from the temporary file. If yes, move the file into
	// note directory and create data file.
	tmpFile := filepath.Clean(fmt.Sprintf("%s/tmp/%s", notemanager.DataDir, n.Id.String()))
	err = os.WriteFile(tmpFile, in, 0600)
	defer os.Remove(tmpFile)
	if err != nil {
		return
	}

	chksumBefore, err := fileSha1(tmpFile)
	if err != nil {
		log.Fatal(err)
	}

	// Run the Editor to edit tmpFile
	cmd := exec.Command(notemanager.Editor, tmpFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	err = cmd.Run()
	if err != nil {
		return
	}

	// Editor Done
	fileinfo, err := os.Stat(tmpFile)
	if err != nil {
		return
	}
	//timestampAfter := fileinfo.ModTime()
	version := fileinfo.ModTime().UTC().Format(notemanager.VersionTimeFormat)
	chksumAfter, err := fileSha1(tmpFile)
	if err != nil {
		log.Fatal(err)
	}
	n.Versions = append(n.Versions, version)

	if chksumBefore != chksumAfter {
		err = n.moveTmpFile()
		if err != nil {
			log.Fatal(err)
		}
		n.DateModified = append(n.DateModified, time.Now().UTC())
		n.WriteData()
		fmt.Println(n.ShortId() + ": Created note version " + version)
	}

	return
}

func notePrintHandler(n Note, args []string) (err error) {
	version := n.LatestVersion()
	if len(args) > 1 {
		err = errors.New("Too many arguments")
	}
	if len(args) > 0 {
		version = args[0]
	}
	fmt.Printf("%s", n.Output(version))
	return
}

func noteReadHandler(n Note, args []string) (err error) {
	version := n.LatestVersion()
	if len(args) > 1 {
		err = errors.New("Too many arguments")
	}
	if len(args) > 0 {
		version = args[0]
	}

	cmd := exec.Command(notemanager.TerminalReader)
	cmd.Stdin = bytes.NewReader(n.Output(version))
	cmd.Stdout = os.Stdout

	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	return
}

// display note versions
func noteVersionsHandler(n Note, args []string) (err error) {
	for i, ver := range n.Versions {
		// "% <INT>d: %s\n"
		//lineTpl := fmt.Sprintf("%%%dd: %%s\n", len(string(len(n.Versions)))+1)
		lineTpl := "%d: %s\n"
		fmt.Printf(lineTpl, i, ver)
	}
	fmt.Printf("---\nTotal: %d\n", len(n.Versions))

	return
}

// display collection of tags
func tagsHandler(filter NoteFilter, args []string) (err error) {
	var optHelp bool
	var optOrder string
	var optFull bool
	fs := flag.NewFlagSet("note tags", flag.ContinueOnError)
	fs.Usage = func() { helpNoteTags() }
	fs.BoolVar(&optHelp, "h", false, "Display usage")
	fs.BoolVar(&optHelp, "help", false, "Display usage")
	fs.BoolVar(&optFull, "f", false, "Display notes along with tags")
	fs.BoolVar(&optFull, "full", false, "Display notes along with tags")
	fs.StringVar(&optOrder, "o", "count", "Ordering of tags. OPTIONS=count|name")
	fs.StringVar(&optOrder, "order", "count", "Ordering of tags. OPTIONS=count|name")
	if err = fs.Parse(args); err != nil {
		return
	}

	if optHelp {
		helpNoteTags()
	}

	args = fs.Args()
	/* filter, rargs, err := parseFilter(args)
	fmt.Sprintln(rargs) */
	notes, err := notes(filter)
	if err != nil {
		log.Fatal(err)
	}

	type TaggedNote struct {
		Id          string
		Title       string
		DateCreated time.Time
	}
	tags := make(map[string][]TaggedNote)
	for _, n := range notes {
		for _, tag := range n.Tags {
			tags[tag] = append(tags[tag], TaggedNote{n.ShortId(), n.Title, n.DateCreated})
		}
	}

	// generate slice of tags
	keys := make([]string, 0, len(tags))
	for key := range tags {
		keys = append(keys, key)
	}

	// sort alphabetically by tag first
	// or results of multiple calls vary slightly.
	sort.Strings(keys)

	switch optOrder {
	case "count":
		sort.SliceStable(keys, func(i, j int) bool {
			return len(tags[keys[i]]) > len(tags[keys[j]])
		})
	}

	// Default Output
	for _, k := range keys {
		fmt.Printf("%s (%d)\n", k, len(tags[k]))
		if optFull {
			for _, t := range tags[k] {
				fmt.Printf("  - %s: %s (%s)\n", t.Id, t.Title, t.DateCreated.Format(notemanager.OutputTimeFormatShort))
			}
			fmt.Println()
		}
	}

	return
}

func searchHandler(filter NoteFilter, args []string) (err error) {
	var optHelp bool
	var optCaseSensitive bool
	fs := flag.NewFlagSet("notemanager search", flag.ContinueOnError)
	fs.Usage = func() { helpNoteSearch() }
	fs.BoolVar(&optHelp, "h", false, "Display usage")
	fs.BoolVar(&optHelp, "help", false, "Display usage")
	fs.BoolVar(&optCaseSensitive, "s", false, "Perform case sensitive search")
	fs.BoolVar(&optCaseSensitive, "case-sensitive", false, "Perform case sensitive search")
	if err = fs.Parse(args); err != nil {
		return
	}

	if optHelp || len(args) == 0 {
		helpNoteSearch()
	}

	rargs := fs.Args()
	notes, err := notes(filter)

	if len(rargs) > 1 {
		Exit("Too many arguments")
	}
	needle := rargs[0]

	type FileMatch struct {
		Id      string
		Lines   []string
		Excerpt string
	}
	var matches []FileMatch
	for _, n := range notes {
		// search case insensitive by default
		//matchString := `(?im)(^.*%s.*$)`
		matchString := `(?im)(.*%s.*)`
		if optCaseSensitive {
			matchString = `(?m)(.*%s.*)`
		}

		r := regexp.MustCompile(fmt.Sprintf(matchString, needle))
		if r.Match(n.latestContent) {
			var lines []string
			sc := bufio.NewScanner(strings.NewReader(string(n.latestContent)))
			for sc.Scan() {
				lines = append(lines, sc.Text())
			}

			var matchLines []string
			for i, line := range lines {
				if r.MatchString(line) {
					matchLines = append(matchLines, strconv.Itoa(i+1))
				}
			}

			if len(matchLines) > 0 {
				matches = append(matches, FileMatch{
					Id:      n.ShortId(),
					Lines:   matchLines,
					Excerpt: "",
				})
			}

			/* 		matches := r.FindSubmatch(n.latestContent)
			for _, v := range matches {
				//	fmt.Printf("Match: %s\n", v)
			} */
		}
	}
	fmt.Println("Search Results for pattern: " + needle)
	fmt.Println("Matches found:", len(matches))

	if len(matches) > 0 {
		fmt.Println()
		fmt.Println(`Matches:`)

		for _, m := range matches {
			//fmt.Printf(" - %s\n", m)
			fmt.Printf(" - %s at lines %s\n", m.Id, strings.Join(m.Lines, ", "))
		}
	}

	return
}

func undeleteHandler(notes []Note, args []string) (err error) {
	for _, n := range notes {
		err = n.Undelete()
		if err != nil {
			return
		}

		fmt.Printf("%s: Undeleted\n", n.ShortId())
	}

	return
}

func deleteHandler(notes []Note, args []string) (err error) {
	for _, n := range notes {
		err = n.Delete()
		if err != nil {
			return
		}

		fmt.Printf("%s: Deleted\n", n.ShortId())
	}

	return
}

func printHandler(notes []Note, args []string) (err error) {
	for _, n := range notes {
		notePrintHandler(n, args)
	}

	return
}

func readHandler(notes []Note, args []string) (err error) {
	for _, n := range notes {
		//version := n.LatestVersion()
		noteReadHandler(n, args)
	}

	return
}

func versionsHandler(notes []Note, args []string) (err error) {
	err = noteVersionsHandler(notes[0], args)
	if err != nil {
		Exit(err.Error())
	}
	return
}

func fileHandler(notes []Note, args []string) (err error) {
	err = noteFileHandler(notes[0], args)
	if err != nil {
		Exit(err.Error())
	}
	return
}
