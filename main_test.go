package main

import (
	"testing"

	"github.com/goreleaser/nfpm"
	"github.com/stretchr/testify/assert"
)

func TestPackageInfo(t *testing.T) {
	assert := assert.New(t)

	args := testArgs()
	pkg := args.Packages()[0]

	info, err := pkg.Info()
	assert.NoError(err)

	assert.Equal(&nfpm.Info{
		Overridables: nfpm.Overridables{
			Files:       map[string]string{},
			ConfigFiles: map[string]string{},
		},
		Name:        args.Name,
		Version:     args.Version.String(),
		Arch:        pkg.Format.Translate(args.Arch),
		Platform:    "linux",
		Maintainer:  args.Inputs.Config.Meta.Maintainer,
		Description: args.Inputs.Config.Meta.Description,
		Vendor:      args.Inputs.Config.Meta.Vendor,
		License:     args.Inputs.Config.Meta.License,
		Target:      "test_1.0.0-1_amd64.deb",
	}, info)
}

func TestPackageOverridables(t *testing.T) {
	assert := assert.New(t)

	args := testArgs()
	args.Inputs.Config.Files = map[string]File{
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
	args.Inputs.Config.Units = []string{
		"unit0.service",
	}

	pkg := args.Packages()[0]

	info, err := pkg.Info()
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

func TestPackageConditional(t *testing.T) {
	assert := assert.New(t)

	var (
		args = testArgs()
		deb  Package
		rpm  Package
		info *nfpm.Info
		err  error
	)

	args.Deb = true
	args.RPM = true

	args.Inputs.Config.Cond = []Cond{
		Cond{When: `format == "deb"`, Units: []string{"deb.service"}},
		Cond{When: `format == "rpm"`, Units: []string{"rpm.service"}},
	}

	for _, pkg := range args.Packages() {
		switch pkg.Format {
		case DEB:
			deb = pkg
		case RPM:
			rpm = pkg
		default:
			panic("unreachable")
		}
	}

	info, err = deb.Info()
	assert.NoError(err)

	assert.Equal(nfpm.Overridables{
		Files:       map[string]string{},
		ConfigFiles: map[string]string{},
		SystemdUnits: []string{
			"deb.service",
		},
	}, info.Overridables)

	info, err = rpm.Info()
	assert.NoError(err)

	assert.Equal(nfpm.Overridables{
		Files:       map[string]string{},
		ConfigFiles: map[string]string{},
		SystemdUnits: []string{
			"rpm.service",
		},
	}, info.Overridables)
}

func TestPackageTarget(t *testing.T) {
	assert := assert.New(t)

	var (
		args = testArgs()
		info *nfpm.Info
		err  error
	)

	args.Deb = false
	args.RPM = false

	args.Format = []Format{DEB}
	info, err = args.Packages()[0].Info()
	assert.NoError(err)
	assert.Equal("test_1.0.0-1_amd64.deb", info.Target)

	args.Format = []Format{RPM}
	info, err = args.Packages()[0].Info()
	assert.NoError(err)
	assert.Equal("test-1.0.0-1.x86_64.rpm", info.Target)
}
