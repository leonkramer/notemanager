package main

import "log"

func helpNote() {
	x := `USAGE
    ./note [FILTER] ACTION [OPTIONS] [PARAMETERS]

    For more specific usage use -h after the ACTION argument. E.g. ./note add -h


COMMANDS
    ./note add [TAG...] TITLE
        Create a note
    ./note [FILTER] [read]
        Read note with pagination
    ./note [FILTER] edit [VERSION]
        Create a new note version based on VERSION
    ./note [FILTER] list [notes|templates]
        List notes
    ./note [FILTER] tags [OPTIONS]
        List note tags
    ./note [FILTER] search [OPTIONS] [REGEXP]
        Search for regular expression matches
    ./note [FILTER] delete
        Mark note as deleted
    ./note [FILTER] print
        Print note content
    ./note [FILTER] versions
        Print the note versions
    ./note [FILTER] file { add | browse | list }
        Manage note file attachments
    ./note [FILTER] modify [TAGMODIFIER...] [TITLE]
        Modify note tags and title
    ./note version
        Display Notemanager version

`
	x += helpFilter()
	log.Fatal(Autobreak(x))
}

func helpNoteList() {
	x := `USAGE
    ./note [FILTER] list [notes|templates]


DESCRIPTION
    List a selection of notes matching the supplied FILTER terms. By default, if no FILTER is supplied, all notes are displayed except deleted notes.


ARGUMENTS
    PARAMETERS
        notes       List notes [Default]
        templates   List templates

`

	log.Fatal(Autobreak(x))
}

// ====================================================================
// Note Width = 72
func helpNoteSearch() {
	x := `USAGE
    ./note [FILTER] search [OPTIONS] [REGEXP]
    

DESCRIPTION
    Search for regular expression in notes matching FILTER.


ARGUMENTS
    FILTER
        For explanation of filters run: ./note -h
    OPTIONS
        -s|--case-sensitive
            Perform case sensitive pattern matching
    REGEXP
        Regular Expression to search for
`
	//x += helpFilter()

	log.Fatal(Autobreak(x))
}

func helpFilter() (x string) {
	x = `FILTER
    DESCRIPTION
        FILTER is a collection of options and terms or just a list of specific note ids to built a selection of notes. All supplied filter terms must match in order for a note to be included in the note selection. By default deleted notes are excluded from the note selection.


    SYNTAX
        [OPTIONS] [TERM...]


    ARGUMENTS
        OPTIONS
            -a|--all
                Select all notes, include deleted notes
            -h|--help   
                Display Notemanager Usage


        TERMS
            Filter notes based on supplied terms. All terms must match. Multiple terms must be separated by white space.
    
            created.after:TIMESTAMP
                Notes created after date
            created.before:TIMESTAMP
                Notes created before date
            modified.after:TIMESTAMP
                Notes modified after date
            modified.before:TIMESTAMP
                Notes modified before date
            +string
                Notes with tag string
            -string
                Notes without tag string
    
    
            TIMESTAMP:
                The TIMESTAMP syntax is YYYY-MM-DD [HH:mm[:ss]], you can optionally omit any separator. Any compontent of the time which is missing, is assumed to be 0.


        NOTE IDS
            One or multiple note ids can be supplied as the most exclusive filtering method. Note Ids must be separated by white spaces. If a note id is supplied the -a option is implicit. Note ids and other filters are mutually exclusive.
            ---
            [a-f0-9]{8}         Specific note with an abbreviated id
            UUID (36-bytes)     Specific note with full UUID
`

	return Autobreak(x)
	//return
}

func helpNoteTags() {
	x := `USAGE
    ./note [FILTER] tags [OPTIONS]


DESCRIPTION
    List all note tags of a selection of notes matching the FILTER terms.


ARGUMENTS
    FILTER
        For explanation of filters run: ./note -h
    OPTIONS
        -f|--full
            Display notes along with tags
        -h|--help
            Display usage
        -o|--order count|name
            Order tags by a parameter. [Default: count]

`

	log.Fatal(Autobreak(x))
}

func helpNoteAdd() {
	x := `USAGE
    ./note add [OPTIONS] [TAG...] [TITLE]


DESCRIPTION
    Create a note with supplied TAG and TITLE parameters. When issueing the command an editor will be started where the note content can be written to.


ARGUMENTS
    OPTIONS
        -t|--template
            Use template file as note layout. [Default=note]
    TAG
        Tags are strings prefixed by a '+' sign. Multiple tags can be supplied by separating them by spaces.
    TITLE
        All remaining arguments after TAG build the note's title. You usually do not need to wrap the title in quotes, unless you want to keep white spaces for some reason. [Default=Undefined]


EXAMPLE
    Create a note with tags 'important' and 'exam' with the title 'Exam Deadline'
        note add +important +exam Exam Deadline
`

	log.Fatal(Autobreak(x))
}

func helpNoteFile() {
	x := `USAGE
    ./note [FILTER] file [add FILE...|browse|delete FILE...|list|purge]


DESCRIPTION
    List or modify file attachments of notes.


ARGUMENTS
    FILTER
        For explanation of filters run: ./note -h
    PARAMETERS
        add FILE...
            Attach files to the note selection
        browse      Start a file manager to browse the files
        delete FILE     
            Mark the file attachments from the note selection as deleted
        list        List all files of note selection
        purge       Purge deleted file attachments of notes

`
	log.Fatal(Autobreak(x))
}

func helpNoteAlias() {
	x := `USAGE
    ./note [NOTE] alias [set STRING|delete|list]


DESCRIPTION
    List or modify alias of notes.


ARGUMENTS
    FILTER
        For explanation of filters run: ./note -h
    PARAMETERS
        set STRING
            Set note aliases
        delete      Delete note alias
        list        List all aliases of note

`
	log.Fatal(Autobreak(x))
}
