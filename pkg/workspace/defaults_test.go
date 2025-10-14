package workspace

import "testing"

func TestEnsureDefaultPorts(t *testing.T) {
	t.Run("no input returns defaults", func(t *testing.T) {
		ports := EnsureDefaultPorts(nil)
		if len(ports) != 1 || ports[0] != DefaultVSCodePort {
			t.Fatalf("expected [%d], got %v", DefaultVSCodePort, ports)
		}
	})

	t.Run("keeps custom ports and injects vscode", func(t *testing.T) {
		ports := EnsureDefaultPorts([]int32{4000, -1, DefaultVSCodePort})
		if len(ports) != 2 {
			t.Fatalf("expected 2 ports, got %d: %v", len(ports), ports)
		}
		if ports[0] != 4000 || ports[1] != DefaultVSCodePort {
			t.Fatalf("expected [4000 %d], got %v", DefaultVSCodePort, ports)
		}
	})

	t.Run("adds vscode when missing", func(t *testing.T) {
		ports := EnsureDefaultPorts([]int32{2222})
		if len(ports) != 2 {
			t.Fatalf("expected 2 ports, got %d: %v", len(ports), ports)
		}
		if ports[0] != 2222 || ports[1] != DefaultVSCodePort {
			t.Fatalf("expected [2222 %d], got %v", DefaultVSCodePort, ports)
		}
	})
}

func TestMergeEnv(t *testing.T) {
	defaults := map[string]string{
		EnvVSCodeQuality:  "stable",
		EnvPasswordAccess: DefaultPasswordAccess,
		"FOO":             "bar",
	}
	user := map[string]string{
		"FOO":            "baz",
		"CUSTOM":         "value",
		"EMPTY_OVERRIDE": "",
	}

	merged := MergeEnv(user, defaults)
	if merged["FOO"] != "baz" {
		t.Fatalf("expected user value to win, got %q", merged["FOO"])
	}
	if _, ok := merged["EMPTY_OVERRIDE"]; ok {
		t.Fatalf("expected empty override to be dropped, got %v", merged)
	}
	if merged[EnvPasswordAccess] != DefaultPasswordAccess {
		t.Fatalf("expected password access default, got %q", merged[EnvPasswordAccess])
	}
	if merged["CUSTOM"] != "value" {
		t.Fatalf("expected custom env value, got %q", merged["CUSTOM"])
	}
}

func TestCopyEnv(t *testing.T) {
	original := map[string]string{"A": "1"}
	copy := CopyEnv(original)
	if copy["A"] != "1" {
		t.Fatalf("expected copy to contain A=1, got %v", copy)
	}
	copy["A"] = "2"
	if original["A"] != "1" {
		t.Fatalf("expected copy to be isolated, original mutated to %q", original["A"])
	}
}
