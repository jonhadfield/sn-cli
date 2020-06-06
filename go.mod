module github.com/jonhadfield/sn-cli

go 1.14

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/divan/num2words v0.0.0-20170904212200-57dba452f942
	github.com/fatih/color v1.9.0
	github.com/jonhadfield/gosn-v2 v0.0.0-20200517210619-52110795737e
	github.com/jonhadfield/sn-persist v0.0.0-20200602194813-8f34c75209e3
	github.com/mitchellh/mapstructure v1.3.1 // indirect
	github.com/pelletier/go-toml v1.8.0 // indirect
	github.com/ryanuber/columnize v2.1.0+incompatible
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli v1.22.4
	golang.org/x/crypto v0.0.0-20200604202706-70a84ac30bf9
	golang.org/x/sys v0.0.0-20200602100848-8d3cce7afc34 // indirect
	gopkg.in/ini.v1 v1.57.0 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/jonhadfield/sn-persist => ../sn-persist

replace github.com/jonhadfield/gosn-v2 => ../gosn-v2
