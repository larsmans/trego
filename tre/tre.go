/*
Go binding for the TRE approximate string matching library.

Copyright 2013 Lars Buitinck

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package tre

/*
#cgo LDFLAGS: -ltre
#include <tre/tre.h>
#include <stdlib.h>
*/
import "C"

import (
	"runtime"
	"unsafe"
)

type Regexp struct {
	preg C.regex_t
}

type Flags int

const (
	CaseI Flags	= C.REG_ICASE
	Nosub		= C.REG_NOSUB
	Newline		= C.REG_NEWLINE
	Literal		= C.REG_LITERAL
	RightAssoc	= C.REG_RIGHT_ASSOC
	Eager		= C.REG_UNGREEDY
)

// Compile regular expression.
func Compile(expr string, flags Flags) (*Regexp, error) {
	var re Regexp

	cstr := C.CString(expr)
	defer C.free(unsafe.Pointer(cstr))

	cflags := C.int(flags) | C.REG_EXTENDED
	if ec := C.tre_regcomp(&re.preg, cstr, cflags); ec != 0 {
		return nil, treError(ec)
	}

	runtime.SetFinalizer(&re, cleanup)

	return &re, nil
}

func cleanup(re *Regexp) {
	C.tre_regfree(&re.preg)
}

type treError int

func (e treError) Error() (msg string) {
	switch e {
	case C.REG_BADPAT:
		msg = "invalid multibyte sequence"
	case C.REG_ECOLLATE:
		msg = "invalid collating element referenced"
	case C.REG_ECTYPE:
		msg = "unknown character class name"
	case C.REG_EESCAPE:
		msg = "last character of regexp was \\"
	case C.REG_ESUBREG:
		msg = "invalid backreference"
	case C.REG_EBRACK:
		msg = "unbalanced [ or ]"
	case C.REG_EPAREN:
		msg = "unbalanced (, \\(, ) or \\)"
	case C.REG_EBRACE:
		msg = "unbalanced {, \\{, } or \\}"
	case C.REG_BADBR:
		msg = "{} content invalid"
	case C.REG_ERANGE:
		msg = "invalid character range"
	case C.REG_ESPACE:
		msg = "out of memory"
	case C.REG_BADRPT:
		msg = "invalid use of repetition operator"
	}
	return
}

// Like regexp.Find in the standard library.
//
// Panics when an error occurs inside TRE (such as out of memory).
func (re *Regexp) Find(b []byte) []byte {
	loc := re.FindIndex(b)
	if loc == nil {
		return nil
	}
	return b[loc[0]:loc[1]]
}

// Like regexp.FindIndex in the standard library.
//
// Panics when an error occurs inside TRE (such as out of memory).
func (re *Regexp) FindIndex(b []byte) []int {
	var pmatch C.regmatch_t

	p := (*C.char)(unsafe.Pointer(&b[0]))
	e := C.tre_regnexec(&re.preg, p, C.size_t(len(b)), 1, &pmatch, 0)
	switch e {
	case 0:
	case C.REG_NOMATCH:
		return nil
	default:
		panic(treError(e))
	}

	return []int{int(pmatch.rm_so), int(pmatch.rm_eo)}
}

// Whether re uses backreferences.
func (re *Regexp) HasBackrefs() bool {
	return C.tre_have_backrefs(&re.preg) != 0
}
