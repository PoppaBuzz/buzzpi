package version

import (
	"strings"
	"testing"
)

func TestConstants(t *testing.T) {
	t.Run("Version", func(t *testing.T) {
		if Version == "" {
			t.Error("Version must not be empty")
		}
	})
	t.Run("Commit", func(t *testing.T) {
		// Can be "unknown" but not empty
		if Commit == "" {
			t.Error("Commit must not be empty")
		}
	})
	t.Run("Date", func(t *testing.T) {
		if Date == "" {
			t.Error("Date must not be empty")
		}
	})
	t.Run("MinRuntimeVersion", func(t *testing.T) {
		if MinRuntimeVersion == "" {
			t.Error("MinRuntimeVersion must not be empty")
		}
	})
	t.Run("BPPVersion", func(t *testing.T) {
		if BPPVersion != "1" {
			t.Errorf("BPPVersion = %q, want \"1\"", BPPVersion)
		}
	})
	t.Run("UserAgent", func(t *testing.T) {
		if !strings.HasPrefix(UserAgent, "BuzzPi-Runtime/") {
			t.Errorf("UserAgent = %q, want prefix \"BuzzPi-Runtime/\"", UserAgent)
		}
		if !strings.Contains(UserAgent, Version) {
			t.Errorf("UserAgent = %q, should contain Version = %q", UserAgent, Version)
		}
	})
}

func TestInfo(t *testing.T) {
	info := Info()
	if !strings.HasPrefix(info, "BuzzPi Runtime ") {
		t.Errorf("Info() = %q, want prefix \"BuzzPi Runtime \"", info)
	}
	if !strings.Contains(info, Commit) {
		t.Errorf("Info() = %q, does not contain Commit = %q", info, Commit)
	}
	if !strings.Contains(info, Date) {
		t.Errorf("Info() = %q, does not contain Date = %q", info, Date)
	}
}

func TestGoVersion(t *testing.T) {
	gv := GoVersion()
	if gv == "" {
		t.Error("GoVersion() returned empty string")
	}
	if !strings.HasPrefix(gv, "go") && !strings.Contains(gv, ".") {
		t.Errorf("GoVersion() = %q, looks invalid", gv)
	}
}
