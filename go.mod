module github.com/jonhadfield/sn-cli

require (
	github.com/BurntSushi/toml v0.3.1 // indirect
	github.com/jonhadfield/gosn v0.0.0-20181217180752-d3e625717966
	github.com/spf13/afero v1.2.0 // indirect
	github.com/spf13/viper v1.3.1
	github.com/stretchr/testify v1.2.2
	golang.org/x/crypto v0.0.0-20181203042331-505ab145d0a9
	golang.org/x/sys v0.0.0-20181213200352-4d1cda033e06 // indirect
	gopkg.in/urfave/cli.v1 v1.20.0
	gopkg.in/yaml.v2 v2.2.2
)

replace github.com/jonhadfield/gosn => ../gosn
