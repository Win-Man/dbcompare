# 20230424 更新说明

sync-diff 在配置表中新增 TidbHint 和 OracleHint 字段，用于控制对比 SQL 的 hint
1. syncdiff_config_result 配置表，新增 tidb_hint 、oracle_hint 字段，用于配置数据查询时的 SQL hint
2. 更新 sync-diff-config.tmpl 模板配置文件，支持生成 oracle-hint 以及 tidb-hint 配置。

# 20230323 更新说明

支持在 Oracle 数据导出以及数据对比节点，通过设置 where 过滤条件，实现增量迁移。

主要更新内容：
1. o2t_config 配置表，新增 dump_filter_clause_ora 字段，用于配置导出阶段 where 过滤条件, 举例： “id > 10 and id < 100”
2. syncdiff_config_result 配置表，之前有设计 filter_clause_tidb 字段和 filter_clause_ora 字段，用于配置 o2t-sync-diff 的数据范围选择，举例: "id > 10 and id <100"
3. 更新 sync-diff-config.tmpl 模板配置文件，支持生成 range 以及 oracle-range 配置。