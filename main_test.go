package main

import (
	"testing"

	"github.com/goreleaser/nfpm"
	"github.com/stretchr/testify/assert"
	"github.com/twpayne/go-vfs/vfst"
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
		Version:     args.Version,
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
	args.Inputs.Config.Dirs = []string{
		"dir0",
	}
	args.Inputs.Config.Units = []string{
		"unit0.service",
	}
	args.Inputs.Config.Scripts = map[Phase]string{
		PreInstall:  "script0",
		PostInstall: "script1",
		PreRemove:   "script2",
		PostRemove:  "script3",
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
		EmptyFolders: []string{
			"dir0",
		},
		SystemdUnits: []string{
			"unit0.service",
		},
		Scripts: nfpm.Scripts{
			PreInstall:  "script0",
			PostInstall: "script1",
			PreRemove:   "script2",
			PostRemove:  "script3",
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

func TestPackagePrepare(t *testing.T) {
	assert := assert.New(t)

	fs, cleanup, err := vfst.NewTestFS(map[string]interface{}{
		"bin0": &vfst.File{Perm: 0o777, Contents: []byte{}},
		"bin1": &vfst.File{Perm: 0o777, Contents: []byte{}},
		"cfg0": &vfst.File{Perm: 0o777},
	})
	if err != nil {
		t.Error(err)
	}
	defer cleanup()

	args := testArgs()
	args.Inputs.Config.Files = map[string]File{
		"bin0": File{
			File: "/bin0",
		},
		"bin1": File{
			File: "/bin1",
			Mode: "0755",
			User: "toor",
		},
	}

	pkg := args.Packages()[0]

	_, err = pkg.Info()
	assert.NoError(err)
	_, err = pkg.Prepare(fs)
	assert.NoError(err)

	vfst.RunTests(t, fs, "prepare",
		vfst.TestPath("/bin0", vfst.TestModePerm(0o644)),
		vfst.TestPath("/bin1", vfst.TestModePerm(0o755)),
	)
}
