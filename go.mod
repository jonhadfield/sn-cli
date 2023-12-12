module github.com/jonhadfield/sn-cli

go 1.21

require (
	github.com/asdine/storm/v3 v3.2.1
	github.com/briandowns/spinner v1.23.0
	github.com/divan/num2words v0.0.0-20170904212200-57dba452f942
	github.com/fatih/color v1.16.0
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/jonhadfield/gosn-v2 v0.0.0-20231211223627-abb88b737146
	github.com/ryanuber/columnize v2.1.2+incompatible
	github.com/spf13/viper v1.18.1
	github.com/stretchr/testify v1.8.4
	github.com/urfave/cli v1.22.14
	golang.org/x/crypto v0.16.0 // indirect
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.3 // indirect
	github.com/danieljoos/wincred v1.2.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/matryer/try v0.0.0-20161228173917-9ac251b645a2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/smartystreets/goconvey v1.8.1 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/zalando/go-keyring v0.2.3 // indirect
	go.etcd.io/bbolt v1.3.8 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/term v0.15.0
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/alessio/shellescape v1.4.2 // indirect
	github.com/google/uuid v1.4.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-retryablehttp v0.7.5 // indirect
	github.com/pelletier/go-toml/v2 v2.1.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/santhosh-tekuri/jsonschema/v5 v5.3.1 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/exp v0.0.0-20231206192017-f3f8817b8deb // indirect
)

// replace github.com/jonhadfield/gosn-v2 => ../gosn-v2
replace github.com/jonhadfield/gosn-v2 => github.com/clayrosenthal/gosn-v2 v0.0.0-20231212073032-3d59f6965163
