package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type Args struct {
	Name    string   `long:"name"    description:"package name"    required:"true"`
	Version string   `long:"version" description:"package version" required:"true"`
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
	Meta    Meta             `yaml:"meta"`
	Files   map[string]File  `yaml:"files"`
	Dirs    []string         `yaml:"dirs"`
	Units   []string         `yaml:"units"`
	Scripts map[Phase]string `yaml:"scripts"`
	Cond    []Cond           `yaml:"conditional"`
	User    string           `yaml:"user"`
}

func ParseArgs() (*Args, error) {
	args := &Args{
		Arch: X86_64,
	}

	parser := flags.NewParser(args, flags.Default)
	_, err := parser.Parse()

	return args, err
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
			Scripts: a.Inputs.Config.Scripts,
			Cond:    a.Inputs.Config.Cond,
			User:    a.Inputs.Config.User,
		})
	}

	return pkgs
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

func (p *Phase) UnmarshalYAML(un func(interface{}) error) error {
	var phase string

	if err := un(&phase); err != nil {
		return errors.WithStack(err)
	}

	switch Phase(phase) {
	case PreInstall, PostInstall, PreRemove, PostRemove:
		*p = Phase(phase)
	default:
		return fmt.Errorf("invalid phase '%s'", phase)
	}

	return nil
}
