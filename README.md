# Notemanager
A lightweight cross platform note manager for CLI.

Are you tired of losing track of your digital note taking? Notemanager is a lightweight software which integrates into your terminal and organizes your notes. 
The syntax is inspired by Taskwarrior, a popular task manager for the CLI. Notemanager lets you use your favorite editor for note taking. As it is written in Go, Notemanager can be compiled and used by hackers using Linux, MacOS or Windows.


# Installation
See the Release section.


# Usage

```
Generic usage:
---
note add [ -t TEMPLATEFILE ] [ +TAG .. ] TITLE
    Add note
note help
    Display usage
note list [ FILTER ]
    List notes which match filter
note version
    Display version


Note specific usage:
---
ID [ read ]
  Read note with pagination
ID print
  Print note
ID delete
  Mark note as deleted
```

# Templates
Create text templates in your template directory. Templates can be used for note creation and allow to use placeholders. Placeholders are wrapped in double curly braces `{{ }}`. The following placeholders will be replaced with note specific data:
* nmId
* nmTitle
* nmTags
* nmDate
* nmTime
* nmTimeOffset


## Examples
### Creating Notes
    note add +foo +bar My Foobar Title
Create note with tags _foo_ and bar and title _My Foobar Title_


    note add -t weekly Weekly Meeting
Create a note with title _Weekly Meeting_ and template file _weekly_


### Listing Notes
    note list +foo +bar
List notes, include only notes with tags _foo_ and _bar_


    note list +foo
List notes, include only notes with tag _foo_


    note list -foo
List notes, exclude notes with tag _foo_
