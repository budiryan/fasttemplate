package fasttemplate

import (
	"io"
	"testing"
)

func TestEmptyTemplate(t *testing.T) {
	tpl := New("", "[", "]")

	s, _ := tpl.ExecuteString(map[string]interface{}{"foo": "bar", "aaa": "bbb"})
	if s != "" {
		t.Fatalf("unexpected string returned %q. Expected empty string", s)
	}
}

func TestEmptyTagStart(t *testing.T) {
	_, err := NewTemplate("foobar", "", "]")
	errExpected := ErrEmptyStartTag
	if errExpected != err {
		t.Fatalf("error: %+v is expected", errExpected)
	}
}

func TestEmptyTagEnd(t *testing.T) {
	_, err := NewTemplate("foobar", "[", "")
	errExpected := ErrEmptyEndTag
	if errExpected != err {
		t.Fatalf("error: %+v is expected", errExpected)
	}
}

func TestNoTags(t *testing.T) {
	template := "foobar"
	tpl := New(template, "[", "]")

	s, _ := tpl.ExecuteString(map[string]interface{}{"foo": "bar", "aaa": "bbb"})
	if s != template {
		t.Fatalf("unexpected template value %q. Expected %q", s, template)
	}
}

func TestEmptyTagName(t *testing.T) {
	template := "foo[]bar"
	tpl := New(template, "[", "]")

	s, _ := tpl.ExecuteString(map[string]interface{}{"": "111", "aaa": "bbb"})
	result := "foo111bar"
	if s != result {
		t.Fatalf("unexpected template value %q. Expected %q", s, result)
	}
}

func TestOnlyTag(t *testing.T) {
	template := "[foo]"
	tpl := New(template, "[", "]")

	s, _ := tpl.ExecuteString(map[string]interface{}{"foo": "111", "aaa": "bbb"})
	result := "111"
	if s != result {
		t.Fatalf("unexpected template value %q. Expected %q", s, result)
	}
}

func TestStartWithTag(t *testing.T) {
	template := "[foo]barbaz"
	tpl := New(template, "[", "]")

	s, _ := tpl.ExecuteString(map[string]interface{}{"foo": "111", "aaa": "bbb"})
	result := "111barbaz"
	if s != result {
		t.Fatalf("unexpected template value %q. Expected %q", s, result)
	}
}

func TestEndWithTag(t *testing.T) {
	template := "foobar[foo]"
	tpl := New(template, "[", "]")

	s, _ := tpl.ExecuteString(map[string]interface{}{"foo": "111", "aaa": "bbb"})
	result := "foobar111"
	if s != result {
		t.Fatalf("unexpected template value %q. Expected %q", s, result)
	}
}

func TestTemplateReset(t *testing.T) {
	template := "foo{bar}baz"
	tpl := New(template, "{", "}")
	s, _ := tpl.ExecuteString(map[string]interface{}{"bar": "111"})
	result := "foo111baz"
	if s != result {
		t.Fatalf("unexpected template value %q. Expected %q", s, result)
	}

	template = "[xxxyyyzz"
	if err := tpl.Reset(template, "[", "]"); err == nil {
		t.Fatalf("expecting error for unclosed tag on %q", template)
	}

	template = "[xxx]yyy[zz]"
	if err := tpl.Reset(template, "[", "]"); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	s, _ = tpl.ExecuteString(map[string]interface{}{"xxx": "11", "zz": "2222"})
	result = "11yyy2222"
	if s != result {
		t.Fatalf("unexpected template value %q. Expected %q", s, result)
	}
}

func TestDuplicateTags(t *testing.T) {
	template := "[foo]bar[foo][foo]baz"
	tpl := New(template, "[", "]")

	s, _ := tpl.ExecuteString(map[string]interface{}{"foo": "111", "aaa": "bbb"})
	result := "111bar111111baz"
	if s != result {
		t.Fatalf("unexpected template value %q. Expected %q", s, result)
	}
}

func TestMultipleTags(t *testing.T) {
	template := "foo[foo]aa[aaa]ccc"
	tpl := New(template, "[", "]")

	s, _ := tpl.ExecuteString(map[string]interface{}{"foo": "111", "aaa": "bbb"})
	result := "foo111aabbbccc"
	if s != result {
		t.Fatalf("unexpected template value %q. Expected %q", s, result)
	}
}

func TestLongDelimiter(t *testing.T) {
	template := "foo{{{foo}}}bar"
	tpl := New(template, "{{{", "}}}")

	s, _ := tpl.ExecuteString(map[string]interface{}{"foo": "111", "aaa": "bbb"})
	result := "foo111bar"
	if s != result {
		t.Fatalf("unexpected template value %q. Expected %q", s, result)
	}
}

func TestIdenticalDelimiter(t *testing.T) {
	template := "foo@foo@foo@aaa@"
	tpl := New(template, "@", "@")

	s, _ := tpl.ExecuteString(map[string]interface{}{"foo": "111", "aaa": "bbb"})
	result := "foo111foobbb"
	if s != result {
		t.Fatalf("unexpected template value %q. Expected %q", s, result)
	}
}

func TestIdenticalDelimiter2(t *testing.T) {
	template := "foo@foo@foo@aaa@hhh"
	tpl := New(template, "@", "@")

	s, _ := tpl.ExecuteString(map[string]interface{}{"foo": "111", "aaa": "bbb"})
	result := "foo111foobbbhhh"
	if s != result {
		t.Fatalf("unexpected template value %q. Expected %q", s, result)
	}
}

func TestMissingClosingTagsInTemplate(t *testing.T) {
	template := "foo{{a}} {{b"
	_, err := NewTemplate(template, "{{", "}}")
	if err == nil {
		t.Fatalf("error should not be nil")
	}
}

func TestNestedTags(t *testing.T) {
	template := "text1 {{ {{{{foo}}}} }} text2 {{ {{||bar}}|| }} text3"
	tpl := New(template, "{{", "}}")

	s, _ := tpl.ExecuteString(map[string]interface{}{
		"{{{{foo}}}}": "foo replaced",
		"{{||bar}}||": "bar replaced",
	})
	result := "text1 foo replaced text2 bar replaced text3"
	if s != result {
		t.Fatalf("unexpected template value %q. Expected %q", s, result)
	}
}

func TestDlimitersWithDistinctSize(t *testing.T) {
	template := "foo<?phpaaa?>bar<?phpzzz?>"
	tpl := New(template, "<?php", "?>")

	s, _ := tpl.ExecuteString(map[string]interface{}{"zzz": "111", "aaa": "bbb"})
	result := "foobbbbar111"
	if s != result {
		t.Fatalf("unexpected template value %q. Expected %q", s, result)
	}
}

func TestEmptyValue(t *testing.T) {
	template := "foobar[foo]"
	tpl := New(template, "[", "]")

	s, _ := tpl.ExecuteString(map[string]interface{}{"foo": "", "aaa": "bbb"})
	result := "foobar"
	if s != result {
		t.Fatalf("unexpected template value %q. Expected %q", s, result)
	}
}

func TestNoValue(t *testing.T) {
	template := "foobar[foo]x[aaa]"
	tpl := New(template, "[", "]")

	s, _ := tpl.ExecuteString(map[string]interface{}{"aaa": "bbb"})
	result := "foobarxbbb"
	if s != result {
		t.Fatalf("unexpected template value %q. Expected %q", s, result)
	}
}

func TestNoEndDelimiter(t *testing.T) {
	template := "foobar[foo"
	_, err := NewTemplate(template, "[", "]")
	if err == nil {
		t.Fatalf("expected non-nil error. got nil")
	}

	expectPanic(t, func() { New(template, "[", "]") })
}

func TestUnsupportedValue(t *testing.T) {
	template := "foobar[foo]"
	tpl := New(template, "[", "]")

	_, err := tpl.ExecuteString(map[string]interface{}{"foo": 123, "aaa": "bbb"})
	if err != ErrInvalidTag {
		t.Fatalf("expected %+v, got %+v", ErrInvalidTag, err)
	}
}

func TestMixedValues(t *testing.T) {
	template := "foo[foo]bar[bar]baz[baz]"
	tpl := New(template, "[", "]")

	s, _ := tpl.ExecuteString(map[string]interface{}{
		"foo": "111",
		"bar": []byte("bbb"),
		"baz": TagFunc(func(w io.Writer, tag string) (int, error) { return w.Write([]byte(tag)) }),
	})
	result := "foo111barbbbbazbaz"
	if s != result {
		t.Fatalf("unexpected template value %q. Expected %q", s, result)
	}
}

func expectPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("missing panic")
		}
	}()
	f()
}
