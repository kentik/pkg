package main

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
)

func TestConfigDefaults(t *testing.T) {
	assert := assert.New(t)

	args := &Args{
		Name:    "test",
		Version: Version{semver.MustParse("0.0.0")},
		Arch:    X86_64,
		Deb:     true,
	}

	args.Args.Config = Config{
		Files: map[string]File{
			"bin0": File{
				File: "bin1",
			},
			"bin1": File{
				File: "bin1",
				Mode: "0755",
				User: "toor",
			},
		},
	}

	args = WithDefaults(args)

	assert.Equal("root", args.Args.Config.Files["bin0"].User)
	assert.Equal("0644", args.Args.Config.Files["bin0"].Mode)

	assert.Equal("toor", args.Args.Config.Files["bin1"].User)
	assert.Equal("0755", args.Args.Config.Files["bin1"].Mode)
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
