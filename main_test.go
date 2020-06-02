package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
