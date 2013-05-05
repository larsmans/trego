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
	"errors"
	"runtime"
	"unsafe"
)

type Regexp struct {
	preg C.regex_t
	str string
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
	if e := C.tre_regcomp(&re.preg, cstr, cflags); e != 0 {
		return nil, re.regerror(e)
	}
	runtime.SetFinalizer(&re, cleanup)

	re.str = expr
	return &re, nil
}

func cleanup(re *Regexp) {
	C.tre_regfree(&re.preg)
}

func (re *Regexp) regerror(code C.int) error {
	length := C.tre_regerror(code, &re.preg, nil, 0)

	buf := make([]byte, length)
	pbuf := (*C.char)(unsafe.Pointer(&buf[0]))
	C.tre_regerror(code, &re.preg, pbuf, length)

	return errors.New(string(buf))
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
		panic(re.regerror(e))
	}

	return []int{int(pmatch.rm_so), int(pmatch.rm_eo)}
}

// Whether re uses backreferences.
func (re *Regexp) HasBackrefs() bool {
	return C.tre_have_backrefs(&re.preg) != 0
}

func (re *Regexp) String() string {
	return re.str
}
