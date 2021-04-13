module github.com/jonhadfield/sn-cli

go 1.14

require (
	github.com/asdine/storm/v3 v3.2.1
	github.com/briandowns/spinner v1.11.1
	github.com/cpuguy83/go-md2man/v2 v2.0.0 // indirect
	github.com/divan/num2words v0.0.0-20170904212200-57dba452f942
	github.com/fatih/color v1.9.0
	github.com/jonhadfield/gosn-v2 v0.0.0-20200709170748-a8769dfb8e8b
	github.com/ryanuber/columnize v2.1.0+incompatible
	github.com/spf13/viper v1.7.1
	github.com/stretchr/testify v1.7.0
	github.com/urfave/cli v1.22.4
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	gopkg.in/yaml.v2 v2.4.0
)

replace github.com/jonhadfield/gosn-v2 => ../gosn-v2
