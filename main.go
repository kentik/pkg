package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/goreleaser/nfpm"
	_ "github.com/goreleaser/nfpm/deb"
	_ "github.com/goreleaser/nfpm/rpm"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type (
	Arch    string
	Package string
	Version struct {
		*semver.Version
	}
)

const (
	X86_64  Arch = "x86_64"
	AArch64 Arch = "aarch64"
	ArmV7   Arch = "armv7"

	DEB Package = "deb"
	RPM Package = "rpm"
)

var (
	GithubAction = os.Getenv("GITHUB_ACTIONS") != ""
	BuildVersion = "0.0.0"
)

type Args struct {
	Name    string    `long:"name"    description:"package name"    required:"true"`
	Version Version   `long:"version" description:"package version" required:"true"`
	Arch    Arch      `long:"arch"    description:"package arch"                   `
	Deb     bool      `long:"deb"     description:"build a .deb package"           `
	RPM     bool      `long:"rpm"     description:"build a .rpm package"           `
	Format  []Package `long:"format"  description:"package format"                 `
	Args    struct {
		Config Config `positional-arg-name:"package.yaml" required:"true"`
	} `positional-args:"true"`
}

type Config struct {
	Meta  Meta            `yaml:"meta"`
	Files map[string]File `yaml:"files"`
	Units []string        `yaml:"units"`
	User  string          `yaml:"user"`
}

type Meta struct {
	Description string `yaml:"description"`
	License     string `yaml:"license"`
	Vendor      string `yaml:"vendor"     `
	Maintainer  string `yaml:"maintainer" `
}

type File struct {
	File string `yaml:"file"`
	Mode string `yaml:"mode"`
	Keep bool   `yaml:"keep"`
	User string `yaml:"user"`
}

func main() {
	args := &Args{
		Arch: X86_64,
	}

	parser := flags.NewParser(args, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}

	log.Printf("pkg %s", BuildVersion)

	if err := build(args); err != nil {
		log.Fatalf("%+v", err)
		os.Exit(1)
	}
}

func build(args *Args) error {
	files := map[string]string{}
	confs := map[string]string{}

	for path, info := range args.Args.Config.Files {
		s, err := os.Stat(info.File)
		if err != nil || !s.Mode().IsRegular() {
			return fmt.Errorf("'%s' is not a file", info.File)
		}

		if !info.Keep {
			files[info.File] = path
		} else {
			confs[info.File] = path
		}

		if mode, err := strconv.ParseInt(info.Mode, 8, 0); err != nil {
			mode := os.FileMode(mode)
			path := info.File
			if err := os.Chmod(path, mode); err != nil {
				return errors.Wrapf(err, "chmod %s", mode)
			}
		}
	}

	info := nfpm.WithDefaults(&nfpm.Info{
		Name:        args.Name,
		Arch:        "",
		Platform:    "",
		Version:     args.Version.String(),
		Description: args.Args.Config.Meta.Description,
		Vendor:      args.Args.Config.Meta.Vendor,
		Maintainer:  args.Args.Config.Meta.Maintainer,
		License:     args.Args.Config.Meta.License,
		Overridables: nfpm.Overridables{
			Files:        files,
			ConfigFiles:  confs,
			SystemdUnits: args.Args.Config.Units,
			User:         args.Args.Config.User,
		},
	})

	for pkg, file := range args.Packages() {
		log.Printf("building %s", file)

		info.Arch = pkg.Translate(args.Arch)
		info.Target = file

		if err := nfpm.Validate(info); err != nil {
			return errors.WithStack(err)
		}

		pkg, err := pkg.Packager()
		if err != nil {
			return errors.WithStack(err)
		}

		f, err := os.Create(file)
		if err != nil {
			return errors.WithStack(err)
		}
		defer f.Close()

		if err := pkg.Package(info, f); err != nil {
			return errors.WithStack(err)
		}

		if GithubAction {
			fmt.Println("::set-output name=package::", file)
		}
	}

	return nil
}

func (a *Args) Packages() map[Package]string {
	wanted := map[Package]struct{}{}

	if a.Deb {
		wanted[DEB] = struct{}{}
	}

	if a.RPM {
		wanted[RPM] = struct{}{}
	}

	for _, pkg := range a.Format {
		wanted[pkg] = struct{}{}
	}

	packages := map[Package]string{}

	for pkg, _ := range wanted {
		packages[pkg] = pkg.Filename(a.Name, a.Version, a.Arch)
	}

	return packages
}

func (p Package) Filename(name string, version Version, arch Arch) string {
	xlated := p.Translate(arch)
	switch p {
	case DEB:
		return fmt.Sprintf("%s_%s-1_%s.deb", name, version.String(), xlated)
	case RPM:
		return fmt.Sprintf("%s-%s-1.%s.rpm", name, version.String(), xlated)
	}
	return ""
}

func (p Package) Translate(arch Arch) string {
	xlate := map[Package]map[Arch]string{
		DEB: map[Arch]string{
			X86_64:  "amd64",
			AArch64: "arm64",
			ArmV7:   "armhf",
		},
		RPM: map[Arch]string{
			X86_64:  "x86_64",
			AArch64: "aarch64",
			ArmV7:   "armv7hl",
		},
	}
	return xlate[p][arch]
}

func (p Package) Packager() (nfpm.Packager, error) {
	return nfpm.Get(string(p))
}

func (v *Version) String() string {
	return v.Version.String()
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

func (p *Package) UnmarshalFlag(value string) error {
	switch strings.ToLower(value) {
	case "deb":
		*p = DEB
	case "rpm":
		*p = RPM
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
