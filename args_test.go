package main

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
)

func testArgs() *Args {
	return &Args{
		Name:    "test",
		Version: Version{semver.MustParse("1.0.0")},
		Arch:    X86_64,
		Deb:     true,
		RPM:     false,
		Inputs: Inputs{
			Config: Config{
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

func TestUnmarshalFormat(t *testing.T) {
	assert := assert.New(t)
	expect := map[string]Format{
		"deb": DEB,
		"rpm": RPM,
		"DEB": DEB,
	}

	for value, fmt := range expect {
		tmp := Format("")
		err := tmp.UnmarshalFlag(value)
		assert.NoError(err)
		assert.Equal(fmt, tmp)
	}
}
