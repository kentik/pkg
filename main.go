package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/Knetic/govaluate"
	"github.com/goreleaser/nfpm"
	_ "github.com/goreleaser/nfpm/deb"
	_ "github.com/goreleaser/nfpm/rpm"
	"github.com/pkg/errors"
	"github.com/twpayne/go-vfs"
)

type (
	Arch   string
	Format string
	Phase  string
)

const (
	X86_64  Arch = "x86_64"
	AArch64 Arch = "aarch64"
	ArmV7   Arch = "armv7"

	DEB Format = "deb"
	RPM Format = "rpm"

	PreInstall  Phase = "pre-install"
	PostInstall Phase = "post-install"
	PreRemove   Phase = "pre-remove"
	PostRemove  Phase = "post-remove"
)

var (
	GithubAction = os.Getenv("GITHUB_ACTIONS") != ""
	BuildVersion = "0.0.0"
)

type Package struct {
	Format  Format
	Name    string
	Version string
	Arch    Arch
	Meta    Meta
	Files   map[string]File
	Dirs    []string
	Units   []string
	Scripts map[Phase]string
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

	if err := build(args, vfs.HostOSFS); err != nil {
		log.Fatalf("%+v", err)
		os.Exit(1)
	}
}

func build(args *Args, fs vfs.FS) error {
	for _, pkg := range args.Packages() {
		info, err := pkg.Info()
		if err != nil {
			return errors.WithStack(err)
		}

		if err := nfpm.Validate(info); err != nil {
			return errors.WithStack(err)
		}

		log.Printf("building %s", info.Target)

		f, err := fs.Create(info.Target)
		if err != nil {
			return errors.WithStack(err)
		}
		defer f.Close()

		pkg, err := pkg.Prepare(fs)
		if err != nil {
			return errors.WithStack(err)
		}

		if err := pkg.Package(info, f); err != nil {
			return errors.WithStack(err)
		}

		if GithubAction {
			fmt.Printf("::set-output name=package::%s\n", info.Target)
		}
	}

	return nil
}

func (p *Package) Filename() string {
	xlated := p.Format.Translate(p.Arch)
	switch p.Format {
	case DEB:
		return fmt.Sprintf("%s_%s-1_%s.deb", p.Name, p.Version, xlated)
	case RPM:
		return fmt.Sprintf("%s-%s-1.%s.rpm", p.Name, p.Version, xlated)
	}
	return ""

}

func (p *Package) Info() (*nfpm.Info, error) {
	var (
		files   = map[string]string{}
		confs   = map[string]string{}
		dirs    = append([]string(nil), p.Dirs...)
		units   = append([]string(nil), p.Units...)
		scripts nfpm.Scripts
	)

	for _, cond := range p.Cond {
		expr, err := govaluate.NewEvaluableExpression(cond.When)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		parameters := map[string]interface{}{
			"arch":    string(p.Arch),
			"version": p.Version,
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

		p.Files[path] = info
	}

	for phase, file := range p.Scripts {
		switch phase {
		case PreInstall:
			scripts.PreInstall = file
		case PostInstall:
			scripts.PostInstall = file
		case PreRemove:
			scripts.PreRemove = file
		case PostRemove:
			scripts.PostRemove = file
		}
	}

	return nfpm.WithDefaults(&nfpm.Info{
		Name:        p.Name,
		Arch:        p.Format.Translate(p.Arch),
		Platform:    "",
		Version:     p.Version,
		Description: p.Meta.Description,
		Vendor:      p.Meta.Vendor,
		Maintainer:  p.Meta.Maintainer,
		License:     p.Meta.License,
		Overridables: nfpm.Overridables{
			Files:        files,
			ConfigFiles:  confs,
			EmptyFolders: dirs,
			Scripts:      scripts,
			SystemdUnits: units,
			User:         p.User,
		},
		Target: p.Filename(),
	}), nil
}

func (p *Package) Prepare(fs vfs.FS) (nfpm.Packager, error) {
	check := func(path string) error {
		s, err := fs.Stat(path)
		if err == nil && !s.Mode().IsRegular() {
			return fmt.Errorf("'%s' is not a file", path)
		} else if os.IsNotExist(err) {
			return fmt.Errorf("'%s' does not exist", path)
		} else {
			return errors.WithStack(err)
		}
	}

	for _, info := range p.Files {
		if err := check(info.File); err != nil {
			return nil, err
		}

		mode, err := strconv.ParseInt(info.Mode, 8, 0)
		if err != nil {
			return nil, fmt.Errorf("invalid mode: '%s'", info.Mode)
		}

		if err := fs.Chmod(info.File, os.FileMode(mode)); err != nil {
			return nil, errors.Wrapf(err, "chmod %s", info.Mode)
		}
	}

	for _, file := range p.Scripts {
		if err := check(file); err != nil {
			return nil, err
		}
	}

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
