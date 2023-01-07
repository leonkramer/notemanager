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

// note UUID ...
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

	var rargs []string
	if len(os.Args) > 3 {
		rargs = os.Args[3:]
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

		case "edit":
			err = noteEditHandler(note)
			if err != nil {
				log.Fatal(err)
			}

		case "read":
			err = noteReadHandler(note, rargs)
			if err != nil {
				log.Fatal(err)
			}

		case "print":
			err = notePrintHandler(note, os.Args[3:])
			if err != nil {
				log.Fatal(err)
			}

		case "versions":
			noteVersionsHandler(note, os.Args[3:])


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
		return
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

// CMD: note UUID edit
func noteEditHandler (n Note) (err error) {
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