package bundler

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"unicode/utf8"
)

// Contant bundler variables
const (
	DefaultLeftDelim  = "//include("
	DefaultRightDelim = ")"
	eof               = -1
)

// Lexed Item Types
const (
	ItemError itemType = iota // error occurred; value is text of error
	ItemEOF
	ItemLeftDelim  // left action delimiter
	ItemRightDelim // right action delimiter
	ItemText       // plain text
	ItemFile       // file keyword
)

// itemType identifies the type of lex items.
type itemType int

// Pos represents a byte position in the original input text from which
// this template was parsed.
type Pos int

// Item represents a token or text string returned from the scanner.
type Item struct {
	Type itemType // The type of this item.
	Pos  Pos      // The starting position, in bytes, of this item in the input string.
	Val  string   // The value of this item.
}

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*Lexer) stateFn

// Lexer holds the state of the scanner.
type Lexer struct {
	id         string    // the lexer id; used only for error reports
	input      string    // the string being scanned
	leftDelim  string    // start of action
	rightDelim string    // end of action
	state      stateFn   // the next lexing function to enter
	pos        Pos       // current position in the input
	start      Pos       // start position of this item
	width      Pos       // width of last rune read from input
	lastPos    Pos       // position of most recent item returned by nextItem
	items      chan Item // channel of scanned items
	parenDepth int       // nesting depth of ( ) exprs
}

// NewLexer creates a new scanner for the input string and returns it for use
func NewLexer(id string, r io.Reader, leftDelim string, rightDelim string) (*Lexer, error) {

	input, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if leftDelim == "" {
		leftDelim = DefaultLeftDelim
	}

	if rightDelim == "" {
		rightDelim = DefaultRightDelim
	}

	l := &Lexer{
		id:         id,
		input:      string(input),
		leftDelim:  leftDelim,
		rightDelim: rightDelim,
		items:      make(chan Item),
	}

	go l.run()

	return l, nil
}

// next returns the next rune in the input.
func (l *Lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		// l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *Lexer) backup() {
	l.pos -= l.width
}

// emit passes an item back to the client.
func (l *Lexer) emit(t itemType) {
	l.items <- Item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *Lexer) reposition() {
	l.start = l.pos
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *Lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- Item{ItemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// NextItem returns the next item from the input.
func (l *Lexer) NextItem() Item {
	item := <-l.items
	l.lastPos = item.Pos
	return item
}

// run runs the state machine for the lexer.
func (l *Lexer) run() {
	for l.state = lexText; l.state != nil; {
		l.state = l.state(l)
	}
}

// lexText scans until an opening action delimiter, "{{".
func lexText(l *Lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], l.leftDelim) {
			if l.pos > l.start {
				l.emit(ItemText)
			}
			return lexLeftDelim
		}
		if l.next() == eof {
			break
		}
	}
	// Correctly reached EOF.
	if l.pos > l.start {
		l.emit(ItemText)
	}
	l.emit(ItemEOF)
	return nil
}

// lexLeftDelim scans the left delimiter, which is known to be present.
func lexLeftDelim(l *Lexer) stateFn {
	l.pos += Pos(len(l.leftDelim))
	l.emit(ItemLeftDelim)
	l.reposition()
	l.parenDepth = 0
	return lexFilename
}

// lexRightDelim scans the right delimiter, which is known to be present.
func lexRightDelim(l *Lexer) stateFn {
	l.pos += Pos(len(l.rightDelim))
	l.emit(ItemRightDelim)
	l.reposition()
	return lexText
}

// lexFilename scans the elements inside action delimiters.
func lexFilename(l *Lexer) stateFn {

	if strings.HasPrefix(l.input[l.pos:], l.rightDelim) {
		if l.parenDepth == 0 {
			l.emit(ItemFile)
			return lexRightDelim
		}
		return l.errorf("unclosed left paren")
	}

	switch r := l.next(); {
	case r == eof || isEndOfLine(r):
		return l.errorf("unclosed action")
	default:
		return lexFilename
	}
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}
