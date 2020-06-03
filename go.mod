module github.com/kentik/pkg

go 1.14

require (
	github.com/Knetic/govaluate v3.0.0+incompatible
	github.com/Masterminds/semver/v3 v3.1.0
	github.com/goreleaser/nfpm v1.3.0
	github.com/jessevdk/go-flags v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.0
	golang.org/x/tools v0.0.0-20200601175630-2caf76543d99 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/goreleaser/nfpm => github.com/kentik/nfpm v1.3.1-0.20200603061708-988e1152d67c
