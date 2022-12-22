package main

import (
	"fmt"
	"flag"
	"log"
	"os"
	"time"
	"bytes"
	"strings"
	"os/exec"
	"github.com/google/uuid"

)

func addHandler() {
	err := addNote()
	if (err != nil) {
		log.Fatal(err)
	}
}

func displayUsageAdd(fs flag.FlagSet) {
	fmt.Println("note add usage:")
	fs.PrintDefaults()
}

func addNote() (err error) {
	var args []string
	var tags []string

	fs := flag.NewFlagSet("add", flag.ContinueOnError)
	template := fs.String("t", "note", "Template of note")
	displayHelp := fs.Bool("h", false, "Display Help")
	if err = fs.Parse(os.Args[2:]); err != nil {
		return
	}

	args = fs.Args()
	for k, arg := range args {
		// Arguments prefixed by + are tags
		if arg[0] == '+' {
			tags = append(tags, arg[1:])
			continue
		}
		args = args[k:]
		break
	}
	// Everything else is the title
	title := strings.Join(args, " ")

	// -h flag given
	if *displayHelp {
		displayUsageAdd(*fs)
		return
	}
	// user needs help
	if fs.NArg() == 1 && fs.Arg(0) == "help" {
		displayUsageAdd(*fs)
		return
	}
	// Note title missing
	if fs.NArg() == 0 {
		displayUsageAdd(*fs)
		return
	}


	in, err := os.ReadFile(notemanager.TemplateDir + "/" + *template)
	if err != nil {
		in = []byte{}
	}
	timestamp := time.Now().UTC()
	id := uuid.New()
	file := fmt.Sprintf("%s/tmp/%s", notemanager.DataDir, id.String())
	
	// Replace placeholders
	in = bytes.ReplaceAll(in, []byte("{{ nmId }}"), []byte(id.String()))
	in = bytes.ReplaceAll(in, []byte("{{ nmDate }}"), []byte(timestamp.Format("2006-01-02")))
	in = bytes.ReplaceAll(in, []byte("{{ nmTime }}"), []byte(timestamp.Format("15:04")))
	in = bytes.ReplaceAll(in, []byte("{{ nmTimeOffset }}"), []byte(timestamp.Format("-07:00")))
	in = bytes.ReplaceAll(in, []byte("{{ nmTitle }}"), []byte(title))
	in = bytes.ReplaceAll(in, []byte("{{ nmTags }}"), []byte(strings.Join(tags, ",")))

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
