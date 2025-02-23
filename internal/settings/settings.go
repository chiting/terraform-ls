// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package settings

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-ls/internal/terraform/datadir"
	"github.com/mitchellh/mapstructure"
)

type ExperimentalFeatures struct {
	ValidateOnSave        bool `mapstructure:"validateOnSave"`
	PrefillRequiredFields bool `mapstructure:"prefillRequiredFields"`
}

type Indexing struct {
	IgnoreDirectoryNames []string `mapstructure:"ignoreDirectoryNames"`
	IgnorePaths          []string `mapstructure:"ignorePaths"`
}

type Terraform struct {
	Path        string `mapstructure:"path"`
	Timeout     string `mapstructure:"timeout"`
	LogFilePath string `mapstructure:"logFilePath"`
}

type Options struct {
	CommandPrefix string   `mapstructure:"commandPrefix"`
	Indexing      Indexing `mapstructure:"indexing"`

	// ExperimentalFeatures encapsulates experimental features users can opt into.
	ExperimentalFeatures ExperimentalFeatures `mapstructure:"experimentalFeatures"`

	IgnoreSingleFileWarning bool `mapstructure:"ignoreSingleFileWarning"`

	Terraform Terraform `mapstructure:"terraform"`

	XLegacyModulePaths              []string `mapstructure:"rootModulePaths"`
	XLegacyExcludeModulePaths       []string `mapstructure:"excludeModulePaths"`
	XLegacyIgnoreDirectoryNames     []string `mapstructure:"ignoreDirectoryNames"`
	XLegacyTerraformExecPath        string   `mapstructure:"terraformExecPath"`
	XLegacyTerraformExecTimeout     string   `mapstructure:"terraformExecTimeout"`
	XLegacyTerraformExecLogFilePath string   `mapstructure:"terraformExecLogFilePath"`
}

func (o *Options) Validate() error {
	if o.Terraform.Path != "" {
		path := o.Terraform.Path
		if !filepath.IsAbs(path) {
			return fmt.Errorf("Expected absolute path for Terraform binary, got %q", path)
		}
		stat, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("Unable to find Terraform binary: %s", err)
		}
		if stat.IsDir() {
			return fmt.Errorf("Expected a Terraform binary, got a directory: %q", path)
		}
	}

	if len(o.Indexing.IgnoreDirectoryNames) > 0 {
		for _, directory := range o.Indexing.IgnoreDirectoryNames {
			if directory == datadir.DataDirName {
				return fmt.Errorf("cannot ignore directory %q", datadir.DataDirName)
			}

			if strings.Contains(directory, string(filepath.Separator)) {
				return fmt.Errorf("expected directory name, got a path: %q", directory)
			}
		}
	}

	return nil
}

type DecodedOptions struct {
	Options    *Options
	UnusedKeys []string
}

func DecodeOptions(input interface{}) (*DecodedOptions, error) {
	var md mapstructure.Metadata
	var options Options

	config := &mapstructure.DecoderConfig{
		Metadata: &md,
		Result:   &options,
	}
	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		panic(err)
	}

	if err := decoder.Decode(input); err != nil {
		return nil, err
	}

	return &DecodedOptions{
		Options:    &options,
		UnusedKeys: md.Unused,
	}, nil
}
