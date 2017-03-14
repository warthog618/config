package flag

import (
	"github.com/warthog618/config"
	"os"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	args := []string{"-avbcvv", "--config-file", "woot"}
	shorts := map[byte]string{
		'c': "config-file",
		'b': "bonus",
		'v': "logging-verbosity",
	}
	f, err := New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	if len(f.nodes) != 2 {
		t.Errorf("incorrect number of nodes, expected 2, got %v", len(f.nodes))
	}
	if len(f.config) != 3 {
		t.Errorf("incorrect number of leaves, expected 3, got %v", len(f.config))
	}
	// test provides config.Reader interface.
	cfg := config.New()
	cfg.AppendReader(f)
}

func TestArgs(t *testing.T) {
	args := []string{
		"-v",
		"-n=44",
		"--leaf", "42",
		"--slice=a,b",
		"--nested-slice", "c,d",
	}
	shorts := map[byte]string{
		'n': "nested-leaf",
		'v': "logging-verbosity",
	}
	f, err := New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	if len(f.Args()) != 0 {
		t.Errorf("found args when none provided.")
	}
	args = append(args, "arg1")
	f, err = New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	fargs := f.Args()
	expected := args[len(args)-1:]
	if !reflect.DeepEqual(fargs, expected) {
		t.Errorf("incorrect args returned - got %v, expected %v", fargs, expected)
	}
	args = append(args, "arg2")
	f, err = New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	fargs = f.Args()
	expected = args[len(args)-2:]
	if !reflect.DeepEqual(fargs, expected) {
		t.Errorf("incorrect args returned - got %v, expected %v", fargs, expected)
	}
	// terminated parsing
	args = []string{
		"-v",
		"-n=44",
		"--leaf", "42",
		"--",
		"--slice=a,b",
		"--nested-slice", "c,d",
		"arg1",
	}
	f, err = New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	fargs = f.Args()
	expected = args[5:]
	if !reflect.DeepEqual(fargs, expected) {
		t.Errorf("incorrect args returned - got %v, expected %v", fargs, expected)
	}
	// default to os.Args
	os.Args = args
	f, err = New([]string(nil), shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	fargs = f.Args()
	expected = args[5:]
	if !reflect.DeepEqual(fargs, expected) {
		t.Errorf("incorrect args returned - got %v, expected %v", fargs, expected)
	}
}

func TestNArg(t *testing.T) {
	args := []string{
		"-v",
		"-n=44",
		"--leaf", "42",
		"--slice=a,b",
		"--nested-slice", "c,d",
	}
	shorts := map[byte]string{
		'n': "nested-leaf",
		'v': "logging-verbosity",
	}
	f, err := New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	if f.NArg() != 0 {
		t.Errorf("found args when none provided.")
	}
	args = append(args, "arg1")
	f, err = New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	expected := 1
	if f.NArg() != expected {
		t.Errorf("miscounted args - got %v, expected %v.", f.NArg(), expected)
	}
	args = append(args, "arg2")
	f, err = New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	expected = 2
	if f.NArg() != expected {
		t.Errorf("miscounted args - got %v, expected %v.", f.NArg(), expected)
	}
	// terminate parsing
	args = []string{
		"-v",
		"-n=44",
		"--leaf", "42",
		"--",
		"--slice=a,b",
		"--nested-slice", "c,d",
		"arg1",
	}
	f, err = New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	expected = 4
	if f.NArg() != expected {
		t.Errorf("miscounted flags - got %v, expected %v.", f.NArg(), expected)
	}
}

func TestNFlag(t *testing.T) {
	args := []string{
		"-v",
		"-n=44",
		"--leaf", "42",
		"--slice=a,b",
		"--nested-slice", "c,d",
	}
	shorts := map[byte]string{
		'n': "nested-leaf",
		'v': "logging-verbosity",
	}
	f, err := New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	expected := 5
	if f.NFlag() != expected {
		t.Errorf("miscounted flags - got %v, expected %v.", f.NFlag(), expected)
	}
	// short and long form of same flag
	args = []string{
		"-v",
		"--logging-verbosity",
	}
	f, err = New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	expected = 1
	if f.NFlag() != expected {
		t.Errorf("miscounted flags - got %v, expected %v.", f.NFlag(), expected)
	}
	// terminate parsing
	args = []string{
		"-v",
		"-n=44",
		"--leaf", "42",
		"--",
		"--slice=a,b",
		"--nested-slice", "c,d",
	}
	f, err = New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	expected = 3
	if f.NFlag() != expected {
		t.Errorf("miscounted flags - got %v, expected %v.", f.NFlag(), expected)
	}
}

func TestSetShortFlag(t *testing.T) {
	args := []string{
		"-abc",
	}
	shorts := map[byte]string{
		'a': "nested-leaf",
		'v': "logging-verbosity",
	}
	f, err := New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	expected := 1
	if f.NFlag() != expected {
		t.Errorf("found incorrect number of flags, got %v, expected %v", f.NFlag(), expected)
	}
	if v, ok := f.Read("nested.leaf"); !ok {
		t.Errorf("couldn't read nested.leaf")
	} else if v != expected {
		t.Errorf("incorrect value for nested.leaf, got %v, expected %v", v, expected)
	}
	if v, ok := f.Read("bonus"); ok {
		t.Errorf("read non-existent flag bonus, got %v", v)
	}
	// add
	f.SetShortFlag('b', "bonus")
	expected = 2
	if f.NFlag() != expected {
		t.Errorf("found incorrect number of flags, got %v, expected %v", f.NFlag(), expected)
	}
	expected = 1
	if v, ok := f.Read("nested.leaf"); !ok {
		t.Errorf("couldn't read nested.leaf")
	} else if v != expected {
		t.Errorf("incorrect value for nested.leaf, got %v, expected %v", v, expected)
	}
	if v, ok := f.Read("bonus"); !ok {
		t.Errorf("couldn't read nested.leaf")
	} else if v != expected {
		t.Errorf("incorrect value for bonus, got %v, expected %v", v, expected)
	}
	// replace
	f.SetShortFlag('a', "addon")
	expected = 2
	if f.NFlag() != expected {
		t.Errorf("found incorrect number of flags, got %v, expected %v", f.NFlag(), expected)
	}
	expected = 1
	if v, ok := f.Read("nested.leaf"); ok {
		t.Errorf("read non-existent flag nested.leaf, got %v", v)
	}
	if v, ok := f.Read("addon"); !ok {
		t.Errorf("couldn't read addon")
	} else if v != expected {
		t.Errorf("incorrect value for addon, got %v, expected %v", v, expected)
	}
	if v, ok := f.Read("bonus"); !ok {
		t.Errorf("couldn't read nested.leaf")
	} else if v != expected {
		t.Errorf("incorrect value for bonus, got %v, expected %v", v, expected)
	}
}

func TestReaderContains(t *testing.T) {
	args := []string{
		"-v",
		"-n=44",
		"--leaf", "42",
		"--slice=a,b",
		"--nested-slice", "c,d",
	}
	shorts := map[byte]string{
		'n': "nested-leaf",
		'v': "logging-verbosity",
	}
	f, err := New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	// leaf
	if !f.Contains("leaf") {
		t.Errorf("doesn't contain leaf")
	}
	if !f.Contains("nested.leaf") {
		t.Errorf("doesn't contain nested.leaf")
	}
	if !f.Contains("logging.verbosity") {
		t.Errorf("doesn't contain logging.verbosity")
	}
	// node
	if !f.Contains("nested") {
		t.Errorf("doesn't contain nested")
	}
	// neither
	if f.Contains("nonsense") {
		t.Errorf("contains nonsense")
	}
}

func TestReaderRead(t *testing.T) {
	args := []string{
		"-vvv",
		"-n=44",
		"--logging-verbosity",
		"--leaf", "42",
		"--slice=a,b",
		"--nested-slice", "c,d",
	}
	shorts := map[byte]string{
		'n': "nested-leaf",
		'v': "logging-verbosity",
	}
	f, err := New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	// leaf
	expected := "42"
	if v, ok := f.Read("leaf"); ok {
		if v != expected {
			t.Errorf("read leaf %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read leaf")
	}
	expectedSlice := []string{"a", "b"}
	if v, ok := f.Read("slice"); ok {
		if !reflect.DeepEqual(v, expectedSlice) {
			t.Errorf("read slice %v, expected %v", v, expectedSlice)
		}
	} else {
		t.Errorf("failed to read slice")
	}
	expectedInt := 4
	if v, ok := f.Read("logging.verbosity"); ok {
		if v != expectedInt {
			t.Errorf("read logging.verbosity %v, expected %v", v, expectedInt)
		}
	} else {
		t.Errorf("failed to read leaf")
	}
	expected = "44"
	if v, ok := f.Read("nested.leaf"); ok {
		if v != expected {
			t.Errorf("read nested.leaf %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read nested.leaf")
	}
	if v, ok := f.Read("nested.slice"); ok {
		if !reflect.DeepEqual(v, []string{"c", "d"}) {
			t.Errorf("read slice %v, expected %v", v, []string{"c", "d"})
		}
	} else {
		t.Errorf("failed to read slice")
	}
	// node
	if v, ok := f.Read("nested"); ok {
		t.Errorf("contains nested - got %v", v)
	}
	// neither
	if v, ok := f.Read("nonsense"); ok {
		t.Errorf("contains nonsense - got %v", v)
	}
	// short grouped
	args = []string{
		"-abc",
	}
	shorts = map[byte]string{
		'a': "addon",
		'b': "bonus",
	}
	f, err = New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	expectedInt = 1
	if v, ok := f.Read("addon"); ok {
		if v != expectedInt {
			t.Errorf("read addon %v, expected %v", v, expectedInt)
		}
	} else {
		t.Errorf("failed to read addon")
	}
	if v, ok := f.Read("bonus"); ok {
		if v != expectedInt {
			t.Errorf("read bonus %v, expected %v", v, expectedInt)
		}
	} else {
		t.Errorf("failed to read bonus")
	}

	// precedence (last wins)
	args = []string{
		"--addon", "first string",
		"-abc",
	}
	f, err = New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	expectedInt = 1
	if v, ok := f.Read("addon"); ok {
		if v != expectedInt {
			t.Errorf("read addon %v, expected %v", v, expectedInt)
		}
	} else {
		t.Errorf("failed to read addon")
	}
	if v, ok := f.Read("bonus"); ok {
		if v != expectedInt {
			t.Errorf("read bonus %v, expected %v", v, expectedInt)
		}
	} else {
		t.Errorf("failed to read bonus")
	}
	args = []string{
		"--addon", "first string",
		"-abc",
		"--addon", "second string",
	}
	f, err = New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	expected = "second string"
	if v, ok := f.Read("addon"); ok {
		if v != expected {
			t.Errorf("read addon %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read addon")
	}
	expectedInt = 1
	if v, ok := f.Read("bonus"); ok {
		if v != expectedInt {
			t.Errorf("read bonus %v, expected %v", v, expectedInt)
		}
	} else {
		t.Errorf("failed to read bonus")
	}
	// terminate if non-flag after group
	args = []string{
		"--addon", "first string",
		"-abc",
		"stophere",
		"--addon", "second string",
	}
	f, err = New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	expectedInt = 1
	if v, ok := f.Read("addon"); ok {
		if v != expectedInt {
			t.Errorf("read addon %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read addon")
	}
	expectedInt = 1
	if v, ok := f.Read("bonus"); ok {
		if v != expectedInt {
			t.Errorf("read bonus %v, expected %v", v, expectedInt)
		}
	} else {
		t.Errorf("failed to read bonus")
	}

	// ignore malformed flag
	args = []string{
		"--addon", "first string",
		"-abc=42",
	}
	f, err = New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	expected = "first string"
	if v, ok := f.Read("addon"); ok {
		if v != expected {
			t.Errorf("read addon %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read addon")
	}
	if v, ok := f.Read("bonus"); ok {
		t.Errorf("read non-existent flag bonus, got %v", v)
	}
}

func TestReaderSetCfgSeparator(t *testing.T) {
	args := []string{"-n=44"}
	shorts := map[byte]string{'n': "nested-leaf"}
	f, err := New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	// single
	f.SetCfgSeparator("_")
	expected := "44"
	if v, ok := f.Read("nested_leaf"); ok {
		if v != expected {
			t.Errorf("read nested_leaf %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read nested_leaf")
	}
	// multi
	f.SetCfgSeparator("_X_")
	if v, ok := f.Read("nested_x_leaf"); ok {
		if v != expected {
			t.Errorf("read nested_x_leaf %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read nested_x_leaf")
	}
	// none
	f.SetCfgSeparator("")
	if v, ok := f.Read("nestedleaf"); ok {
		if v != expected {
			t.Errorf("read nestedleaf %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read nestedleaf")
	}
}

func TestReaderSetFlagSeparator(t *testing.T) {
	args := []string{"-n=44", "--leaf", "42"}
	shorts := map[byte]string{'n': "nested-leaf"}
	f, err := New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	// single
	f.SetFlagSeparator("-")
	expected := "44"
	if v, ok := f.Read("nested.leaf"); ok {
		if v != expected {
			t.Errorf("read nested.leaf %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read nested.leaf")
	}
	// multi
	f.SetFlagSeparator("ted-")
	if v, ok := f.Read("nes.leaf"); ok {
		if v != expected {
			t.Errorf("read nes.leaf %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read nes.leaf")
	}
	// none
	f.SetFlagSeparator("")
	expected = "42"
	if v, ok := f.Read("l.e.a.f"); ok {
		if v != expected {
			t.Errorf("read l.e.a.f %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read l.e.a.f")
	}
}

func TestReaderSetListSeparator(t *testing.T) {
	args := []string{"-s", "a,#b"}
	shorts := map[byte]string{'s': "slice"}
	f, err := New(args, shorts)
	if err != nil {
		t.Fatalf("new returned error", err)
	}
	// single
	f.SetListSeparator(",")
	expectedSlice := []string{"a", "#b"}
	if v, ok := f.Read("slice"); ok {
		if !reflect.DeepEqual(v, expectedSlice) {
			t.Errorf("read slice %v, expected %v", v, expectedSlice)
		}
	} else {
		t.Errorf("failed to read slice")
	}
	// multi
	f.SetListSeparator(",#")
	expectedSlice = []string{"a", "b"}
	if v, ok := f.Read("slice"); ok {
		if !reflect.DeepEqual(v, expectedSlice) {
			t.Errorf("read slice %v, expected %v", v, expectedSlice)
		}
	} else {
		t.Errorf("failed to read slice")
	}
	// none
	f.SetListSeparator("")
	expected := "a,#b"
	if v, ok := f.Read("slice"); ok {
		if v != expected {
			t.Errorf("read slice %v, expected %v", v, expected)
		}
	} else {
		t.Errorf("failed to read slice")
	}
}