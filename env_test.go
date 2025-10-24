package flage

import (
	"os"
	"testing"
)

func TestSystemEnv(t *testing.T) {
	os.Setenv("FLAGE_TEST", "test")
	env := EnvSystem(nil)
	if env.Get("FLAGE_TEST") != "test" {
		t.Error("env.Get(FLAGE_TEST) != test")
	}

	if env.Get("FLAGE_TEST_NOT_EXIST") != "" {
		t.Error("env.Get(FLAGE_TEST_NOT_EXIST) != ''")
	}

	if _, err := env.GetOrError("FLAGE_TEST_NOT_EXIST", "error"); err == nil {
		t.Error("env.GetOrError(FLAGE_TEST_NOT_EXIST, error) == nil")
	}

	if env.GetOr("FLAGE_TEST_NOT_EXIST", "default") != "default" {
		t.Error("env.GetOr(FLAGE_TEST_NOT_EXIST, default) != default")
	}

	if env.GetOr("FLAGE_TEST", "default") != "test" {
		t.Error("env.GetOr(FLAGE_TEST, default) != test")
	}

	if _, err := env.GetOrError("FLAGE_TEST", "error"); err != nil {
		t.Error("env.GetOrError(FLAGE_TEST, error) != nil")
	}
}

func TestEnvTree(t *testing.T) {
	parent := NewEnv(nil, EnvMap{
		"FLAGE_TEST":        {"test"},
		"FLAGE_TEST_PARENT": {"parent"},
	})

	child := NewEnv(parent, EnvMap{
		"FLAGE_TEST_CHILD": {"test_child"},
		"FLAGE_TEST":       {"test_override"},
	})

	if child.Get("FLAGE_TEST") != "test_override" {
		t.Error("child.Get(FLAGE_TEST) != test_override")
	}

	if child.Get("FLAGE_TEST_PARENT") != "parent" {
		t.Error("child.Get(FLAGE_TEST_PARENT) != parent")
	}

	if child.Get("FLAGE_TEST_CHILD") != "test_child" {
		t.Error("child.Get(FLAGE_TEST_CHILD) != test_child")
	}

	if parent.Get("FLAGE_TEST") != "test" {
		t.Error("parent.Get(FLAGE_TEST) != test")
	}

	if parent.Get("FLAGE_TEST_PARENT") != "parent" {
		t.Error("parent.Get(FLAGE_TEST_PARENT) != parent")
	}

	if _, err := parent.GetOrError("FLAGE_TEST_NOT_EXIST", "error"); err == nil {
		t.Error("parent.GetOrError(FLAGE_TEST_NOT_EXIST, error) == nil")
	}

	if child.GetOr("FLAGE_TEST_NOT_EXIST", "default") != "default" {
		t.Error("parent.GetOr(FLAGE_TEST_NOT_EXIST, default) != default")
	}
}

func TestCapturingEnv(t *testing.T) {
	capture := &capturingEnvMap{}
	env := NewEnv(nil, capture)

	_ = env.Get("FLAGE_TEST")
	_ = env.GetOr("FLAGE_TEST_DEFAULT", "default")
	_, _ = env.GetOrError("FLAGE_TEST_ERROR", "error")

	capture.UsagesAsEnviron("REQUIRED")
	expected := [][2]string{
		{"FLAGE_TEST", ""},
		{"FLAGE_TEST_DEFAULT", "default"},
		{"FLAGE_TEST_ERROR", "REQUIRED"},
	}
	if len(capture.UsagesAsEnviron("REQUIRED")) != len(expected) {
		t.Errorf("%d != %d", len(capture.UsagesAsEnviron("REQUIRED")), len(expected))
	}
	for i, v := range capture.UsagesAsEnviron("REQUIRED") {
		if v != expected[i] {
			t.Errorf("capture.UsagesAsEnviron(REQUIRED) != expected")
		}
	}

	// Test Keys() method
	keys := capture.Keys()
	if keys != nil {
		t.Errorf("Expected Keys() to return nil, got %v", keys)
	}
}

func TestEnvMapMethods(t *testing.T) {
	t.Run("Set method", func(t *testing.T) {
		em := make(EnvMap)
		err := em.Set("KEY1=value1")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(em["KEY1"]) != 1 || em["KEY1"][0] != "value1" {
			t.Errorf("Expected KEY1=value1, got %v", em["KEY1"])
		}

		// Set without value
		err = em.Set("KEY2")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(em["KEY2"]) != 1 || em["KEY2"][0] != "" {
			t.Errorf("Expected KEY2='', got %v", em["KEY2"])
		}

		// Set with equals in value
		err = em.Set("KEY3=value=with=equals")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(em["KEY3"]) != 1 || em["KEY3"][0] != "value=with=equals" {
			t.Errorf("Expected KEY3=value=with=equals, got %v", em["KEY3"])
		}

		// Set multiple values for same key
		err = em.Set("KEY1=value2")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(em["KEY1"]) != 2 {
			t.Errorf("Expected 2 values for KEY1, got %d", len(em["KEY1"]))
		}
	})

	t.Run("String method", func(t *testing.T) {
		em := EnvMap{
			"KEY1": {"value1"},
			"KEY2": {"value2"},
		}
		str := em.String()
		// The order might vary, so we just check both keys are present
		if !containsSubstring(str, "KEY1") || !containsSubstring(str, "KEY2") {
			t.Errorf("String() output missing expected keys: %s", str)
		}
		if !containsSubstring(str, "value1") || !containsSubstring(str, "value2") {
			t.Errorf("String() output missing expected values: %s", str)
		}
	})

	t.Run("Reset method", func(t *testing.T) {
		em := EnvMap{
			"KEY1": {"value1"},
			"KEY2": {"value2"},
		}
		em.Reset()
		if len(em) != 0 {
			t.Errorf("Expected empty map after Reset(), got %v", em)
		}
	})

	t.Run("Keys method", func(t *testing.T) {
		em := EnvMap{
			"KEY1": {"value1"},
			"KEY2": {"value2"},
			"KEY3": {"value3"},
		}
		keys := em.Keys()
		if len(keys) != 3 {
			t.Errorf("Expected 3 keys, got %d", len(keys))
		}
		// Keys should be sorted
		if keys[0] != "KEY1" || keys[1] != "KEY2" || keys[2] != "KEY3" {
			t.Errorf("Expected sorted keys [KEY1, KEY2, KEY3], got %v", keys)
		}
	})
}

func TestEnvFile(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("successful read", func(t *testing.T) {
		testFile := tmpDir + "/test-env.txt"
		content := "KEY1=value1\nKEY2=value2"
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		env, err := EnvFile(nil, testFile)
		if err != nil {
			t.Errorf("EnvFile() error = %v", err)
			return
		}

		if env.Get("KEY1") != "value1" {
			t.Errorf("Expected KEY1=value1, got %s", env.Get("KEY1"))
		}
		if env.Get("KEY2") != "value2" {
			t.Errorf("Expected KEY2=value2, got %s", env.Get("KEY2"))
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := EnvFile(nil, tmpDir+"/nonexistent.txt")
		if err == nil {
			t.Error("EnvFile() expected error for nonexistent file")
		}
	})

	t.Run("with parent", func(t *testing.T) {
		testFile := tmpDir + "/test-env2.txt"
		content := "CHILD_KEY=child_value"
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		parent := NewEnv(nil, EnvMap{
			"PARENT_KEY": {"parent_value"},
		})

		env, err := EnvFile(parent, testFile)
		if err != nil {
			t.Errorf("EnvFile() error = %v", err)
			return
		}

		if env.Get("CHILD_KEY") != "child_value" {
			t.Errorf("Expected CHILD_KEY=child_value, got %s", env.Get("CHILD_KEY"))
		}
		if env.Get("PARENT_KEY") != "parent_value" {
			t.Errorf("Expected PARENT_KEY=parent_value, got %s", env.Get("PARENT_KEY"))
		}
	})
}

func TestEnvKeys(t *testing.T) {
	env := NewEnv(nil, EnvMap{
		"KEY1": {"value1"},
		"KEY2": {"value2"},
	})

	keys := env.Keys()
	if len(keys) != 2 {
		t.Errorf("Expected 2 keys, got %d", len(keys))
	}
}

func TestEnvMap(t *testing.T) {
	env := NewEnv(nil, EnvMap{
		"KEY1": {"value1"},
		"KEY2": {"value2a", "value2b"},
	})

	m := env.Map()
	if len(m) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(m))
	}
	if len(m["KEY2"]) != 2 {
		t.Errorf("Expected KEY2 to have 2 values, got %d", len(m["KEY2"]))
	}
}

func TestEnvSlice(t *testing.T) {
	env := NewEnv(nil, EnvMap{
		"KEY1": {"value1"},
		"KEY2": {"value2a", "value2b"},
	})

	slice := env.Slice()
	if len(slice) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(slice))
	}
}

func TestEnvLookup(t *testing.T) {
	env := NewEnv(nil, EnvMap{
		"KEY1": {"value1"},
	})

	value, ok := env.Lookup("KEY1")
	if !ok {
		t.Error("Expected Lookup to find KEY1")
	}
	if value != "value1" {
		t.Errorf("Expected value1, got %s", value)
	}

	_, ok = env.Lookup("NONEXISTENT")
	if ok {
		t.Error("Expected Lookup to not find NONEXISTENT")
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
