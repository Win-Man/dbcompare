module github.com/Win-Man/dbcompare

go 1.16

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
	github.com/benbjohnson/clock v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.7.1
	github.com/godror/godror v0.40.3
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.7.0
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/zap v1.26.0
	golang.org/x/crypto v0.14.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gorm.io/driver/mysql v1.5.2
	gorm.io/gorm v1.25.5
)
