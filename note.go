package main

import (
	"errors"
	"fmt"
	"os"
	"log"
	"regexp"
	"os/exec"
	"bytes"
	_"time"
	"path/filepath"
	"github.com/google/uuid"
)

func noteHandler() {
	// is first arg an uuid?
	id, err := uuid.Parse(os.Args[1])
	if (err != nil) {
		// check if abbreviated uuid
		if isUuidAbbr(os.Args[1]) == false {
			fmt.Println("Invalid note syntax")
			os.Exit(2)
		}

		id, err = uuidByAbbr(os.Args[1])
		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}
	}

	note, err := loadNote(id.String())

	if err != nil {
		fmt.Println("err: ", err)
	}

	action := "read"
	if len(os.Args) > 2 {
		action = os.Args[2]
	}

	switch(action) {
		case "delete":
			err := note.Delete()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("OK")

		case "undelete":
			err := note.Undelete()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("OK")


		case "read":
			cmd := exec.Command(notemanager.TerminalReader)
			cmd.Stdin = bytes.NewReader(note.Output())
			cmd.Stdout = os.Stdout

			err := cmd.Run()
    		if err != nil {
        		log.Fatal(err)
    		}

		case "print":
			fmt.Printf("%s", note.Output())


		default:
			fmt.Println("unknown")
	}
}

// Generate a uuid from a abbreviated uuid string.
// The abbreviated string must be at the start of the uuid and be
// 8 bytes long.
// In order to get the real uuid, it must exist in note directory.
// Example: 1cf77aeb => 1cf77aeb-fcb2-44ad-87d6-69717dba1d0c
func uuidByAbbr(x string) (r uuid.UUID, err error) {
	if isUuidAbbr(x) == false {
		err = errors.New("No such note id")
		return
	}
	
	matches, err := filepath.Glob(notemanager.NoteDir + "/" + x + "*")
	if (err != nil) {
		return
	}

	switch len(matches) {
	case 0:
		err = errors.New("No such note id")

	case 1:
		r, err = uuid.Parse(filepath.Base(matches[0]))

	default:
		err = errors.New("At least 2 notes found starting with " + x + ", use full uuid")
	}

	return
}

// Checks if input string is a valid Uuid abbreviation.
// Should be syntax: [a-f0-9]{8}
func isUuidAbbr(x string) (bool) {
	reg := regexp.MustCompile(`^[a-f0-9]{8}$`)
	if reg.MatchString(x) {
		return true
	}
	return false
}