package main

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/goreleaser/nfpm"
	"github.com/stretchr/testify/assert"
)

func testArgs() *Args {
	return &Args{
		Name:    "test",
		Version: Version{semver.MustParse("1.0.0")},
		Arch:    X86_64,
		Deb:     true,
		RPM:     true,
		Inputs: Inputs{
			Package: Config{
				Meta: Meta{
					Description: "foo",
					License:     "bar",
					Vendor:      "baz",
					Maintainer:  "quux",
				},
			},
		},
	}
}

func TestGenerateInfo(t *testing.T) {
	assert := assert.New(t)

	args := testArgs()
	info, err := args.Info(DEB)
	assert.NoError(err)

	assert.Equal(&nfpm.Info{
		Overridables: nfpm.Overridables{
			Files:       map[string]string{},
			ConfigFiles: map[string]string{},
		},
		Name:        args.Name,
		Version:     args.Version.String(),
		Arch:        "",
		Platform:    "linux",
		Maintainer:  args.Inputs.Package.Meta.Maintainer,
		Description: args.Inputs.Package.Meta.Description,
		Vendor:      args.Inputs.Package.Meta.Vendor,
		License:     args.Inputs.Package.Meta.License,
	}, info)
}

func TestGenerateOverridables(t *testing.T) {
	assert := assert.New(t)

	args := testArgs()
	args.Inputs.Package.Files = map[string]File{
		"bin0": File{
			File: "bin0",
		},
		"bin1": File{
			File: "bin1",
			Mode: "0755",
			User: "toor",
		},
		"cfg0": File{
			File: "cfg0",
			Keep: true,
		},
	}
	args.Inputs.Package.Units = []string{
		"unit0.service",
	}

	info, err := args.Info(DEB)
	assert.NoError(err)

	assert.Equal(nfpm.Overridables{
		Files: map[string]string{
			"bin0": "bin0:root:0644",
			"bin1": "bin1:toor:0755",
		},
		ConfigFiles: map[string]string{
			"cfg0": "cfg0:root:0644",
		},
		SystemdUnits: []string{
			"unit0.service",
		},
	}, info.Overridables)
}

func TestConfigConditional(t *testing.T) {
	assert := assert.New(t)

	var (
		args = testArgs()
		info *nfpm.Info
		err  error
	)

	args.Inputs.Package.Cond = []Cond{
		Cond{When: `format == "deb"`, Units: []string{"deb.service"}},
		Cond{When: `format == "rpm"`, Units: []string{"rpm.service"}},
	}

	info, err = args.Info(DEB)
	assert.NoError(err)

	assert.Equal(nfpm.Overridables{
		Files:       map[string]string{},
		ConfigFiles: map[string]string{},
		SystemdUnits: []string{
			"deb.service",
		},
	}, info.Overridables)

	info, err = args.Info(RPM)
	assert.NoError(err)

	assert.Equal(nfpm.Overridables{
		Files:       map[string]string{},
		ConfigFiles: map[string]string{},
		SystemdUnits: []string{
			"rpm.service",
		},
	}, info.Overridables)
}

func TestUnmarshalArch(t *testing.T) {
	assert := assert.New(t)
	expect := map[string]Arch{
		"x86_64":  X86_64,
		"amd64":   X86_64,
		"aarch64": AArch64,
		"arm64":   AArch64,
		"armv7":   ArmV7,
		"AMD64":   X86_64,
	}

	for value, arch := range expect {
		tmp := Arch("")
		err := tmp.UnmarshalFlag(value)
		assert.NoError(err)
		assert.Equal(arch, tmp)
	}
}

func TestUnmarshalPackage(t *testing.T) {
	assert := assert.New(t)
	expect := map[string]Package{
		"deb": DEB,
		"rpm": RPM,
		"DEB": DEB,
	}

	for value, pkg := range expect {
		tmp := Package("")
		err := tmp.UnmarshalFlag(value)
		assert.NoError(err)
		assert.Equal(pkg, tmp)
	}
}
