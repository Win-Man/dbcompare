module github.com/Win-Man/dbcompare

go 1.16

replace (
	github.com/Win-Man/dbcompare/compare => ./dbcompare/compare
	github.com/Win-Man/dbcompare/config => ./dbcompare/config
	github.com/Win-Man/dbcompare/database => ./dbcompare/database
	github.com/Win-Man/dbcompare/cmd => ./dbcompare/cmd
)

require (
	github.com/BurntSushi/toml v0.4.1
	github.com/go-sql-driver/mysql v1.6.0
	github.com/godror/godror v0.30.2
	github.com/jmoiron/sqlx v1.3.4
	github.com/mattn/go-sqlite3 v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.5.0 // indirect
	go.uber.org/zap v1.19.1
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
