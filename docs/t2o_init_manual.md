| Author | Reviewer | Version | Update Time | link |
| ------ | -------- | ------- | ----------- | ---- |
| Win-Man | - | v1.0 | 2022-10-08 | - |

# t2o-init 介绍
t2o-init 工具是将 TiDB 全量数据导入到 Oracle 中调度程序，使用 dumpling 将 TiDB 数据导出成 CSV 文件格式，然后使用 sqlldr 工具将 CSV 导入到 Oracle 中。t2o-init 不包含 TiDB 到 Oracle 的表结构转换功能。

## sql-diff 选项介绍

```shell
t2o-init

Usage:
  dbcompare t2o-init <prepare|dump-data|generate-ctl|load-data|all> [flags]

Flags:
  -C, --config string      config file path
  -h, --help               help for t2o-init
  -L, --log-level string   log level: info, debug, warn, error, fatal
```

| 主要选项 | 用途 | 默认值 |
| ------ | --- | ----- |
| -L或--log-level | 设置日志打印级别 | info |
| -C或-config | 指定配置文件路径 | "" |
| -h | 打印帮助 | |


## sql-diff 配置文件说明

```toml
[log]
# 日志打印级别
log-level = "debug"
# dbcompare 进程日志打印路径，默认为当前路径下
log-path = "dbcompare.log"
# dbcompare 运行过程中调用其他组件产生的日志存放目录，与 log-path 无关
log-dir = "./work/log"

[performance]
# dbcompare 运行程序并发度控制，表示同时运行多少个任务，一张表为一个任务
concurrency = 5
# 是否开启行数校验，在导出数据后和导入数据后简单 select count(1) 统计单表记录数
check-row-count = true


[tidb-config]
host = '127.0.0.1'
port = 4000
user = "root"
password = ""
database = "test"
status-port = 10080
pd-addr = "127.0.0.1:2379"


[oracle-config]
host = '1.2.3.4'
port = 1539
user = "user"
password = "123456"
service-name = "orcl"
schema-name = ""

[t2o-init]
# dumpling 工具二进制包路径
dumpling-bin-path = "./dumpling"
# sqlldr 工具二进制包路径
sqlldr-bin-path = "./sqlldr"
# sqlldr 命令附加参数，会被添加到 sqlldr 命令末尾
sqlldr-extra-args = ""
# dumpling 导出备份存放路径
dump-data-dir = "./work/tidbdump"
# dumpling 命令附加参数，会被添加到 dumpling 命令末尾
dump-extra-args = "--filetype csv -t 8 --csv-separator \"|+|\"  --csv-delimiter \"\" --csv-null-value \"\" --no-header --escape-backslash"
# 生成的 Oracle ctl 配置文件存放路径
oracle-ctl-file-dir = "./work/ctl"
# ctl 模板路径
ctl-template = "./config/ctl.tmpl"
# 是否在 sqlldr 导入之前 truncate 执行 truncate 命令，也可以在 ctl 模板中配置
truncate-before-load = false
```



### 使用说明

* 创建 t2o_config 配置表

```shell
./dbcompare t2o-init prepare -C oto_config.toml
```

* 导出 TiDB 数据

```shell
./dbcompare t2o-init dump-data -C oto_config.toml
```

* 生成 ctl 控制文件

```shell
./dbcompare t2o-init generate-ctl -C oto_config.toml
```

* 使用 sqlldr 导入数据到 Oracle 

```shell
./dbcompare t2o-init load-data -C oto_config.toml
```


## 流程图
![](https://raw.githubusercontent.com/Win-Man/pic-storage/master/img/t2o_init.jpg)

## 日志说明


## Roadmap



