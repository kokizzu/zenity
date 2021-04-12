# Zenity dialogs for Golang, Windows and macOS

[![PkgGoDev](https://pkg.go.dev/badge/image)](https://pkg.go.dev/github.com/ncruces/zenity)
[![Go Report](https://goreportcard.com/badge/github.com/ncruces/zenity)](https://goreportcard.com/report/github.com/ncruces/zenity)

This repo includes both a cross-platform Go package providing
[Zenity](https://help.gnome.org/users/zenity/stable/)-like dialogs
(simple dialogs that interact graphically with the user),
as well as a *“port”* of the `zenity` command to both Windows and macOS based on that library.

Implemented dialogs:
* [message](https://github.com/ncruces/zenity/wiki/Message-dialog) (error, info, question, warning)
* [text entry](https://github.com/ncruces/zenity/wiki/Text-Entry-dialog)
* [list](https://github.com/ncruces/zenity/wiki/List-dialog) (limited)
* [password](https://github.com/ncruces/zenity/wiki/Password-dialog)
* [file selection](https://github.com/ncruces/zenity/wiki/File-Selection-dialog)
* [color selection](https://github.com/ncruces/zenity/wiki/Color-Selection-dialog)
* [notification](https://github.com/ncruces/zenity/wiki/Notification)

Behavior on Windows, macOS and other Unixes might differ slightly.
Some of that is intended (reflecting platform differences),
other bits are unfortunate limitations,
others still are open to be fixed.

## Why?

There are a bunch of other dialog packages for Go.\
Why reinvent this particular wheel?

#### Benefits:

* no `cgo` (see [benefits](https://dave.cheney.net/2016/01/18/cgo-is-not-go), mostly cross-compilation)
* no main loop (or any other threading or initialization requirements)
* cancelation through [`context`](https://golang.org/pkg/context/)
* on Windows:
  * no additional dependencies
    * Explorer shell not required
    * works in Server Core
  * Unicode support
  * High DPI support (no manifest required)
  * WSL/Cygwin/MSYS2 [support](https://github.com/ncruces/zenity/wiki/Zenity-for-WSL,-Cygwin,-MSYS2)
* on macOS:
  * only dependency is `osascript`
* on other Unixes:
  * wraps either one of `zenity`, `qarma`, `matedialog`
