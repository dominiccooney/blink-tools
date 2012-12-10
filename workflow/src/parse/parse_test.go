package parse

import (
	"testing"
	"unicode"
)

func TestHere_InitialPosition(t *testing.T) {
	s := NewUtf8Scanner("")
	if row, col := s.Here(); row != 0 || col != 0 {
		t.Errorf("the initial position should be (0, 0) was (%d, %d)", row, col)
	}
}

func TestHere(t *testing.T) {
	s := NewUtf8Scanner("a\nb\ncrabapples")
	for i := 0; i < 10; i++ {
		s.ReadRune()
	}
	if row, col := s.Here(); row != 2 || col != 6 {
		t.Errorf("position should be (2, 6) was (%d, %d)", row, col)
	}
	for i := 0; i < 8; i++ {
		s.UnreadRune()
	}
	if row, col := s.Here(); row != 1 || col != 0 {
		t.Errorf("position should be (1, 0) was (%d, %d)", row, col)
	}
}

func TestSkipTo(t *testing.T) {
	input := "now is the time for all good men to come to the aid of the party"
	s := NewUtf8Scanner(input)
	s.SkipTo("to ")
	_, col := s.Here()
	if input[col:] != "to come to the aid of the party" {
		t.Errorf("should have skipped to the first 'to' but got '%s'", input[col:])
	}
	s.ReadRune()
	s.SkipTo("to ")
	_, col = s.Here()
	if input[col:] != "to the aid of the party" {
		t.Errorf("should have skipped to the next 'to' but got '%s'", input[col:])
	}
}

func TestSkipPast(t *testing.T) {
	s := NewUtf8Scanner("crabzebra")
	s.SkipPast("ab")
	r, _, _ := s.ReadRune()
	if r != 'z' {
		t.Errorf("expected to be positioned at zebra but got '%c'", r)
	}
}

func TestReadWhile(t *testing.T) {
	s := NewUtf8Scanner("1234x")
	yield, _ := s.ReadWhile(func (r rune) bool {
		return unicode.IsDigit(r)
	})
	if yield != "1234" {
		t.Errorf("expected '1234' but got '%s'", yield)
	}
}

func TestInt(t *testing.T) {
	s := NewUtf8Scanner("1024 apples baked in a pie")
	expected := 1024
	n, _ := s.Int()
	if n != expected {
		t.Errorf("expected %d but got %d", expected, n)
	}
}

func TestLit(t *testing.T) {
	s := NewUtf8Scanner("crabapplebear")
	if err := s.Lit("crab"); err != nil {
		t.Errorf("expected to parse 'crab' but failed with %s", err.Error())
	}
	if err := s.Lit("bear"); err == nil {
		t.Errorf("expected to fail to parse 'bear' but succeeded")
	}
	if err := s.Lit("apple"); err != nil {
		t.Errorf("expected to parse 'apple' but failed with %s", err.Error())
	}
	if err := s.Lit("past the end should not panic"); err == nil {
		t.Errorf("expected to fail looking for literal at end-of-input but succeeded")
	}
}
