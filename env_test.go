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
}
