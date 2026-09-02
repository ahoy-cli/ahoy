package main

import "errors"

// EnvFile is a single entry in an `env:` list: the path to a file of
// KEY=VALUE lines, and whether Ahoy should stay quiet when it is absent.
//
// A file is required by default. Ahoy still runs when a required file is
// missing, because that has always been the behaviour and changing it would
// break existing configurations, but it now says so on stderr. Marking an
// entry optional is how you say the absence is deliberate, which is the same
// meaning `optional` already carries for `imports`.
type EnvFile struct {
	Path     string
	Optional bool
}

// EnvFiles allows `env:` to be written as a single file, a list of files, or
// a list mixing plain paths with entries that mark a file optional:
//
//	env: .env
//
//	env:
//	  - .env.base
//	  - .env.local
//
//	env:
//	  - .env
//	  - path: .env.local
//	    optional: true
type EnvFiles []EnvFile

// errEnvFileNoPath is returned for a mapping entry that never names a file,
// e.g. `- optional: true` on its own.
var errEnvFileNoPath = errors.New("env entry must be a file path, or a mapping with a 'path' key (e.g. `- path: .env.local` with `optional: true`)")

// UnmarshalYAML accepts either a bare path or a mapping with `path` and
// `optional` keys, so both forms can appear in the same list.
func (f *EnvFile) UnmarshalYAML(unmarshal func(any) error) error {
	var path string
	if err := unmarshal(&path); err == nil {
		f.Path = path
		f.Optional = false
		return nil
	}

	// A named type avoids recursing into this method.
	var entry struct {
		Path     string `yaml:"path"`
		Optional bool   `yaml:"optional"`
	}
	if err := unmarshal(&entry); err != nil {
		return err
	}
	if entry.Path == "" {
		return errEnvFileNoPath
	}

	f.Path = entry.Path
	f.Optional = entry.Optional
	return nil
}

// UnmarshalYAML accepts a list of entries or a single entry, preserving the
// single-value shorthand that `env:` has always allowed.
func (e *EnvFiles) UnmarshalYAML(unmarshal func(any) error) error {
	var multi []EnvFile
	if err := unmarshal(&multi); err == nil {
		*e = multi
		return nil
	}

	var single EnvFile
	if err := unmarshal(&single); err != nil {
		return err
	}
	*e = EnvFiles{single}
	return nil
}
