package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Knetic/govaluate"
	"github.com/Masterminds/semver/v3"
	"github.com/goreleaser/nfpm"
	_ "github.com/goreleaser/nfpm/deb"
	_ "github.com/goreleaser/nfpm/rpm"
	"github.com/pkg/errors"
)

type (
	Arch    string
	Format  string
	Version struct {
		*semver.Version
	}
)

const (
	X86_64  Arch = "x86_64"
	AArch64 Arch = "aarch64"
	ArmV7   Arch = "armv7"

	DEB Format = "deb"
	RPM Format = "rpm"
)

var (
	GithubAction = os.Getenv("GITHUB_ACTIONS") != ""
	BuildVersion = "0.0.0"
)

type Package struct {
	Format  Format
	Name    string
	Version Version
	Arch    Arch
	Meta    Meta
	Files   map[string]File
	Units   []string
	Cond    []Cond
	User    string
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

type Unit struct {
	File string `yaml:"file"`
}

type Cond struct {
	When  string   `yaml:"when"`
	Units []string `yaml:"units"`
}

func main() {
	args, err := ParseArgs()
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
	if err := args.Validate(); err != nil {
		return err
	}

	for _, pkg := range args.Packages() {
		info, err := pkg.Info()
		if err != nil {
			return errors.WithStack(err)
		}

		log.Printf("building %s", info.Target)

		if err := nfpm.Validate(info); err != nil {
			return errors.WithStack(err)
		}

		pkg, err := pkg.Packager()
		if err != nil {
			return errors.WithStack(err)
		}

		f, err := os.Create(info.Target)
		if err != nil {
			return errors.WithStack(err)
		}
		defer f.Close()

		if err := pkg.Package(info, f); err != nil {
			return errors.WithStack(err)
		}

		if GithubAction {
			fmt.Println("::set-output name=package::", info.Target)
		}
	}

	return nil
}

func (p *Package) Filename() string {
	xlated := p.Format.Translate(p.Arch)
	switch p.Format {
	case DEB:
		return fmt.Sprintf("%s_%s-1_%s.deb", p.Name, p.Version.String(), xlated)
	case RPM:
		return fmt.Sprintf("%s-%s-1.%s.rpm", p.Name, p.Version.String(), xlated)
	}
	return ""

}

func (p *Package) Info() (*nfpm.Info, error) {
	files := map[string]string{}
	confs := map[string]string{}
	units := append([]string(nil), p.Units...)

	for _, cond := range p.Cond {
		expr, err := govaluate.NewEvaluableExpression(cond.When)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		parameters := map[string]interface{}{
			"arch":    string(p.Arch),
			"version": p.Version.String(),
			"format":  string(p.Format),
		}

		result, err := expr.Evaluate(parameters)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		if r, ok := result.(bool); ok && r {
			for _, unit := range cond.Units {
				units = append(units, unit)
			}
		}
	}

	for path, info := range p.Files {
		if info.User == "" {
			info.User = "root"
		}

		if info.Mode == "" {
			info.Mode = "0644"
		}

		spec := fmt.Sprintf("%s:%s:%s", path, info.User, info.Mode)

		if !info.Keep {
			files[info.File] = spec
		} else {
			confs[info.File] = spec
		}
	}

	return nfpm.WithDefaults(&nfpm.Info{
		Name:        p.Name,
		Arch:        p.Format.Translate(p.Arch),
		Platform:    "",
		Version:     p.Version.String(),
		Description: p.Meta.Description,
		Vendor:      p.Meta.Vendor,
		Maintainer:  p.Meta.Maintainer,
		License:     p.Meta.License,
		Overridables: nfpm.Overridables{
			Files:        files,
			ConfigFiles:  confs,
			SystemdUnits: units,
			User:         p.User,
		},
		Target: p.Filename(),
	}), nil
}

func (p *Package) Packager() (nfpm.Packager, error) {
	return nfpm.Get(string(p.Format))
}

func (f Format) Translate(arch Arch) string {
	xlate := map[Format]map[Arch]string{
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
	return xlate[f][arch]
}

func (v *Version) String() string {
	return v.Version.String()
}
