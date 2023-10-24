module github.com/Win-Man/dbcompare

go 1.19

replace (
	github.com/Win-Man/dbcompare/cmd => ./dbcompare/cmd
	github.com/Win-Man/dbcompare/compare => ./dbcompare/compare
	github.com/Win-Man/dbcompare/config => ./dbcompare/config
	github.com/Win-Man/dbcompare/database => ./dbcompare/database
	github.com/Win-Man/dbcompare/models => ./dbcompare/models
	github.com/Win-Man/dbcompare/pkg => ./dbcompare/pkg
)

require (
	github.com/BurntSushi/toml v1.3.2
	github.com/go-sql-driver/mysql v1.7.1
	github.com/godror/godror v0.40.3
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.7.0
	go.uber.org/zap v1.26.0
	golang.org/x/crypto v0.14.0
	gorm.io/driver/mysql v1.5.2
	gorm.io/gorm v1.25.5
)

require (
	github.com/go-logfmt/logfmt v0.6.0 // indirect
	github.com/godror/knownpb v0.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9 // indirect
	golang.org/x/sys v0.13.0 // indirect
	google.golang.org/protobuf v1.30.0 // indirect
)
