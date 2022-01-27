## dbcompare 介绍

dbcompare 是为了对比相同的 SQL 在不同数据库之间执行结果是否有区别而写的一个工具。目前支持仅支持 MySQL、TiDB、Oracle 数据库之间的执行情况对比。

## dbcompare 选项介绍

```
Usage of ./dbcompare:
  -L string
    	log level: info, debug, warn, error, fatal (default "info")
  -config string
    	config file path
  -h	this help
  -log-path string
    	The path of log file
  -output string
    	print|file
  -sql string
    	single sql statement
```

| 主要选项 | 用途 | 默认值 |
| ------ | --- | ----- |
| -L | 设置日志打印级别 | info |
| -config | 指定配置文件路径 | "" |
| -h | 打印帮助 | |
| -log-path | 指定输出日志的路径 | 默认为 "" ,表示日志输出在当前目录下 |
| -output | 指定输出结果的路径，可以设置为 `print` 或者 `file`。`print` 表示直接将输出结果打印在屏幕上，不保存结果；`file` 表示将输出结果输出到文件中保存 | 默认为 "" |
| -sql | 指定单个 SQL 执行 | 默认是 "" |

## dbcompare 配置文件说明



### 使用说明

```
./dbcompare -config config/dev.toml -output=print
./dbcompare -config config/dev.toml -output=file
./dbcompare -config config/dev.toml -sql="select * from t1" -output=file
```

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


<!-- LICENSE -->
## License

dbcompare is under the Apache 2.0 license. See the [LICENSE](./LICENSE) file for detail
