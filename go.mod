module github.com/kentik/pkg

go 1.14

require (
	github.com/Masterminds/semver/v3 v3.1.0
	github.com/goreleaser/nfpm v1.3.0
	github.com/jessevdk/go-flags v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/goreleaser/nfpm => github.com/kentik/nfpm v1.3.1-0.20200529022637-4f43e41f673c
