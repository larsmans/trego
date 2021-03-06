package tre

import (
	"bytes"
	"fmt"
	"testing"
)

var validREs = []string{
	"foo",
	"fo{2}{~1}",
	"^bar$",
	"(foo){~2} bar$",
}

var invalidREs = []string{
	"(foo!",
	"{bla",
	"irregular expression{)",
}

func Test_Compile(t *testing.T) {
	for _, s := range validREs {
		re, err := Compile(s, 0)
		switch {
		case err != nil:
			msg := fmt.Sprintf("failed to compile %s: %s", s, err)
			t.Error(msg)
		case re == nil:
			t.Error(fmt.Sprintf("compiling %s returned nil", s))
		default:
			t.Log(fmt.Sprintf("successfully compiled %s", s))
		}
		if re.String() != s {
			msg := fmt.Sprintf("wrong String() for %s: %s",
				re.String(), s)
			t.Error(msg)
		}
	}

	for _, s := range invalidREs {
		re, err := Compile(s, 0)
		switch {
		case err == nil:
			msg := fmt.Sprintf("Compile(%s) should have failed", s)
			t.Error(msg)
		case re != nil:
			msg := fmt.Sprintf("error %s but non-nil re for %s",
				err, s)
			t.Error(msg)
		default:
			t.Log(fmt.Sprintf("correct failure %s for %s", err, s))
		}
	}
}

// Test Find and FindString.
func Test_Find(t *testing.T) {
	pattern := "(regular){~1}\\s+(expression){~2}"
	text := "match this with your regulor  exzpressyon!"

	re, err := Compile(pattern, 0)
	if err != nil {
		msg := fmt.Sprintf("failed to compile %s: %s", pattern, err)
		t.Error(msg)
	}

	match := re.Find([]byte(text))
	switch {
	case match == nil:
		msg := fmt.Sprintf("failed to Find %s in \"%s\"", re, text)
		t.Error(msg)
	case bytes.Compare(match, []byte("regulor  exzpressyon")) != 0:
		msg := fmt.Sprintf("%s found wrong substring \"%s\"", re, match)
		t.Error(msg)
	default:
		t.Log(fmt.Sprintf("successfully matched \"%s\" with %s",
			re, text))
	}

	smatch := re.FindString(text)
	switch {
	case smatch == "":
		msg := fmt.Sprintf("failed to FindString %s in \"%s\"",
			re, text)
		t.Error(msg)
	case smatch != "regulor  exzpressyon":
		msg := fmt.Sprintf("%s found wrong substring \"%s\"",
			re, smatch)
		t.Error(msg)
	default:
		t.Log(fmt.Sprintf("successfully matched \"%s\" with %s",
			re, text))
	}
}
