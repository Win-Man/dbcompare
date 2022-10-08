| Author | Reviewer | Version | Update Time | link |
| ------ | -------- | ------- | ----------- | ---- |
| Win-Man | - | v1.0 | 2022-10-08 | - |

# o2t-init 介绍
o2t-init 工具是将 Oracle 全量数据导入到 TiDB 中调度程序，使用 sqluldr2 将 Oracle 数据导出成 CSV 文件格式，然后使用 tidb-lightning 工具将 CSV 导入到 TiDB 中。o2t-init 不包含 Oracle 到 TiDB 的表结构转换功能。

## sql-diff 选项介绍

```shell
o2t-init

Usage:
  dbcompare o2t-init <prepare|dump-data|generate-conf|load-data|all> [flags]

Flags:
  -C, --config string      config file path
  -h, --help               help for o2t-init
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


[o2t-init]
# sqluldr2 工具二进制包路径
sqluldr2-bin-path = "./sqlldr2"
# sqluldr2 命令附加参数，会被添加到 sqluldr2 命令末尾
sqluldr2-extra-args = " batch=yes field=\"0x7c0x230x7c\" record=\"0x7c0x2b0x7c0x0a\" charset=ZHS16GBK safe=yes"
# sqluldr2 导出备份存放路径
dump-data-dir = "./work/oracledump"
# lightning 工具二进制包路径
lightning-bin-path = "./tidb-lightning"
# lightning 配置文件模板
lightning-toml-template = "./config/lightning_toml.tmpl"
# 生成的 lightning 配置文件存放路径
lightning-toml-dir = "./config/"
# lighting 命令附加参数，会被添加到 lightning 命令末尾
lightning-extra-args = " "


```



### 使用说明

* 创建 o2t_config 配置表

```shell
./dbcompare o2t-init prepare -C oto_config.toml
```

* 导出 Oracle 数据

```shell
./dbcompare o2t-init dump-data -C oto_config.toml
```

* 生成 lighting 配置文件

```shell
./dbcompare o2t-init generate-conf -C oto_config.toml
```

* 使用 lighting 导入数据到 TiDB 

```shell
./dbcompare o2t-init load-data -C oto_config.toml
```


## 流程图
![](https://raw.githubusercontent.com/Win-Man/pic-storage/master/img/o2t_init.jpg)

## 日志说明


## Roadmap



