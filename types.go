package main

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/exp/slices"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)


type Note struct {
	Id 	uuid.UUID `yaml:"id"`
	Title string `yaml:"title"`
	Attachments []Attachment `yaml:"attachments,omitempty"`
	Versions []string `yaml:"versions"`
	Tags []string `yaml:"tags,omitempty"`
	DateCreated time.Time `yaml:"created"`
	DateModified []time.Time `yaml:"modified,omitempty"`
	DateDeleted time.Time `yaml:"deleted,omitempty"`
	latestContent []byte
	/*
	Tags []string
	Bytes int64 `yaml:"-"`
	File string
	Version	uint16
	Meta Metadata
	Date time.Time
	*/
}

type NoteFilter struct {
	TagsInclude []string
	TagsExclude []string
	CreatedAfter time.Time
	CreatedBefore time.Time
	ModifiedAfter time.Time
	ModifiedBefore time.Time
	IncludeDeleted bool
	IsDeleted bool
	HasFile bool
	Notes []string
}


type Metadata struct {
	Id 	uuid.UUID `yaml:"id"`
	Title string `yaml:"title"`
	Tags []string `yaml:"tags,omitempty"`
	Attachments []Attachment `yaml:"attachments,omitempty"`
	Versions []string `yaml:"-"`
	DateCreated time.Time `yaml:"created"`
	DateModified []time.Time `yaml:"modified,omitempty"`
	DateDeleted time.Time `yaml:"deleted,omitempty"`
}


type Config struct {
	DataDir string
	Editor string
	NoteDir string
	TempDir string
	TemplateDir string
	TerminalReader string
	FileManager string
	VersionTimeFormat string
	OutputTimeFormatShort string
	OutputTimeFormatLong string
	FilePermission os.FileMode
	FilePermissionReadonly os.FileMode
	DirPermission os.FileMode
}


type Attachment struct {
	Filename string `yaml:"filename"`
	Sha1 string `yaml:"sha1"` 
	DateCreated time.Time `yaml:"dateCreated"`
}


// checks if note exists
func (n Note) Exists() bool {
	if _, err := os.Stat(filepath.Clean(notemanager.NoteDir + "/" + n.Id.String())); errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		return true
	}
}


// write yaml encoded note struct to data file
func (n Note) WriteData() (err error) {
	err = os.WriteFile(filepath.Clean(notemanager.NoteDir + "/" + n.Id.String() + "/data"), n.Yaml(), 0600)
	if err != nil {
		log.Fatal(err)
	}

	return
}


// encode note struct to yaml
func (n Note) Yaml() (encodedYaml []byte) {
	encodedYaml, err := yaml.Marshal(n)
	if err != nil {
		log.Fatal(err)
	}

	return
}

func newNote(noteId string) (n Note, err error) {
	id, err := uuid.Parse(noteId)
	if (err != nil) {
		// check if abbreviated uuid
		if isUuidAbbr(noteId) == false {
			err = errors.New("Invalid note syntax")
			return
		}

		id, err = uuidByAbbr(noteId)
		if err != nil {
			err = errors.New("Note not found")
			return
		}
	}

	n.Id = id
	/* err = n.Metadata()
	if err != nil {
		fmt.Println("err: ", err)
	} */

	//n.Title = n.Meta.Title
	//fmt.Println(n)
	return
}

// 
func (n Note) Content(version ...string) (content []byte, err error) {
	var v string

	if len(version) > 1 {
		log.Fatal("Method Note.Content: Too many arguments in call")
	}
	if len(version) == 0 {
		v = n.LatestVersion()
	}
	if len(version) == 1 {
		v = version[0]
	}

	//fmt.Println(notemanager.NoteDir + "/" + n.Id.String() + "/" + v)
	content, err = os.ReadFile(filepath.Clean(notemanager.NoteDir + "/" + n.Id.String() + "/" + v))
	if err != nil {
		return
	}

	return
}


func loadNote(id string) (n Note, err error) {
	yml, err := os.ReadFile(filepath.Clean(notemanager.NoteDir + "/" + id + "/data"))
	if err != nil {
		return
	}

	err = yaml.Unmarshal(yml, &n)
	if err != nil {
		log.Fatal(err)
	}

	n.latestContent, err = n.Content()
	return
}


// creates text output of note.
func (n Note) Output(version string) (b []byte) {
	tpl :=  `+
+ Title:       %s
+ Date:        %s
+
+ Tags:        %s
+ Attachments: %d
+ Version:     %s
+


%s
` 

	
	content, err := n.Content(version)
	if err != nil {
		log.Fatal(err)
	}
	s := fmt.Sprintf(tpl,
		n.Title,
		n.DateCreated,
		strings.Join(n.Tags, ", "),
		len(n.Attachments),
		//n.Versions[len(n.Versions)-1],
		version,
		content,
	)

	b = []byte(s)

	return
}

// Mark a note as deleted by setting the DateDeleted value to current time stamp.
func (n Note) Delete() (err error) {
	if n.DateDeleted != (time.Time{}) {
		return
	}

	n.DateDeleted = time.Now().UTC()
	err = n.WriteData()
	return
}

// Delete the DateDeleted value and save to data file
func (n Note) Undelete() (err error) {
	n.DateDeleted = time.Time{}
	err = n.WriteData()
	return
}

// UUIDs are long and clumsy
func (n Note) ShortId() (s string) {
	return n.Id.String()[0:8]
}

// Checks if note matches the given filter.
func (n Note) MatchesFilter(filter NoteFilter) (ret bool, err error) {
	if filter.IncludeDeleted == false {
		if n.DateDeleted.IsZero() == false {
			ret = false
			return
		}
	}

	// Must have tags
	for _, x := range filter.TagsInclude {
		exists := false
		for _, t := range n.Tags {
			if t == x {
				exists = true
				continue
			}
		}
		if exists == false {
			ret = false
			return
		}
	}

	// Must not have tags
	for _, x := range filter.TagsExclude {
		exists := false
		for _, t := range n.Tags {
			if t == x {
				exists = true
				continue
			}
		}
		if exists == true {
			ret = false
			return
		}
	}

	if len(filter.Notes) > 0 {
		if slices.Contains(filter.Notes, n.Id.String()) == false {
			ret = false
			return
		}
	}

	ret = true
	return
}


// add tags to note. if one of the notes already exist
// the single tag is skipped, but the other tags are added
func (n *Note) AddTags(t []string) {
	for _, v := range t {
		if slices.Contains(n.Tags, v) {
			continue
		}
		n.Tags = append(n.Tags, v)
	}
}

// removes tags from note. if one of the notes does not exist
// the other notes are still removed
func (n *Note) RemoveTags(t []string) {
	RESTART:
	for k, v := range n.Tags {
		if slices.Contains(t, v) {
			n.Tags = slices.Delete(n.Tags, k, k+1)
			// slice has changed, could be out of bounds.
			// Therefore loop through it again.
			goto RESTART
		}
	}
}

func (n *Note) AddTag(t string) {
	n.AddTags([]string{t})
}

func (n *Note) RemoveTag(t string) {
	n.RemoveTags([]string{t})
}

// returns path of note: $NoteDir/$UUID
func (n Note) Path() (string) {
	return filepath.Clean(notemanager.NoteDir + `/` + n.Id.String())
}

// returns latest version of note
func (n Note) LatestVersion() (string) {
	return n.Versions[len(n.Versions)-1]
}

// moves temporary note from tempDir to specific note directory inside noteDir
func (n Note) moveTmpFile() (err error) {
	oldFile := filepath.Clean(notemanager.TempDir + `/` + n.Id.String())
	newFile := filepath.Clean(n.Path() + `/` + n.LatestVersion())

	os.MkdirAll(n.Path(), notemanager.DirPermission)
	err = os.Rename(oldFile, newFile)

	return
}