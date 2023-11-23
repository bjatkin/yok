package ekit

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/bjatkin/yok/v2/token"
)

const (
	maxWidth     = 128
	defaultWidth = 64
)

type Title string

const (
	TitleInvalidStatement    = Title("INVALID STATEMENT")
	TitleInvalidExpression   = Title("INVALID EXPRESSION")
	TitleInvalidControllFlow = Title("INVALID CONTROLL FLOW")
	TitleInvalidBlock        = Title("INVALID BLOCK")
	TitleUnknownToken        = Title("UNKNOWN TOKEN")
)

type Condition struct {
	Start      token.Token
	Conditions []string
}

// TODO: It would be nice if conditions could point to specific spots in the code as well.
// it would not always be needed but somtimes it would be nice
//
// Or maybe just allowing characters to be colored/ hilighted...
func NewCondition(start token.Token, conditions ...string) *Condition {
	return &Condition{
		Start:      start,
		Conditions: conditions,
	}
}

func (c *Condition) Error() string {
	// TODO: how should these be presented...
	// it's a bug if they show up but it would be nice if they
	// didn't look terrible
	return strings.Join(c.Conditions, "\n")
}

type Err struct {
	File string
	Src  string

	// TODO: support multiple start tokens so we can have a few lines showing when needed
	Start   token.Token
	Title   Title
	Message string
	Extra   Condition
}

func (e *Err) Error() string {
	var lines []string

	// get the header for the message
	terminalFd := int(os.Stdout.Fd())
	width, _, _ := term.GetSize(terminalFd)
	if width == 0 {
		width = defaultWidth
	}
	if width > maxWidth {
		width = maxWidth
	}

	var rightPadd string
	for i := len(e.Title) + 6; i < width; i++ {
		rightPadd += "-"
	}
	header := colorString(cyan, fmt.Sprintf("-- %s %s", e.Title, rightPadd))

	lines = append(lines, header)

	// get the cod snippet and related message
	var line, column int
	for i, r := range e.Src {
		if i == e.Start.Start {
			break
		}

		column++

		if r == '\n' {
			line++
			column = 0
		}
	}

	// this is safe since we just calculated the line number
	snippet := strings.Split(e.Src, "\n")[line]

	// get the path relative to the working dir
	path := e.File
	pwd, _ := os.Getwd()
	if strings.HasPrefix(path, pwd) {
		path = "." + strings.TrimPrefix(path, pwd)
	}

	lines = append(lines, colorString(darkGray, fmt.Sprintf("  %s:%d:%d", path, line+1, column+1)))

	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("% 3d| %s", line, snippet))
	padding := strings.Repeat(" ", column)
	errorMessage := colorString(red, "^ "+strings.ReplaceAll(e.Message, "\n", `\n`))
	lines = append(lines, fmt.Sprintf("   | %s%s", padding, errorMessage))
	lines = append(lines, "")

	// add all the relevent context
	for _, c := range e.Extra.Conditions {
		lines = append(lines, "  * "+strings.ReplaceAll(c, "\n", `\n`))
	}

	return strings.Join(lines, "\n") + "\n"
}

func (e *Err) AddCondition(conditions ...string) *Err {
	e.Extra.Conditions = append(e.Extra.Conditions, conditions...)
	return e
}

func (e *Err) At(start token.Token) *Err {
	e.Start = start
	return e
}

type ErrList struct {
	errs          []error
	maxErrorCount int
}

func NewErrList(maxErrorCount int) *ErrList {
	return &ErrList{
		maxErrorCount: maxErrorCount,
	}
}

func (e *ErrList) HasErrors() bool {
	return len(e.errs) > 0
}

func (e *ErrList) AddErr(err error) {
	if err == nil {
		return
	}

	e.errs = append(e.errs, err)
}

func (e *ErrList) Error() string {
	var ret string
	for i, err := range e.errs {
		if i >= e.maxErrorCount {
			stopMessage := fmt.Sprintf("too many errors: showing only %d of the %d total errors\n", i, len(e.errs))
			ret += colorString(green, stopMessage)
			return ret
		}
		ret += err.Error() + "\n"
	}

	return ret
}

type color string

const (
	black       color = "\033[0;30m"
	red         color = "\033[0;31m"
	green       color = "\033[0;32m"
	orange      color = "\033[0;33m"
	blue        color = "\033[0;34m"
	purple      color = "\033[0;35m"
	cyan        color = "\033[0;36m"
	lightGray   color = "\033[0;37m"
	darkGray    color = "\033[1;30m"
	lightRed    color = "\033[1;31m"
	lightGreen  color = "\033[1;32m"
	yellow      color = "\033[1;33m"
	lightBlue   color = "\033[1;34m"
	lightPurple color = "\033[1;35m"
	lightCyan   color = "\033[1;36m"
	white       color = "\033[1;37m"
)

func colorString(c color, s string) string {
	stopColor := "\033[0m"
	return fmt.Sprintf("%s%s%s", c, s, stopColor)
}
