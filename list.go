package main

import (
	"os"
	"fmt"
	"log"
	"regexp"
	"errors"
	"strings"
	"github.com/google/uuid"
)

func listHandler() {
	if (len(os.Args) > 2) {
		switch os.Args[2] {
			case "templates":
				listTemplates(notemanager.TemplateDir)
  
		  	case "notes":
				listNotes()
  
		  	default:
				fmt.Println("default")
		}
	} else {
		listNotes()
	}
}

func listTemplates(path string) {
	fmt.Println("Note Templates:")
	files, err := os.ReadDir(path)
	if err != nil {
	  log.Fatal(err)
	}
	for _, file := range files {
	  info, err := file.Info()
	  if err != nil {
		log.Fatal(err)
	  }
	  fmt.Printf("   %s (%d Bytes, modified: %s)\n", file.Name(), info.Size(), info.ModTime())
	}
	//fmt.Println(files)
	return
}


func listNotes() {
	files, err := os.ReadDir(notemanager.NoteDir)
	if err != nil {
        log.Fatal(err)
    }

	var notes []Note
	for _, file := range files {
		
		if file.IsDir() == false {
			continue
		}

		noteId := uuid.MustParse(file.Name())

		note, err := loadNote(noteId.String())
		if (err != nil) {
			log.Fatal(err)
		}

		notes = append(notes, note)
		
        if string(file.Name()[0]) == "." {
          continue
        }
    }

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

func getNote(n Note, version string) {
	files, err := os.ReadDir(notemanager.NoteDir + "/" + n.Id.String())
	if err != nil {
		log.Fatal(err)
	}

	var versions []string

	for _, file := range files {
		if file.Name() == "data" {
			continue
		}
		versions = append(versions, file.Name())
		fmt.Println(file)
	}

	if version == "latest" {
		version := versions[len(versions)-1]
		fmt.Println(version)
	} else {
		fmt.Println("not latest")
	}
}


// Return array of all note versions.
// Each note version is represented by the file name with syntax YYYYMMDD-HHMMSS.
func noteVersions(n Note) (versions []string) {
	files, err := os.ReadDir(notemanager.NoteDir + "/" + n.Id.String())
	if err != nil {
		log.Fatal(err)
	}

	re := regexp.MustCompile(`^[0-9]{8}\-[0-9]{6}`)
	
	for _, file := range files {
		if re.MatchString(file.Name()) {
			versions = append(versions, file.Name())
		}
	}

	return
}

// display latest note version.
// Each note version is represented by the file name with syntax YYYYMMDD-HHMMSS.
// The sort is automatically by name, hence the last file is the latest.
func noteLatestVersion(n Note) (version string, err error) {
	versions := noteVersions(n)
	if len(versions) == 0 {
		return version, errors.New("No note version found")
	}
	version = versions[len(versions)-1]
	return
}