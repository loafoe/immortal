package immortal

import (
	"flag"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

/* Test Helpers */
func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected: %v (type %v)  Got: %v (type %v)", a, reflect.TypeOf(a), b, reflect.TypeOf(b))
	}
}

func TestParseExist(t *testing.T) {
	p := Parse{}
	expect(t, false, p.exists("/dev/null/non-existent"))
	expect(t, true, p.exists("/"))
}

func TestParseHelp(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "-h"}
	p := &Parse{}
	var helpCalled = false
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.Usage = func() { helpCalled = true }
	p.Parse(fs)
	if !helpCalled {
		t.Fatal("help was not called")
	}
}

func TestParseDefault(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", ""}
	p := &Parse{}
	var helpCalled = false
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.Usage = func() { helpCalled = true }
	flags, err := p.Parse(fs)
	if err != nil {
		t.Error(err)
	}
	if helpCalled {
		t.Error("help called for regular flag")
	}
	expect(t, false, flags.Ctrl)
	expect(t, false, flags.Version)
	expect(t, "", flags.Configfile)
	expect(t, "", flags.Wrkdir)
	expect(t, "", flags.Envdir)
	expect(t, "", flags.FollowPid)
	expect(t, "", flags.Logfile)
	expect(t, "", flags.Logger)
	expect(t, "", flags.ChildPid)
	expect(t, "", flags.ParentPid)
	expect(t, "", flags.User)
}

func TestParseFlags(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	var flagTest = []struct {
		flag     []string
		name     string
		expected interface{}
	}{
		{[]string{"cmd", "-v"}, "Version", true},
		{[]string{"cmd", "-ctrl"}, "Ctrl", true},
		{[]string{"cmd", "-c", "run.yml"}, "Configfile", "run.yml"},
		{[]string{"cmd", "-d", "/arena/wrkdir"}, "Wrkdir", "/arena/wrkdir"},
		{[]string{"cmd", "-e", "/path/to/envdir"}, "Envdir", "/path/to/envdir"},
		{[]string{"cmd", "-f", "/path/to/pid"}, "FollowPid", "/path/to/pid"},
		{[]string{"cmd", "-l", "/path/to/log"}, "Logfile", "/path/to/log"},
		{[]string{"cmd", "-logger", "logger"}, "Logger", "logger"},
		{[]string{"cmd", "-p", "/path/to/child"}, "ChildPid", "/path/to/child"},
		{[]string{"cmd", "-P", "/path/to/parent"}, "ParentPid", "/path/to/parent"},
		{[]string{"cmd", "-u", "nbari"}, "User", "nbari"},
	}
	var helpCalled = false
	for _, f := range flagTest {
		os.Args = f.flag
		p := &Parse{}
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.Usage = func() { helpCalled = true }
		flags, err := p.Parse(fs)
		if err != nil {
			t.Error(err)
		}
		if helpCalled {
			t.Error("help called for regular flag")
			helpCalled = false // reset for next test
		}
		refValue := reflect.ValueOf(flags).Elem().FieldByName(f.name)
		switch refValue.Kind() {
		case reflect.Bool:
			expect(t, f.expected, refValue.Bool())
		case reflect.String:
			expect(t, f.expected, refValue.String())
		}
	}
}

func TestParseArgsHelp(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "-h"}
	parser := &Parse{}
	var helpCalled = false
	fs := flag.NewFlagSet("TestParseArgsHelp", flag.ContinueOnError)
	fs.Usage = func() { helpCalled = true }
	ParseArgs(parser, fs)
	if !helpCalled {
		t.Fatal("help was not called")
	}
}

func TestParseArgsVersion(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "-v"}
	parser := &Parse{}
	var helpCalled = false
	fs := flag.NewFlagSet("TestParseArgsVersion", flag.ContinueOnError)
	fs.Usage = func() { helpCalled = true }
	ParseArgs(parser, fs)
	if helpCalled {
		t.Error("help called for regular flag")
	}
}

func TestParseArgsVersion2(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "-v", "-c", "xyz"}
	parser := &Parse{}
	var helpCalled = false
	fs := flag.NewFlagSet("TestParseArgsVersion2", flag.ContinueOnError)
	fs.Usage = func() { helpCalled = true }
	ParseArgs(parser, fs)
	if helpCalled {
		t.Error("help called for regular flag")
	}
}

func TestParseArgsNoargs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd"}
	parser := &Parse{}
	var helpCalled = false
	fs := flag.NewFlagSet("TestParseArgsNoargs", flag.ContinueOnError)
	fs.Usage = func() { helpCalled = true }
	_, err := ParseArgs(parser, fs)
	if helpCalled {
		t.Error("help called for regular flag")
	}
	if err == nil {
		t.Error("Expecting error")
	}
}

func TestParseArgsTable(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	dir, err := ioutil.TempDir("", "TestParseArgsTable")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	var flagTest = []struct {
		flag     []string
		expected bool
	}{
		{[]string{"cmd", "-v"}, false},
		{[]string{"cmd", "-ctrl"}, true},
		{[]string{"cmd", "-c", "run.yml"}, true},
		{[]string{"cmd", "-c", "run.yml", "cmd"}, true},
		{[]string{"cmd", "-d", "/arena/wrkdir"}, true},
		{[]string{"cmd", "-d", "/dev/null", "cmd"}, true},
		{[]string{"cmd", "-d", dir, "cmd"}, false},
		{[]string{"cmd", "-e", "/path/to/envdir"}, true},
		{[]string{"cmd", "-e", "/dev/null", "cmd"}, true},
		{[]string{"cmd", "-e", dir, "cmd"}, false},
		{[]string{"cmd", "-f", "/path/to/pid"}, true},
		{[]string{"cmd", "-f", "/path/to/pid", "cmd"}, false},
		{[]string{"cmd", "-l", "/path/to/log"}, true},
		{[]string{"cmd", "-logger", "logger"}, true},
		{[]string{"cmd", "-p", "/path/to/child"}, true},
		{[]string{"cmd", "-P", "/path/to/parent"}, true},
		{[]string{"cmd", "-u", "root"}, true},
		{[]string{"cmd", "-u", "root", "cmd"}, false},
		{[]string{"cmd", "-u", "toor-nonexistent", "cmd"}, true},
	}
	var helpCalled = false
	for _, f := range flagTest {
		os.Args = f.flag
		parser := new(Parse)
		fs := flag.NewFlagSet("TestParseArgsTable", flag.ContinueOnError)
		fs.Usage = func() { helpCalled = true }
		_, err := ParseArgs(parser, fs)
		if f.expected {
			if err == nil {
				t.Error("Expecting error")
			}
		} else {
			if err != nil {
				t.Error(err)
			}
		}
		if helpCalled {
			t.Error("help called for regular flag")
			helpCalled = false // reset for next test
		}
	}
}