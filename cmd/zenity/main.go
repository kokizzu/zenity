package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ncruces/zenity"
	"github.com/ncruces/zenity/internal/zenutil"
)

//go:generate go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo

const (
	unspecified = "\x00"
)

var (
	// Application Options
	errorDlg          bool
	infoDlg           bool
	warningDlg        bool
	questionDlg       bool
	entryDlg          bool
	listDlg           bool
	passwordDlg       bool
	fileSelectionDlg  bool
	colorSelectionDlg bool
	notification      bool

	// General options
	title       string
	width       uint
	height      uint
	okLabel     string
	cancelLabel string
	extraButton string
	text        string
	icon        string
	multiple    bool

	// Message options
	noWrap        bool
	ellipsize     bool
	defaultCancel bool

	// Entry options
	entryText string
	hideText  bool

	// List options
	columns    int
	allowEmpty bool

	// File selection options
	save             bool
	directory        bool
	confirmOverwrite bool
	confirmCreate    bool
	showHidden       bool
	filename         string
	fileFilters      zenity.FileFilters

	// Color selection options
	defaultColor string
	showPalette  bool

	// Windows specific options
	cygpath bool
	wslpath bool
)

func init() {
	prevUsage := flag.Usage
	flag.Usage = func() {
		prevUsage()
		os.Exit(-1)
	}
}

func main() {
	setupFlags()
	flag.Parse()
	validateFlags()
	opts := loadFlags()
	zenutil.Command = true
	if zenutil.Timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(zenutil.Timeout)*time.Second)
		opts = append(opts, zenity.Context(ctx))
		_ = cancel
	}

	switch {
	case errorDlg:
		okResult(zenity.Error(text, opts...))
	case infoDlg:
		okResult(zenity.Info(text, opts...))
	case warningDlg:
		okResult(zenity.Warning(text, opts...))
	case questionDlg:
		okResult(zenity.Question(text, opts...))

	case entryDlg:
		strOKResult(zenity.Entry(text, opts...))

	case listDlg:
		if multiple {
			listResult(zenity.ListMultiple(text, flag.Args(), opts...))
		} else {
			strOKResult(zenity.List(text, flag.Args(), opts...))
		}

	case passwordDlg:
		_, pw, ok, err := zenity.Password(opts...)
		strOKResult(pw, ok, err)

	case fileSelectionDlg:
		switch {
		default:
			strResult(egestPath(zenity.SelectFile(opts...)))
		case save:
			strResult(egestPath(zenity.SelectFileSave(opts...)))
		case multiple:
			listResult(egestPaths(zenity.SelectFileMutiple(opts...)))
		}

	case colorSelectionDlg:
		colorResult(zenity.SelectColor(opts...))

	case notification:
		errResult(zenity.Notify(text, opts...))
	}

	flag.Usage()
}

func setupFlags() {
	// Application Options
	flag.BoolVar(&errorDlg, "error", false, "Display error dialog")
	flag.BoolVar(&infoDlg, "info", false, "Display info dialog")
	flag.BoolVar(&warningDlg, "warning", false, "Display warning dialog")
	flag.BoolVar(&questionDlg, "question", false, "Display question dialog")
	flag.BoolVar(&entryDlg, "entry", false, "Display text entry dialog")
	flag.BoolVar(&listDlg, "list", false, "Display list dialog")
	flag.BoolVar(&passwordDlg, "password", false, "Display password dialog")
	flag.BoolVar(&fileSelectionDlg, "file-selection", false, "Display file selection dialog")
	flag.BoolVar(&colorSelectionDlg, "color-selection", false, "Display color selection dialog")
	flag.BoolVar(&notification, "notification", false, "Display notification")

	// General options
	flag.StringVar(&title, "title", "", "Set the dialog `title`")
	flag.UintVar(&width, "width", 0, "Set the `width`")
	flag.UintVar(&height, "height", 0, "Set the `height`")
	flag.StringVar(&okLabel, "ok-label", "", "Set the label of the OK button")
	flag.StringVar(&cancelLabel, "cancel-label", "", "Set the label of the Cancel button")
	flag.StringVar(&extraButton, "extra-button", "", "Add an extra button")
	flag.StringVar(&text, "text", "", "Set the dialog `text`")
	flag.StringVar(&icon, "window-icon", "", "Set the window `icon` (error, info, question, warning)")
	flag.BoolVar(&multiple, "multiple", false, "Allow multiple items to be selected")

	// Message options
	flag.StringVar(&icon, "icon-name", "", "Set the dialog `icon` (dialog-error, dialog-information, dialog-question, dialog-warning)")
	flag.BoolVar(&noWrap, "no-wrap", false, "Do not enable text wrapping")
	flag.BoolVar(&ellipsize, "ellipsize", false, "Enable ellipsizing in the dialog text")
	flag.BoolVar(&defaultCancel, "default-cancel", false, "Give Cancel button focus by default")

	// Entry options
	flag.StringVar(&entryText, "entry-text", "", "Set the entry `text`")
	flag.BoolVar(&hideText, "hide-text", false, "Hide the entry text")

	// List options
	flag.Var(funcValue(addColumn), "column", "Set the column header")
	flag.Bool("hide-header", true, "Hide the column headers")
	flag.BoolVar(&allowEmpty, "allow-empty", true, "Allow empty selection (macOS only)")

	// File selection options
	flag.BoolVar(&save, "save", false, "Activate save mode")
	flag.BoolVar(&directory, "directory", false, "Activate directory-only selection")
	flag.BoolVar(&confirmOverwrite, "confirm-overwrite", false, "Confirm file selection if filename already exists")
	flag.BoolVar(&confirmCreate, "confirm-create", false, "Confirm file selection if filename does not yet exist (Windows only)")
	flag.BoolVar(&showHidden, "show-hidden", false, "Show hidden files (Windows and macOS only)")
	flag.StringVar(&filename, "filename", "", "Set the `filename`")
	flag.Var(funcValue(addFileFilter), "file-filter", "Set a filename filter (NAME | PATTERN1 PATTERN2 ...)")

	// Color selection options
	flag.StringVar(&defaultColor, "color", "", "Set the `color`")
	flag.BoolVar(&showPalette, "show-palette", false, "Show the palette")

	// Windows specific options
	if runtime.GOOS == "windows" {
		flag.BoolVar(&cygpath, "cygpath", false, "Use cygpath for path translation (Windows only)")
		flag.BoolVar(&wslpath, "wslpath", false, "Use wslpath for path translation (Windows only)")
	}

	// Command options
	flag.IntVar(&zenutil.Timeout, "timeout", 0, "Set dialog `timeout` in seconds")
	flag.StringVar(&zenutil.Separator, "separator", "|", "Set output `separator` character")

	// Detect unspecified values
	title = unspecified
	okLabel = unspecified
	cancelLabel = unspecified
	extraButton = unspecified
	text = unspecified
	icon = unspecified
}

func validateFlags() {
	var n int
	if errorDlg {
		n++
	}
	if infoDlg {
		n++
	}
	if warningDlg {
		n++
	}
	if questionDlg {
		n++
	}
	if entryDlg {
		n++
	}
	if listDlg {
		n++
	}
	if passwordDlg {
		n++
	}
	if fileSelectionDlg {
		n++
	}
	if colorSelectionDlg {
		n++
	}
	if notification {
		n++
	}
	if n != 1 {
		flag.Usage()
	}
}

func loadFlags() []zenity.Option {
	var opts []zenity.Option

	// Defaults

	setDefault := func(s *string, val string) {
		if *s == unspecified {
			*s = val
		}
	}
	switch {
	case errorDlg:
		setDefault(&title, "Error")
		setDefault(&icon, "dialog-error")
		setDefault(&text, "An error has occurred.")
		setDefault(&okLabel, "OK")
	case infoDlg:
		setDefault(&title, "Information")
		setDefault(&icon, "dialog-information")
		setDefault(&text, "All updates are complete.")
		setDefault(&okLabel, "OK")
	case warningDlg:
		setDefault(&title, "Warning")
		setDefault(&icon, "dialog-warning")
		setDefault(&text, "Are you sure you want to proceed?")
		setDefault(&okLabel, "OK")
	case questionDlg:
		setDefault(&title, "Question")
		setDefault(&icon, "dialog-question")
		setDefault(&text, "Are you sure you want to proceed?")
		setDefault(&okLabel, "Yes")
		setDefault(&cancelLabel, "No")
	case entryDlg:
		setDefault(&title, "Add a new entry")
		setDefault(&text, "Enter new text:")
		setDefault(&okLabel, "OK")
		setDefault(&cancelLabel, "Cancel")
	case listDlg:
		setDefault(&title, "Select items from the list")
		setDefault(&text, "Select items from the list below:")
		setDefault(&okLabel, "OK")
		setDefault(&cancelLabel, "Cancel")
	case passwordDlg:
		setDefault(&title, "Type your password")
		setDefault(&icon, "dialog-password")
		setDefault(&okLabel, "OK")
		setDefault(&cancelLabel, "Cancel")
	default:
		setDefault(&text, "")
	}

	// General options

	if title != unspecified {
		opts = append(opts, zenity.Title(title))
	}
	opts = append(opts, zenity.Width(width))
	opts = append(opts, zenity.Height(height))
	if okLabel != unspecified {
		opts = append(opts, zenity.OKLabel(okLabel))
	}
	if cancelLabel != unspecified {
		opts = append(opts, zenity.CancelLabel(cancelLabel))
	}
	if extraButton != unspecified {
		opts = append(opts, zenity.ExtraButton(extraButton))
	}

	var ico zenity.DialogIcon
	switch icon {
	case "error", "dialog-error":
		ico = zenity.ErrorIcon
	case "info", "dialog-information":
		ico = zenity.InfoIcon
	case "question", "dialog-question":
		ico = zenity.QuestionIcon
	case "important", "warning", "dialog-warning":
		ico = zenity.WarningIcon
	case "dialog-password":
		ico = zenity.PasswordIcon
	case "":
		ico = zenity.NoIcon
	}
	opts = append(opts, zenity.Icon(ico))

	// Message options

	if noWrap {
		opts = append(opts, zenity.NoWrap())
	}
	if ellipsize {
		opts = append(opts, zenity.Ellipsize())
	}
	if defaultCancel {
		opts = append(opts, zenity.DefaultCancel())
	}

	// Entry options

	opts = append(opts, zenity.EntryText(entryText))
	if hideText {
		opts = append(opts, zenity.HideText())
	}

	// List options

	if !allowEmpty {
		opts = append(opts, zenity.DisallowEmpty())
	}

	// File selection options

	if directory {
		opts = append(opts, zenity.Directory())
	}
	if confirmOverwrite {
		opts = append(opts, zenity.ConfirmOverwrite())
	}
	if confirmCreate {
		opts = append(opts, zenity.ConfirmCreate())
	}
	if showHidden {
		opts = append(opts, zenity.ShowHidden())
	}
	if filename != "" {
		opts = append(opts, zenity.Filename(ingestPath(filename)))
	}
	opts = append(opts, fileFilters)

	// Color selection options

	if defaultColor != "" {
		opts = append(opts, zenity.Color(zenutil.ParseColor(defaultColor)))
	}
	if showPalette {
		opts = append(opts, zenity.ShowPalette())
	}

	return opts
}

func errResult(err error) {
	if os.IsTimeout(err) {
		os.Exit(5)
	}
	if err == zenity.ErrExtraButton {
		os.Stdout.WriteString(extraButton)
		os.Stdout.WriteString(zenutil.LineBreak)
		os.Exit(1)
	}
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString(zenutil.LineBreak)
		os.Exit(-1)
	}
	os.Exit(0)
}

func okResult(ok bool, err error) {
	if err != nil {
		errResult(err)
	}
	if ok {
		os.Exit(0)
	}
	os.Exit(1)
}

func strResult(s string, err error) {
	if err != nil {
		errResult(err)
	}
	if s == "" {
		os.Exit(1)
	}
	os.Stdout.WriteString(s)
	os.Stdout.WriteString(zenutil.LineBreak)
	os.Exit(0)
}

func listResult(l []string, err error) {
	if err != nil {
		errResult(err)
	}
	if l == nil {
		os.Exit(1)
	}
	os.Stdout.WriteString(strings.Join(l, zenutil.Separator))
	os.Stdout.WriteString(zenutil.LineBreak)
	os.Exit(0)
}

func colorResult(c color.Color, err error) {
	if err != nil {
		errResult(err)
	}
	if c == nil {
		os.Exit(1)
	}
	os.Stdout.WriteString(zenutil.UnparseColor(c))
	os.Stdout.WriteString(zenutil.LineBreak)
	os.Exit(0)
}

func strOKResult(s string, ok bool, err error) {
	if err != nil {
		errResult(err)
	}
	if !ok {
		os.Exit(1)
	}
	os.Stdout.WriteString(s)
	os.Stdout.WriteString(zenutil.LineBreak)
	os.Exit(0)
}

func ingestPath(path string) string {
	if runtime.GOOS == "windows" && path != "" {
		var args []string
		switch {
		case wslpath:
			args = []string{"wsl", "wslpath", "-m"}
		case cygpath:
			args = []string{"cygpath", "-C", "UTF8", "-m"}
		}
		if args != nil {
			args = append(args, path)
			out, err := exec.Command(args[0], args[1:]...).Output()
			if err == nil {
				path = string(bytes.TrimSuffix(out, []byte{'\n'}))
			}
		}
	}
	return path
}

func egestPath(path string, err error) (string, error) {
	if runtime.GOOS == "windows" && path != "" && err == nil {
		var args []string
		switch {
		case wslpath:
			args = []string{"wsl", "wslpath", "-u"}
		case cygpath:
			args = []string{"cygpath", "-C", "UTF8", "-u"}
		}
		if args != nil {
			var out []byte
			args = append(args, filepath.ToSlash(path))
			out, err = exec.Command(args[0], args[1:]...).Output()
			if err == nil {
				path = string(bytes.TrimSuffix(out, []byte{'\n'}))
			}
		}
	}
	return path, err
}

func egestPaths(paths []string, err error) ([]string, error) {
	if runtime.GOOS == "windows" && err == nil && (wslpath || cygpath) {
		paths = append(paths[:0:0], paths...)
		for i, p := range paths {
			paths[i], err = egestPath(p, nil)
			if err != nil {
				break
			}
		}
	}
	return paths, err
}

type funcValue func(string) error

func (f funcValue) String() string     { return "" }
func (f funcValue) Set(s string) error { return f(s) }

func addColumn(s string) error {
	columns++
	if columns <= 1 {
		return nil
	}
	return errors.New("multiple columns not supported")
}

func addFileFilter(s string) error {
	var filter zenity.FileFilter

	if split := strings.SplitN(s, "|", 2); len(split) > 1 {
		filter.Name = strings.TrimSpace(split[0])
		s = split[1]
	}

	filter.Patterns = strings.Split(strings.TrimSpace(s), " ")
	fileFilters = append(fileFilters, filter)

	return nil
}
