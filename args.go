package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Args struct {
	Name    string   `long:"name"    description:"package name"    required:"true"`
	Version Version  `long:"version" description:"package version" required:"true"`
	Arch    Arch     `long:"arch"    description:"package arch"                   `
	Deb     bool     `long:"deb"     description:"build a .deb package"           `
	RPM     bool     `long:"rpm"     description:"build a .rpm package"           `
	Format  []Format `long:"format"  description:"package format"                 `
	Inputs  Inputs   `positional-args:"true"`
}

type Inputs struct {
	Config Config `positional-arg-name:"package.yaml" required:"true"`
}

type Config struct {
	Meta  Meta            `yaml:"meta"`
	Files map[string]File `yaml:"files"`
	Dirs  []string        `yaml:"dirs"`
	Units []string        `yaml:"units"`
	Cond  []Cond          `yaml:"conditional"`
	User  string          `yaml:"user"`
}

func ParseArgs() (*Args, error) {
	args := &Args{
		Arch: X86_64,
	}

	parser := flags.NewParser(args, flags.Default)
	_, err := parser.Parse()

	return args, err
}

func (a *Args) Validate() error {
	for _, info := range a.Inputs.Config.Files {
		s, err := os.Stat(info.File)
		if err != nil || !s.Mode().IsRegular() {
			return fmt.Errorf("'%s' is not a file", info.File)
		}

		if mode, err := strconv.ParseInt(info.Mode, 8, 0); err != nil {
			mode := os.FileMode(mode)
			path := info.File
			if err := os.Chmod(path, mode); err != nil {
				return errors.Wrapf(err, "chmod %s", mode)
			}
		}
	}
	return nil
}

func (a *Args) Packages() []Package {
	wanted := map[Format]struct{}{}

	if a.Deb {
		wanted[DEB] = struct{}{}
	}

	if a.RPM {
		wanted[RPM] = struct{}{}
	}

	for _, pkg := range a.Format {
		wanted[pkg] = struct{}{}
	}

	pkgs := []Package{}

	for fmt := range wanted {
		pkgs = append(pkgs, Package{
			Format:  fmt,
			Name:    a.Name,
			Version: a.Version,
			Arch:    a.Arch,
			Meta:    a.Inputs.Config.Meta,
			Files:   a.Inputs.Config.Files,
			Dirs:    a.Inputs.Config.Dirs,
			Units:   a.Inputs.Config.Units,
			Cond:    a.Inputs.Config.Cond,
			User:    a.Inputs.Config.User,
		})
	}

	return pkgs
}

func (v *Version) UnmarshalFlag(value string) error {
	version, err := semver.NewVersion(value)
	if err != nil {
		return err
	}

	*v = Version{version}

	return nil
}

func (a *Arch) UnmarshalFlag(value string) error {
	switch strings.ToLower(value) {
	case "x86_64", "amd64":
		*a = X86_64
	case "aarch64", "arm64":
		*a = AArch64
	case "armv7":
		*a = ArmV7
	default:
		return fmt.Errorf("unsupported architecture: %s", value)
	}
	return nil
}

func (f *Format) UnmarshalFlag(value string) error {
	switch strings.ToLower(value) {
	case "deb":
		*f = DEB
	case "rpm":
		*f = RPM
	default:
		return fmt.Errorf("unsupported format: %s", value)
	}
	return nil
}

func (c *Config) UnmarshalFlag(value string) error {
	f, err := os.Open(value)
	if err != nil {
		return errors.WithStack(err)
	}

	defer f.Close()

	dec := yaml.NewDecoder(f)
	cfg := &Config{}

	if err := dec.Decode(cfg); err != nil {
		return errors.WithStack(err)
	}

	*c = *cfg

	return nil
}
