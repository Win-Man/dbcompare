# dbcompare - 数据库比较和迁移工具

一款用于在不同数据库系统之间进行比较和迁移数据的强大命令行工具，主要用于 Oracle↔MySQL、Oracle↔TiDB 和 MySQL↔MySQL 数据库间的比较。

## 概述

dbcompare 是一个多用途的数据库比较工具，提供以下功能：

- **SQL 查询结果比较**: 在不同数据库系统上执行相同的 SQL 查询并比较结果以验证兼容性
- **数据库迁移辅助**: 促进 Oracle、MySQL 和 TiDB 等数据库之间的数据迁移过程
- **表结构和数据一致性检查**: 验证数据库迁移后的数据完整性
- **自动化数据传输**: 使用可配置的过程简化批量数据移动

## 功能特性

### 1. SQL 比较 (sql-diff)
- 在源数据库和目标数据库上执行相同的 SQL 查询
- 比较查询结果以验证等价性
- 支持多种数据库类型（MySQL、Oracle、TiDB）
- 输入方法：命令行、文件或基于配置的 SQL 语句

### 2. 数据同步控制 (sync-diff) 
- 并行同步不同数据库间的数据，支持可配置的并发性
- 为 TiDB Lightning 自动生成配置文件
- 进度跟踪和状态监控
- 支持基于快照的一致性读取
- 多线程处理，支持可配置的并发数

### 3. Oracle 到目标初始化 (o2t-init)
- 使用 sqluldr2 将 Oracle 数据批量导出为 CSV 文件
- 自动为 TiDB Lightning 生成配置文件
- 带有进度监控的目标数据库加载
- 行计数验证功能
- 多阶段处理：导出 → 生成配置 → 加载 → 验证

### 4. 目标到 Oracle 初始化 (t2o-init) 
- 使用 Dumpling 从 MySQL/TiDB 导出数据
- 自动为 Oracle SQL*Loader 生成控制文件
- 使用 SQL*Loader 的多线程加载到 Oracle
- 可配置的截断和预处理选项
- 全面的错误处理和日志记录

## 支持的数据库

- **MySQL**（以及通过 MySQL 协议兼容的系统如 TiDB）
- **Oracle**（使用 godror 驱动）
- **TiDB**（作为 MySQL 兼容目标）

## 架构

### 核心组件
- **CLI 层**: 基于 Cobra 框架构建的直观命令行界面
- **数据库抽象**: 支持多种数据库类型的统一接口
- **配置管理**: 灵活的 YAML 和 TOML 配置支持
- **并发引擎**: 用于高效并行处理的多线程技术  
- **进度监控**: 实时进度条和统计信息

### 文件结构
- **cmd/**: 基于 Cobra 的 CLI 命令和入口点
- **database/**: 数据库连接器和抽象层（MySQL 通过 GORM，Oracle 通过原生 SQL）
- **compare/**: SQL 比较逻辑
- **config/**: 配置加载和验证
- **models/**: 用于内部元数据的 GORM 模型定义
- **pkg/**: 可重用的实用程序包

## 安装

### 依赖项

1. **外部工具**（用于数据加载）：
   - `sqluldr2` 用于 Oracle 数据提取
   - `sqlldr` 用于 Oracle 数据加载
   - `tidb-lightning` 用于快速 MySQL/TiDB 加载
   - `dumpling` 用于 MySQL/TiDB 提取

2. **Go 模块要求**：
   - Go 1.19 或更高版本

### 构建

1. 克隆仓库：
   ```bash
   git clone https://github.com/Win-Man/dbcompare.git
   ```

2. 安装依赖：
   ```bash
   go mod download
   ```

3. 使用 Makefile 构建：
   ```bash
   make build
   ```

## 配置

所有工具都需要配置文件，指定以下参数：

- 数据库连接详情（主机、端口、用户、密码）
- 表结构信息
- 处理参数
- 临时文件路径配置
- 性能设置

请参考分发中包含的 `.tmpl` 文件获取示例配置。

## 使用示例

### SQL 比较
```bash
./bin/dbcompare sql-diff --config config.yaml --sql "SELECT COUNT(*) FROM users"
```

### 数据同步（准备阶段）
```bash
./bin/dbcompare sync-diff prepare --config sync_config.yaml
```

### 数据同步（运行阶段）
```bash
./bin/dbcompare sync-diff run --config sync_config.yaml
```

### Oracle 到 MySQL/迁移准备
```bash
./bin/dbcompare o2t-init all --config o2t_config.yaml
```

### MySQL 到 Oracle
```bash
./bin/dbcompare t2o-init all --config t2o_config.yaml
```

## 项目状态

- 针对数据库迁移开发的活跃工具
- 专门用于 Oracle↔MySQL/TiDB 迁移场景
- 使用经过实战检验的组件如 Dumpling、TiDB Lightning 和 SQL*Loader
- 包括先进的进度跟踪和错误处理

## 贡献

- 欢迎为额外数据库连接器支持提供贡献
- 为性能优化提供增强提议
- 针对数据转换中的边缘情况提出错误报告和修复

## 许可证

Apache 2.0 - 详细信息请参见 LICENSE 文件

## 注意事项

- 需要设置外部数据库工具（sqluldr2、sqlldr、dumpling、tidb-lightning）
- 在构建后配置模板在 bin/ 目录中提供
- 专为企业数据库迁移场景设计
- 为高效的大型转换实现健壮的多线程机制

## 开发路线图

根据原始待办清单指出：
- 增强的日志和完成状态报告
- 更多数据库类型支持
- 与更多外部数据库工具集成