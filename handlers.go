package main

import (
	"fmt"
	"errors"
	"os"
	"path/filepath"
	"log"
	"bytes"
	"os/exec"
	"time"
	"github.com/google/uuid"
	"strings"
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
		err = copyRegularFile(file, filepath.Clean(dstDir + "/" + basename))
		if err != nil {
			if err.Error() == "File already exists. Aborting." {
				fmt.Printf("File already attached. Use another name: %s.\n", basename)
			}
			fmt.Println(err)
			continue
		}
		note.Attachments = append(note.Attachments, Attachment{
			Filename: basename,
			Sha1: sha1,
			DateCreated: time.Now().UTC(),
		})
		note.WriteData()
		fmt.Printf("%s: Attached file %s.\n", note.ShortId(), file)
	}

	return
}

func noteFileHandler(note Note, args []string) (err error) {
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

		case "help":
			fmt.Println("<Insert help here.>")
	}

	return
}


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

		case "file":
			noteFileHandler(note, os.Args[3:])

		case "modify":
			noteModifyHandler(note, os.Args[3:])


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
			fmt.Println("Unknown command")
	}
}

func noteFileBrowseHandler(note Note) {
	path := filepath.Clean(fmt.Sprintf("%s/%s/attachments", notemanager.NoteDir, note.Id.String()))
	runFileManager(path)

}

func noteModifyHandler(n Note, args []string) (err error) {
	if len(args) == 0 {
		fmt.Println("Not enough parameters")
	}
	addTags, delTags, rargs, err := parseTagModifiers(args)
	if err != nil {
		log.Fatal(err)
	}

	n.AddTags(addTags)
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