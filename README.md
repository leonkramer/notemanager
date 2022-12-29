# Notemanager
A lightweight cross platform note manager for CLI.

Are you tired of losing track of your digital note taking? Notemanager is a lightweight software which integrates into your terminal and organizes your notes. 
The syntax is inspired by Taskwarrior, a popular task manager for the CLI. Notemanager lets you use your favorite editor for note taking. As it is written in Go, Notemanager can be compiled and used by hackers using Linux, MacOS or Windows.


# Installation
See the Release section.


# Usage

```
Notemanager Usage:
	
Generic usage:
---
note add [ +TAG .. ] TITLE
	Add note
note help
	Display usage
note list [ OPTIONS ] [ FILTER ]
	List notes which match filter
	Options:
		-a		List all notes, include deleted
note version
	Display version


Note specific usage:
---
note ID [ read ]
	Read note with pagination
note ID print
	Print note
note ID delete
	Mark note as deleted
```

# Templates
Create text templates in your template directory. Templates can be used for note creation and allow to use placeholders. Placeholders are wrapped in double curly braces `{{ }}`. The following placeholders will be replaced with note specific data:
* `nmId`
* `nmTitle`
* `nmTags`
* `nmDate`
* `nmTime`
* `nmTimeOffset`

The following template is an example of a specific template layout. Assuming you name this file _weekly_, you can use the parameter `-t weekly` with the `note add` command to use the template.
```
---
Weekly Meeting
Participants:   
Date:           {{ nmDate }}
---

# Weekly Meeting @ {{ nmDate }}

## Topics
1.

## Todos
1. 

```


## Examples
### Creating Notes
    note add +foo +bar My Foobar Title
Create note with tags **foo, bar** and title **My Foobar Title**


    note add -t weekly Weekly Meeting
Create a note with title **Weekly Meeting** and template file **weekly**


### Listing Notes
    note list +foo +bar
List notes, include only notes with tags **foo** and **bar**


    note list +foo
List notes, include only notes with tag **foo**


    note list -- -foo
List notes, exclude notes with tag **foo**. -- Terminates the parsing of options. Without command tries to understand -foo as an option.
