module github.com/jonhadfield/sncli

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/jonhadfield/gosn v0.0.0-20181120211442-db2aeaf434a9
	github.com/spf13/viper v1.2.1
	github.com/stretchr/testify v1.2.2
	golang.org/x/crypto v0.0.0-20181112202954-3d3f9f413869
	gopkg.in/urfave/cli.v1 v1.20.0
	gopkg.in/yaml.v2 v2.2.1
)

replace github.com/jonhadfield/gosn => ../gosn
