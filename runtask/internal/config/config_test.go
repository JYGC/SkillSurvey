package config

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeConfigJSONFile(t *testing.T, fields map[string]any) string {
	t.Helper()
	b, err := json.Marshal(fields)
	if err != nil {
		t.Fatalf("marshal config fields: %v", err)
	}
	path := filepath.Join(t.TempDir(), "runtask.json")
	if err := os.WriteFile(path, b, 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}
	return path
}

func TestLoadFromFileResolvesDynamicContentExtractionTimeout(t *testing.T) {
	tests := []struct {
		name            string
		timeoutFieldSet bool
		timeoutSeconds  int
		expectedTimeout time.Duration
		expectLoadError bool
		errorNamesField string
	}{
		{
			name:            "timeout field absent",
			timeoutFieldSet: false,
			expectedTimeout: 60 * time.Second,
		},
		{
			name:            "timeout field zero",
			timeoutFieldSet: true,
			timeoutSeconds:  0,
			expectedTimeout: 60 * time.Second,
		},
		{
			name:            "timeout field positive",
			timeoutFieldSet: true,
			timeoutSeconds:  180,
			expectedTimeout: 180 * time.Second,
		},
		{
			name:            "timeout field negative",
			timeoutFieldSet: true,
			timeoutSeconds:  -1,
			expectLoadError: true,
			errorNamesField: "DynamicContentExtractionTimeoutSeconds",
		},
		{
			name:            "timeout field large enough to overflow time.Duration seconds",
			timeoutFieldSet: true,
			timeoutSeconds:  math.MaxInt64,
			expectLoadError: true,
			errorNamesField: "DynamicContentExtractionTimeoutSeconds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := map[string]any{
				"ErrorLogFile": filepath.Join(t.TempDir(), "error.log"),
			}
			if tt.timeoutFieldSet {
				fields["DynamicContentExtractionTimeoutSeconds"] = tt.timeoutSeconds
			}
			configPath := writeConfigJSONFile(t, fields)

			cfg, err := loadFromFile(configPath)

			if tt.expectLoadError {
				if err == nil {
					t.Fatal("expected loadFromFile to return an error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errorNamesField) {
					t.Fatalf("expected error to name field %q, got: %v", tt.errorNamesField, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected loadFromFile to succeed, got: %v", err)
			}
			if got := cfg.DynamicContentExtractionTimeout(); got != tt.expectedTimeout {
				t.Errorf("expected timeout %v, got %v", tt.expectedTimeout, got)
			}
		})
	}
}

func TestZeroValueConfigResolvesToDefaultTimeout(t *testing.T) {
	var cfg Config
	if got := cfg.DynamicContentExtractionTimeout(); got != 60*time.Second {
		t.Errorf("expected zero-value Config to resolve to 60s, got %v", got)
	}
}

// TestConfigWithOverflowingTimeoutSecondsResolvesToDefaultTimeout guards a Config
// built directly as a struct literal (bypassing validate(), exactly like the five
// literals documented in design.md) against the same overflow that
// TestLoadFromFileResolvesDynamicContentExtractionTimeout proves is rejected on
// the file-loading path. time.Duration is int64 nanoseconds, so a large enough
// positive seconds value wraps around to a non-positive duration when multiplied
// by time.Second — which is exactly the "already expired" failure mode
// requirements.md rejects negative values to avoid.
func TestConfigWithOverflowingTimeoutSecondsResolvesToDefaultTimeout(t *testing.T) {
	cfg := Config{DynamicContentExtractionTimeoutSeconds: math.MaxInt64}
	if got := cfg.DynamicContentExtractionTimeout(); got != 60*time.Second {
		t.Errorf("expected an overflowing timeout to resolve to the 60s default, got %v", got)
	}
}

func TestConfigValidate(t *testing.T) {
	validErrorLogFile := filepath.Join(t.TempDir(), "error.log")

	tests := []struct {
		name             string
		cfg              Config
		expectError      bool
		errorNamesFields []string
	}{
		{
			name: "well-formed config",
			cfg: Config{
				ErrorLogFile:        validErrorLogFile,
				PocketBaseUrl:       "http://127.0.0.1:8090",
				SmtpPort:            587,
				ServiceAccountEmail: "svc@example.com",
				SenderEmail:         "sender@example.com",
				EmailRecipient:      "recipient@example.com",
			},
			expectError: false,
		},
		{
			name:             "empty ErrorLogFile",
			cfg:              Config{ErrorLogFile: ""},
			expectError:      true,
			errorNamesFields: []string{"ErrorLogFile"},
		},
		{
			name: "PocketBaseUrl without scheme",
			cfg: Config{
				ErrorLogFile:  validErrorLogFile,
				PocketBaseUrl: "127.0.0.1:8090",
			},
			expectError:      true,
			errorNamesFields: []string{"PocketBaseUrl"},
		},
		{
			name: "PocketBaseUrl without host",
			cfg: Config{
				ErrorLogFile:  validErrorLogFile,
				PocketBaseUrl: "http://",
			},
			expectError:      true,
			errorNamesFields: []string{"PocketBaseUrl"},
		},
		{
			name: "PocketBaseUrl with non-HTTP scheme",
			cfg: Config{
				ErrorLogFile:  validErrorLogFile,
				PocketBaseUrl: "ftp://127.0.0.1:8090",
			},
			expectError:      true,
			errorNamesFields: []string{"PocketBaseUrl"},
		},
		{
			name: "empty PocketBaseUrl produces no error",
			cfg: Config{
				ErrorLogFile:  validErrorLogFile,
				PocketBaseUrl: "",
			},
			expectError: false,
		},
		{
			name: "SmtpPort in range",
			cfg: Config{
				ErrorLogFile: validErrorLogFile,
				SmtpPort:     587,
			},
			expectError: false,
		},
		{
			name: "SmtpPort out of range",
			cfg: Config{
				ErrorLogFile: validErrorLogFile,
				SmtpPort:     70000,
			},
			expectError:      true,
			errorNamesFields: []string{"SmtpPort"},
		},
		{
			name: "SmtpPort zero produces no error",
			cfg: Config{
				ErrorLogFile: validErrorLogFile,
				SmtpPort:     0,
			},
			expectError: false,
		},
		{
			name: "malformed SenderEmail",
			cfg: Config{
				ErrorLogFile: validErrorLogFile,
				SenderEmail:  "not-an-email",
			},
			expectError:      true,
			errorNamesFields: []string{"SenderEmail"},
		},
		{
			name: "empty email fields produce no error",
			cfg: Config{
				ErrorLogFile:        validErrorLogFile,
				ServiceAccountEmail: "",
				SenderEmail:         "",
				EmailRecipient:      "",
			},
			expectError: false,
		},
		{
			name: "two simultaneous failures reported together",
			cfg: Config{
				ErrorLogFile: "",
				SmtpPort:     70000,
			},
			expectError:      true,
			errorNamesFields: []string{"ErrorLogFile", "SmtpPort"},
		},
		{
			name: "cleanfs-shaped config with only ErrorLogFile set validates successfully",
			cfg: Config{
				ErrorLogFile: validErrorLogFile,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.validate()

			if !tt.expectError {
				if err != nil {
					t.Fatalf("expected validate() to succeed, got: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("expected validate() to return an error, got nil")
			}
			for _, field := range tt.errorNamesFields {
				if !strings.Contains(err.Error(), field) {
					t.Errorf("expected error to name field %q, got: %v", field, err)
				}
			}
		})
	}
}
