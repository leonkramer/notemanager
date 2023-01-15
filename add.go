package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func addHandler(args []string) {
	err := addNote(args)
	if (err != nil) {
		log.Fatal(err)
	}
}

func displayUsageAdd(fs flag.FlagSet) {
	fmt.Println("note add usage:")
	fs.PrintDefaults()
}

func addNote(args []string) (err error) {
	var tags []string

	var optHelp bool
	var optTemplate string
	fs := flag.NewFlagSet("note add", flag.ContinueOnError)
	fs.Usage = func() { helpNoteAdd() }
	fs.BoolVar(&optHelp, "h", false, "Display Help")
	fs.BoolVar(&optHelp, "help", false, "Display Help")
	fs.StringVar(&optTemplate, "t", "note", "Template of note")
	fs.StringVar(&optTemplate, "template", "note", "Template of note")
	if err = fs.Parse(args); err != nil {
		return
	}

	rargs := fs.Args()
	for k, arg := range rargs {
		// Arguments prefixed by + are tags
		if arg[0] == '+' {
			tags = append(tags, arg[1:])
			rargs = args[k+1:]
			continue
		}
		//rargs = args[k:]
		break
	}

	title := "Undefined"
	if len(rargs) > 0 {
		// Everything else is the title
		title = strings.Join(rargs, " ")
	}

	// -h flag given
	if optHelp {
		helpNoteAdd()
	}

	// user needs help
	if fs.NArg() == 1 && fs.Arg(0) == "help" {
		helpNoteAdd()
	}
	// Note title missing
	if fs.NArg() == 0 {
		helpNoteAdd()
	}


	in, err := os.ReadFile(filepath.Clean(notemanager.TemplateDir + "/" + optTemplate))
	if err != nil {
		in = []byte{}
	}
	timestamp := time.Now().UTC()
	id := uuid.New()
	file := filepath.Clean(fmt.Sprintf("%s/tmp/%s", notemanager.DataDir, id.String()))
	
	// Replace placeholders
	in = bytes.ReplaceAll(in, []byte("{{ nm.id }}"), []byte(id.String()))
	in = bytes.ReplaceAll(in, []byte("{{ nm.created.date }}"), []byte(timestamp.Format("2006-01-02")))
	in = bytes.ReplaceAll(in, []byte("{{ nm.created.time }}"), []byte(timestamp.Format("15:04")))
	in = bytes.ReplaceAll(in, []byte("{{ nm.created.offset }}"), []byte(timestamp.Format("-07:00")))
	in = bytes.ReplaceAll(in, []byte("{{ nm.title }}"), []byte(title))
	in = bytes.ReplaceAll(in, []byte("{{ nm.tags }}"), []byte(strings.Join(tags, ",")))

	// Create a file in temporary dir.
	// Once the note editor has been closed check if timestamp 
	// is newer than the file. If newer, move the file into
	// note directory and create data file.
	err = os.WriteFile(file, in, 0600)
	fileinfo, err := os.Stat(file)
	if err != nil {
		return
	}

	timestampInitial := fileinfo.ModTime()
	cmd := exec.Command(notemanager.Editor, file)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	
	err = cmd.Run()
	if err != nil {
		return
	}
	fileinfo, err = os.Stat(file)
	if err != nil {
		return
	}
	timestampAfter := fileinfo.ModTime()
	versions := []string{
		timestampAfter.UTC().Format(notemanager.VersionTimeFormat),
	}

	note := Note{
		Id: id,
		Title: title,
		Versions: versions,
		Tags: tags,
		DateCreated: timestampAfter.UTC(),
	}
	if timestampAfter != timestampInitial {
		moveFile(id.String(), timestampAfter.UTC().Format(notemanager.VersionTimeFormat))
		//metadata.Write()
		note.WriteData()
		fmt.Println("Note " + id.String() + " created.")
	}

	return
}
