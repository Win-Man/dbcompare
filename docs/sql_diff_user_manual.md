| Author | Reviewer | Version | Update Time | link |
| ------ | -------- | ------- | ----------- | ---- |
| Win-Man | - | v1.0 | 2022-09-23 | - |

# sql-diff 介绍
sql-diff 是为了对比相同的 SQL 在不同数据库之间执行结果是否有区别而写的一个工具。目前支持仅支持 MySQL、TiDB、Oracle 数据库之间的执行情况对比。

## sql-diff 选项介绍

```shell
sql-diff

Usage:
  dbcompare sql-diff [flags]

Flags:
  -C, --config string      config file path
  -h, --help               help for sql-diff
  -L, --log-level string   log level: info, debug, warn, error, fatal
      --log-path string    The path of log file
      --output string      print|file
      --sql string         single sql statement
```

| 主要选项 | 用途 | 默认值 |
| ------ | --- | ----- |
| -L或--log-level | 设置日志打印级别 | info |
| -C或-config | 指定配置文件路径 | "" |
| -h | 打印帮助 | |
| --log-path | 指定输出日志的路径 | 默认为 "" ,表示日志输出在当前目录下 |
| --output | 指定输出结果的路径，可以设置为 `print` 或者 `file`。`print` 表示直接将输出结果打印在屏幕上，不保存结果；`file` 表示将输出结果输出到文件中保存 | 默认为 "" |
| --sql | 指定单个 SQL 执行 | 默认是 "" |

## sql-diff 配置文件说明

```toml
[log]
# 设置日志级别
log-level = "Debug"
# 设置日志路径以及日志名
log-path = "dbcompare.log"

# MySQL 连接配置
[mysql-config]
# MySQL 连接地址
host = '127.0.0.1'
# MySQL 连接端口
port = 3306
# MySQL 连接用户名
user = "root"
# MySQL 连接密码
password = "pass"
# MySQL 连接库
database = "dev"

# TiDB 连接配置
[tidb-config]
# TiDB 连接地址
host = '1.2.3.4'
# TiDB 连接端口
port = 4000
# TiDB 连接用户名
user = "root"
# TiDB 连接密码
password = ""
# TiDB 连接库
database = "test"

# Oracle 连接配置
[oracle-config]
# Oracle 连接地址
host = '1.2.3.4'
# Oracle 连接端口
port = 1539
# Oracle 连接用户名
user = "user"
# Oracle 连接密码
password = "123456"
# Oracle 连接服务
service-name = "orcl"
# Oracle 连接 schema
schema-name = ""

# 对比相关配置
[compare-config]
# 对比数据库的类型，会根据类型自动读取对应的连接配置，目前仅支持两两对比
source-type = "tidb" ## tidb|mysql|oracle
dest-type = "oracle"
# 需要对比的 SQL 语句文件存放路径
sqlfile = "./config/a.sql"
# SQL 语句文件中语句的分隔符
sqlfile-delimiter = ";"
# 指定输出结果的路径，可以设置为 `print` 或者 `file`。`print` 表示直接将输出结果打印在屏幕上，不保存结果；`file` 表示将输出结果输出到文件中保存
output = "file"    ## file|print 
# 当 output 输出设置为 file 的时候，指定输出文件的前缀名
outputprefix = "output"
```



### 使用说明

* 指定将输出结果打印在屏幕上

```shell
./dbcompare_amd64 sql-diff -C config/dev.toml -output=print
```

* 指定将输出结果保存在文件中

``` shell
./dbcompare_amd64 sql-diff -C config/dev.toml -output=file
```

* 测试单条 SQL 执行结果

```shell
./dbcompare_amd64 sql-diff -C  config/dev.toml -sql="select * from t1" -output=print
```

## 日志说明


## Roadmap

- [ ] 支持主流关系型数据库
    - [x] MySQL
    - [x] TiDB
    - [x] Oracle
    - [ ] DB2
    - [ ] PostgreSQL
- [x] 支持执行 SQL 结果对比
- [ ] 支持表结构对比
- [ ] 支持数据对比

