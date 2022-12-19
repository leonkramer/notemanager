package main

import (
	"time"
	"fmt"
	"log"
	"errors"
	"os"
	"strings"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
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
	Before time.Time
	After time.Time
	Deleted bool
	Attachments bool
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
	VersionTimeFormat string
	OutputTimeFormatShort string
	OutputTimeFormatLong string
}


type Attachment struct {
	Filename string `yaml:"filename"`
	Sha1 string `yaml:"sha1"` 
	Bytes uint `yaml:"bytes"`
	DateCreated time.Time `yaml:"dateCreated"`
}

func (m Metadata) Write() (err error) {
	err = os.WriteFile(notemanager.NoteDir + "/" + m.Id.String() + "/meta", m.Yaml(), 0600)
	if err != nil {
		log.Fatal(err)
	}

	return
}

func (m Metadata) Yaml() (encodedYaml []byte) {
	encodedYaml, err := yaml.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}

	return
}

func (n Note) Exists() bool {
	if _, err := os.Stat(notemanager.NoteDir + "/" + n.Id.String()); errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		return true
	}
}


func (n Note) WriteData() (err error) {
	err = os.WriteFile(notemanager.NoteDir + "/" + n.Id.String() + "/data", n.Yaml(), 0600)
	if err != nil {
		log.Fatal(err)
	}

	return
}

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

func (n Note) ReadContent(version ...string) (content []byte, err error) {
	var v string

	if len(version) > 1 {
		log.Fatal("Method Note.Content: Too many arguments in call")
	}
	if len(version) == 0 {
		v = n.Versions[0]
	}
	if len(version) == 1 {
		v = version[0]
	}

	//fmt.Println(notemanager.NoteDir + "/" + n.Id.String() + "/" + v)
	content, err = os.ReadFile(notemanager.NoteDir + "/" + n.Id.String() + "/" + v)
	if err != nil {
		return
	}

	return
}

/* func (n *Note) Data() (err error) {
	n.Meta, err = readMetadataFile(n.Id.String())
	return
} */


func loadNote(id string) (n Note, err error) {
	yml, err := os.ReadFile(notemanager.NoteDir + "/" + id + "/data")
	if err != nil {
		return
	}

	err = yaml.Unmarshal(yml, &n)
	if err != nil {
		log.Fatal(err)
	}

	n.latestContent, err = n.ReadContent()
	return
}

func (n Note) Output() (b []byte) {
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

	s := fmt.Sprintf(tpl,
		n.Title,
		n.DateCreated,
		strings.Join(n.Tags, ", "),
		len(n.Attachments),
		n.Versions[len(n.Versions)-1],
		n.latestContent,
	)

	b = []byte(s)

	return
}

func (n Note) Delete() (err error) {
	if n.DateDeleted != (time.Time{}) {
		return
	}

	n.DateDeleted = time.Now().UTC()
	err = n.WriteData()
	return
}

func (n Note) ShortId() (s string) {
	return n.Id.String()[0:8]
}

// Checks if note matches the given filter
func (n Note) MatchesFilter(filter NoteFilter) (ret bool, err error) {
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

	ret = true
	return
}