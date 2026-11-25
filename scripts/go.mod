module github.com/ctoforaday/special-circumstances/scripts

go 1.27.1

require (
	github.com/bufbuild/protocompile v0.14.1
	github.com/ctoforaday/special-circumstances/shared v0.0.0
	golang.org/x/tools v0.53.0
	google.golang.org/protobuf v1.36.12
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/gofrs/flock v0.13.1 // indirect
	golang.org/x/mod v0.43.0 // indirect
	golang.org/x/sync v0.25.0 // indirect
	golang.org/x/sys v0.50.0 // indirect
	golang.org/x/telemetry v0.0.0-20261116191925-e77a43154acb // indirect
)

tool golang.org/x/tools/cmd/deadcode

replace github.com/ctoforaday/special-circumstances/shared => ../shared
