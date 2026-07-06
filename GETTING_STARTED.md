
# 快速开始

## 前置要求

- Go 1.23+
- Docker & Docker Compose

## 本地开发

### 1. 启动依赖服务

```bash
make up
```

这将启动 PostgreSQL 和 Redis。

### 2. 初始化测试数据

在另一个终端中：

```bash
cd scripts
go run init_test_data.go
```

### 3. 启动应用

```bash
go run cmd/api/main.go
```

或者使用 make：

```bash
make dev
```

## 测试端点

### 健康检查

```
http://localhost:8080/health
```

### 访问 SVG 计数器

使用我们刚才创建的测试项目：

```
http://localhost:8080/svg/my-awesome-project/counter/visits.svg
```

你也可以自定义颜色和标签：

```
http://localhost:8080/svg/my-awesome-project/counter/visits.svg?color=%23ff6b6b&label=Visitors
```

### 访问 SVG 徽章

```
http://localhost:8080/svg/my-awesome-project/badge/downloads.svg?label=Downloads
```

使用 flat 样式：

```
http://localhost:8080/svg/my-awesome-project/badge/downloads.svg?label=Downloads&style=flat
```

### 获取统计数据

```
http://localhost:8080/api/v1/projects/my-awesome-project/stats
```

## 在 GitHub README 中使用

将 SVG 链接直接放到你的 README.md 中：

```markdown
![Visits](http://localhost:8080/svg/my-awesome-project/counter/visits.svg)

![Downloads](http://localhost:8080/svg/my-awesome-project/badge/downloads.svg?label=Downloads)
```

## 停止服务

```bash
make down
```

清理所有数据：

```bash
make clean
```

