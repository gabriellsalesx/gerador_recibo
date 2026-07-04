package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func Load(path string) (Config, error) {
	if path == "" {
		path = DefaultPath()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	cfg := Default()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("configuração inválida: %w", err)
	}
	applyDefaults(&cfg)
	return cfg, nil
}

func LoadOrDefault(path string) (Config, string, bool, error) {
	if path == "" {
		path = DefaultPath()
	}
	cfg, err := Load(path)
	if err == nil {
		return cfg, path, true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return Default(), path, false, nil
	}
	return Config{}, path, false, err
}

func Save(path string, cfg Config) error {
	if path == "" {
		path = DefaultPath()
	}
	applyDefaults(&cfg)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), "config-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func applyDefaults(cfg *Config) {
	def := Default()
	if cfg.Version == 0 {
		cfg.Version = def.Version
	}
	if cfg.Output.Directory == "" {
		cfg.Output.Directory = def.Output.Directory
	}
	if cfg.Output.DateLayout == "" {
		cfg.Output.DateLayout = def.Output.DateLayout
	}
	if cfg.PDF.Template == "" {
		cfg.PDF.Template = def.PDF.Template
	}
	if cfg.PDF.AccentColor == "" {
		cfg.PDF.AccentColor = def.PDF.AccentColor
	}
	if cfg.Numbering.Padding == 0 {
		cfg.Numbering.Padding = def.Numbering.Padding
	}
	if cfg.Numbering.NextNumber == 0 {
		cfg.Numbering.NextNumber = def.Numbering.NextNumber
	}
	if cfg.Defaults.Currency == "" {
		cfg.Defaults.Currency = def.Defaults.Currency
	}
	if cfg.Defaults.PaymentMethod == "" {
		cfg.Defaults.PaymentMethod = def.Defaults.PaymentMethod
	}
	if cfg.Defaults.Notes == "" {
		cfg.Defaults.Notes = def.Defaults.Notes
	}
	if cfg.Pix.TxIDPrefix == "" {
		cfg.Pix.TxIDPrefix = def.Pix.TxIDPrefix
	}
	applyDocumentDefaults(cfg, def)
}

func applyDocumentDefaults(cfg *Config, def Config) {
	q := &cfg.Documents.Quote
	dq := def.Documents.Quote
	if q.Template == "" {
		q.Template = dq.Template
	}
	if q.Numbering.Padding == 0 {
		q.Numbering.Padding = dq.Numbering.Padding
	}
	if q.Numbering.NextNumber == 0 {
		q.Numbering.NextNumber = dq.Numbering.NextNumber
		q.Numbering.Enabled = dq.Numbering.Enabled
	}
	if q.Numbering.Prefix == "" {
		q.Numbering.Prefix = dq.Numbering.Prefix
	}
	if q.Defaults.Currency == "" {
		q.Defaults.Currency = dq.Defaults.Currency
	}
	if q.Defaults.ValidityDays == 0 {
		q.Defaults.ValidityDays = dq.Defaults.ValidityDays
	}
	if q.Defaults.PaymentTerms == "" {
		q.Defaults.PaymentTerms = dq.Defaults.PaymentTerms
	}

	c := &cfg.Documents.Contract
	dc := def.Documents.Contract
	if c.Template == "" {
		c.Template = dc.Template
	}
	if c.Numbering.Padding == 0 {
		c.Numbering.Padding = dc.Numbering.Padding
	}
	if c.Numbering.NextNumber == 0 {
		c.Numbering.NextNumber = dc.Numbering.NextNumber
		c.Numbering.Enabled = dc.Numbering.Enabled
	}
	if c.Numbering.Prefix == "" {
		c.Numbering.Prefix = dc.Numbering.Prefix
	}
	if c.Defaults.Place == "" {
		c.Defaults.Place = dc.Defaults.Place
	}
	if len(c.Defaults.Clauses) == 0 {
		c.Defaults.Clauses = dc.Defaults.Clauses
	}
}
