package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/mail"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const defaultDynamicContentExtractionTimeoutSeconds = 60

// maxDynamicContentExtractionTimeoutSeconds is the largest seconds value that
// converts to a positive time.Duration without overflowing its int64
// nanosecond representation. Anything beyond it wraps around to a non-positive
// duration — the same "already expired" failure mode requirements.md rejects
// negative values to avoid.
const maxDynamicContentExtractionTimeoutSeconds int64 = math.MaxInt64 / int64(time.Second)

type Config struct {
	PocketBaseUrl                          string
	ServiceAccountEmail                    string
	ServiceAccountPassword                 string
	SeekConfigFile                         string
	JoraConfigFile                         string
	ErrorLogFile                           string
	SmtpDomain                             string
	SmtpPort                               int
	SenderEmail                            string
	SenderEmailPassword                    string
	EmailRecipient                         string
	DynamicContentExtractionTimeoutSeconds int
}

// DynamicContentExtractionTimeout resolves the configured dynamic content
// extraction timeout, treating an absent or explicit-zero value the same way
// since decoding JSON into a plain int field cannot distinguish them.
func (c Config) DynamicContentExtractionTimeout() time.Duration {
	if c.DynamicContentExtractionTimeoutSeconds <= 0 {
		return defaultDynamicContentExtractionTimeoutSeconds * time.Second
	}
	timeout := time.Duration(c.DynamicContentExtractionTimeoutSeconds) * time.Second
	if timeout <= 0 {
		return defaultDynamicContentExtractionTimeoutSeconds * time.Second
	}
	return timeout
}

// Load reads runtask.json from the directory containing the executable.
func Load() (Config, error) {
	exe, err := os.Executable()
	if err != nil {
		return Config{}, fmt.Errorf("resolve executable path: %w", err)
	}
	configPath := filepath.Join(filepath.Dir(exe), "runtask.json")
	return loadFromFile(configPath)
}

func loadFromFile(configFilePath string) (Config, error) {
	f, err := os.Open(configFilePath)
	if err != nil {
		return Config{}, fmt.Errorf("open %s: %w", configFilePath, err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// validate checks every field of Config for which an invariant is stated in
// requirements.md, reporting every failing field in a single error rather than
// only the first — runtask.json is hand-edited over SSH, so a one-error-per-run
// cycle is needlessly slow to work through.
func (c Config) validate() error {
	var validationErrors []error

	if c.ErrorLogFile == "" {
		validationErrors = append(validationErrors, errors.New("ErrorLogFile must not be empty"))
	}

	if c.PocketBaseUrl != "" {
		if err := validateHttpUrl("PocketBaseUrl", c.PocketBaseUrl); err != nil {
			validationErrors = append(validationErrors, err)
		}
	}

	if c.SmtpPort != 0 && (c.SmtpPort < 1 || c.SmtpPort > 65535) {
		validationErrors = append(validationErrors, fmt.Errorf("SmtpPort must be between 1 and 65535, got %d", c.SmtpPort))
	}

	emailFields := []struct {
		name  string
		value string
	}{
		{"ServiceAccountEmail", c.ServiceAccountEmail},
		{"SenderEmail", c.SenderEmail},
		{"EmailRecipient", c.EmailRecipient},
	}
	for _, field := range emailFields {
		if field.value == "" {
			continue
		}
		if _, err := mail.ParseAddress(field.value); err != nil {
			validationErrors = append(validationErrors, fmt.Errorf("%s must be a valid email address, got %q: %w", field.name, field.value, err))
		}
	}

	if c.DynamicContentExtractionTimeoutSeconds < 0 {
		validationErrors = append(validationErrors, fmt.Errorf("DynamicContentExtractionTimeoutSeconds must not be negative, got %d", c.DynamicContentExtractionTimeoutSeconds))
	}
	if int64(c.DynamicContentExtractionTimeoutSeconds) > maxDynamicContentExtractionTimeoutSeconds {
		validationErrors = append(validationErrors, fmt.Errorf("DynamicContentExtractionTimeoutSeconds must not exceed %d (larger values overflow the timeout duration), got %d", maxDynamicContentExtractionTimeoutSeconds, c.DynamicContentExtractionTimeoutSeconds))
	}

	return errors.Join(validationErrors...)
}

func validateHttpUrl(fieldName, value string) error {
	parsed, err := url.Parse(value)
	if err != nil {
		return fmt.Errorf("%s must be a valid URL, got %q: %w", fieldName, value, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("%s must have an http or https scheme, got %q", fieldName, value)
	}
	if parsed.Host == "" {
		return fmt.Errorf("%s must have a non-empty host, got %q", fieldName, value)
	}
	return nil
}
