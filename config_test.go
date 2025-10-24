package flage

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseConfigFile(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
		wantErr  bool
	}{
		{
			name:     "simple flags",
			input:    "-load ./file.txt -secret mysecret",
			expected: []string{"-load", "./file.txt", "-secret", "mysecret"},
			wantErr:  false,
		},
		{
			name:     "with comment at start",
			input:    "# This is a comment\n-load ./file.txt",
			expected: []string{"-load", "./file.txt"},
			wantErr:  false,
		},
		{
			name:     "with comment with leading whitespace",
			input:    "   # This is a comment\n-load ./file.txt",
			expected: []string{"-load", "./file.txt"},
			wantErr:  false,
		},
		{
			name:     "multiline flags",
			input:    "-load ./file.txt\n-secret mysecret",
			expected: []string{"-load", "./file.txt", "-secret", "mysecret"},
			wantErr:  false,
		},
		{
			name:     "quoted strings",
			input:    `-load "file with spaces.txt"`,
			expected: []string{"-load", "file with spaces.txt"},
			wantErr:  false,
		},
		{
			name:     "multiple comments",
			input:    "# Comment 1\n-flag1 value1\n# Comment 2\n-flag2 value2",
			expected: []string{"-flag1", "value1", "-flag2", "value2"},
			wantErr:  false,
		},
		{
			name:     "empty lines",
			input:    "-flag1 value1\n\n-flag2 value2",
			expected: []string{"-flag1", "value1", "-flag2", "value2"},
			wantErr:  false,
		},
		{
			name:     "comment not at start of line",
			input:    "-flag1 value1 # this is not a comment",
			expected: []string{"-flag1", "value1"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConfigFile(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseConfigFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ParseConfigFile() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestReadConfigFile(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	t.Run("successful read", func(t *testing.T) {
		// Create a test config file
		testFile := filepath.Join(tmpDir, "test-config.txt")
		content := "# Test config\n-flag1 value1\n-flag2 value2"
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		got, err := ReadConfigFile(testFile)
		if err != nil {
			t.Errorf("ReadConfigFile() error = %v", err)
			return
		}

		expected := []string{"-flag1", "value1", "-flag2", "value2"}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("ReadConfigFile() = %v, want %v", got, expected)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := ReadConfigFile(filepath.Join(tmpDir, "nonexistent.txt"))
		if err == nil {
			t.Error("ReadConfigFile() expected error for nonexistent file")
		}
	})
}

func TestParseEnvironFile(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected [][2]string
		wantErr  bool
	}{
		{
			name:  "simple key-value pairs",
			input: "KEY1=value1\nKEY2=value2",
			expected: [][2]string{
				{"KEY1", "value1"},
				{"KEY2", "value2"},
			},
			wantErr: false,
		},
		{
			name:  "with comments",
			input: "# This is a comment\nKEY1=value1\n# Another comment\nKEY2=value2",
			expected: [][2]string{
				{"KEY1", "value1"},
				{"KEY2", "value2"},
			},
			wantErr: false,
		},
		{
			name:  "with empty lines",
			input: "KEY1=value1\n\nKEY2=value2",
			expected: [][2]string{
				{"KEY1", "value1"},
				{"KEY2", "value2"},
			},
			wantErr: false,
		},
		{
			name:  "values with equals signs",
			input: "KEY1=value=with=equals",
			expected: [][2]string{
				{"KEY1", "value=with=equals"},
			},
			wantErr: false,
		},
		{
			name:     "lines without equals",
			input:    "KEY1=value1\nINVALIDLINE\nKEY2=value2",
			expected: [][2]string{
				{"KEY1", "value1"},
				{"KEY2", "value2"},
			},
			wantErr: false,
		},
		{
			name:     "empty input",
			input:    "",
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "only comments",
			input:    "# Comment 1\n# Comment 2",
			expected: nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEnvironFile([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEnvironFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ParseEnvironFile() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestReadEnvironFile(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	t.Run("successful read", func(t *testing.T) {
		// Create a test environ file
		testFile := filepath.Join(tmpDir, "test-env.txt")
		content := "# Test environ\nKEY1=value1\nKEY2=value2"
		if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		got, err := ReadEnvironFile(testFile)
		if err != nil {
			t.Errorf("ReadEnvironFile() error = %v", err)
			return
		}

		expected := [][2]string{
			{"KEY1", "value1"},
			{"KEY2", "value2"},
		}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("ReadEnvironFile() = %v, want %v", got, expected)
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := ReadEnvironFile(filepath.Join(tmpDir, "nonexistent.txt"))
		if err == nil {
			t.Error("ReadEnvironFile() expected error for nonexistent file")
		}
	})
}
