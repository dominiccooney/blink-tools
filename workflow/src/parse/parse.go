package parse

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Scanner interface {
	io.Seeker
	io.RuneScanner
	Here() (row int, col int)
	Peek() rune
	ReadWhile(func (r rune) bool) (string, error)
	SkipTo(s string) error
	SkipPast(s string) error
	SkipWhile(func (r rune) bool)
	SkipSpace()
	Int() (int, error)
	Lit(s string) error
}

type StringScanner struct {
	s string
	i int  // offset
	row int
	startOfRow int
}

func NewUtf8Scanner(s string) Scanner {
	return &StringScanner{s: s, i: 0, row: 0, startOfRow: 0}
}

func (s *StringScanner) Peek() rune {
	r, _, _ := s.ReadRune()
	s.UnreadRune()
	return r
}

func (s *StringScanner) Seek(offset int64, whence int) (ret int64, err error) {
	var newOffset int64

	switch whence {
	case 0:
		newOffset = offset
	case 1:
		newOffset = int64(s.i) + offset
	case 2:
		newOffset = int64(len(s.s)) + offset
	}

	switch {
	case newOffset < 0:
		return 0, errors.New("seek past beginning of underlying string")
	case newOffset > int64(len(s.s)):
		return 0, errors.New("seek past end of underlying string")
	}

	switch {
	case newOffset < int64(s.i):
		for int64(s.i) > newOffset {
			s.UnreadRune()
		}
	case int64(s.i) < newOffset:
		for int64(s.i) < newOffset {
			s.ReadRune()
		}
	}

	return int64(s.i), nil
}

func (s *StringScanner) UnreadRune() error {
	if s.i == 0 {
		return errors.New("UnreadRune at beginning of underlying string")
	}
	r, size := utf8.DecodeLastRuneInString(s.s[:s.i])
	if r == '\n' {
		s.row--
		j := s.i - size
		for j > 0 {
			t, tSize := utf8.DecodeLastRuneInString(s.s[:j])
			if t == '\n' {
				break
			}
			j -= tSize
		}
		s.startOfRow = j
	}
	s.i -= size
	return nil
}

func (s *StringScanner) ReadRune() (r rune, size int, err error) {
	if s.i == len(s.s) {
		return rune(0), 0, io.EOF
	}
	r, size = utf8.DecodeRuneInString(s.s[s.i:])
	s.i += size
	if r == '\n' {
		s.row++
		s.startOfRow = s.i
	}
	err = nil
	return
}

func (s *StringScanner) Here() (row int, col int) {
	return s.row, utf8.RuneCountInString(s.s[s.startOfRow:s.i])
}

func (s *StringScanner) ReadWhile(f func(r rune) bool) (string, error) {
	start := s.i
	for s.i < len(s.s) {
		r, _, err := s.ReadRune()
		if err != nil {
			return "", err
		}
		if !f(r) {
			err := s.UnreadRune()
			return s.s[start:s.i], err
		}
	}
	return s.s[start:s.i], io.EOF
}

func (s *StringScanner) SkipTo(search string) error {
	index := strings.Index(s.s[s.i:], search)
	if index == -1 {
		return fmt.Errorf("\"%s\" not found", search)
	}
	_, err := s.Seek(int64(index), 1)
	return err
}

func (s *StringScanner) SkipPast(search string) error {
	err := s.SkipTo(search)
	if err != nil {
		return err
	}
	_, err = s.Seek(int64(len(search)), 1)
	return err
}

func (s *StringScanner) SkipWhile(f func (r rune) bool) {
	for s.i < len(s.s) {
		r, _, err := s.ReadRune()
		if err != nil {
			break
		}
		if !f(r) {
			s.UnreadRune()
			break
		}
	}
}

func (s *StringScanner) SkipSpace() {
	s.SkipWhile(func (r rune) bool {
		return unicode.IsSpace(r)
	})
}

func (s *StringScanner) Int() (int, error) {
	n := 0
	ok := false
	_, err := s.ReadWhile(func (r rune) bool {
		if !unicode.IsDigit(r) {
			return false
		}
		ok = true
		n *= 10
		n += int(r) - int('0')
		return true
	})
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, errors.New("no digits")
	}
	return n, nil
}

func (s *StringScanner) Lit(text string) error {
	if s.i + len(text) > len(s.s) || s.s[s.i:s.i + len(text)] != text {
		return fmt.Errorf("failed to parse literal '%s'", text)
	}
	s.Seek(int64(len(text)), 1)
	return nil
}
