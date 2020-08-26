module github.com/kentik/pkg

go 1.14

require (
	github.com/Knetic/govaluate v3.0.0+incompatible
	github.com/goreleaser/nfpm v1.3.0
	github.com/jessevdk/go-flags v1.4.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	github.com/twpayne/go-vfs v1.4.2
	golang.org/x/tools v0.0.0-20200601175630-2caf76543d99 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/goreleaser/nfpm => github.com/kentik/nfpm v1.3.1-0.20200826053905-b30b78b4a5d5
