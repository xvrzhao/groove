# Groove

Groove 是一个极简的 HTTP/Cron 服务脚手架，集成了基础的 Web 开发 Package，包括 JWT 鉴权、日志、密码哈希、分页查询、标准化 HTTP 响应格式等，并附带适用于敏捷开发的 CRUD 接口一键生成工具，非常适合小型单体后端服务的开发。

## 使用方法

```bash
# 安装 Groove 命令行工具
go install github.com/xvrzhao/groove@latest

# 创建您的 Groove 项目
groove create my-app

# 使用 vscode 开始开发您的项目
code ./my-app
```

## Groove 设计哲学

### 环境变量

#### 环境变量的读取

Groove 的设计理念为 Go 应用本身是一个无状态应用，它仅仅是一个逻辑空壳，而不同的环境变量（状态）将赋予它不同的行为。

Go 应用自身不处理环境变量的导出或配置文件的读取，环境变量在启动时由 `bin/go` (本地环境)、`bin/exec` (容器中) 脚本读取项目根路径下的 .env 文件并导出给应用，Go 应用只需从环境变量中读取即可。所以，您在执行 `go` 命令时，需要使用 `bin/go` 来代替，例如：`bin/go test ./...`、`bin/go run ./cmd/http` 等。

**这样设计的另一个原因是，若 Go 应用中处理配置文件的读取，在执行单元测试时工作目录 (`pwd`) 将会变更到单元测试的包中，会导致配置文件读取不到而报错。**

#### 环境变量文件

无论是本地环境还是线上容器环境，启动 Groove 应用都将从项目根路径下读取 .env 文件，为了避免重要信息被泄露，该文件已列入 .gitignore，所以您在启动项目前需先将 .env.example 拷贝一份并命名为 .env。

#### 环境变量与镜像发布

Groove 脚手架集成了 `make publish` 快捷指令来打包发布镜像，如果您细心读了 Makefile 和 Dockerfile 两个文件，您会发现在构建镜像过程中将根据 `make publish` 的 `env` 参数或 `docker build` 的 `PUBLISH_MODE` 参数来制定发布环境，而在 Dockerfile 中将根据制定的环境，将 **.env.环境名称** 文件拷贝成 .env 来供应用使用。

所以，您可以创建出多个不同的环境变量文件，如 **.env.development**、**.env.production** 等，这样，在执行例如 `make publish version=1.0.0 env=development` 或 `make publish version=1.0.0 env=production` 时，将会分别应用到对应的环境变量文件。

#### 环境变量的读取时机

Go 应用由若干个包组成，而包是独立自治的单位，对外提供可用的接口，所以为了保证包对外提供的内容始终可用，应当在包的初始化阶段 (`init` 函数) 就将对外提供的内容初始化好，所以应在 `init` 函数中读取环境变量并初始化对外暴露的内容。

您可参考 db/conn.go 文件，db 包对外仅提供 Client 变量，Client 是一个数据库客户端，其初始化和配置的读取在 `init` 阶段完成。

**Groove 非常不倡导将初始化工作放在运行阶段(`main` 函数开始执行之后)，这在某些情况下同样将导致单元测试无法进行，除非在单元测试中再手动写入初始化的代码，例如初始化数据库连接。**

### 镜像与版本

Groove 是由 HTTP 和 Cron 两个应用构成，由于 Groove 是一个单体项目，故对公共包的修改将对两个应用都会产生影响，所以 Groove 认为两个程序入口应当保持同样的版本号，镜像构建过程会将两个入口都编译在同一镜像中，您在部署时只需针对同一镜像使用不同的启动指令 (Dockerfile 的 `CMD` 或 docker-compose 的 `command`) 即可。

### HTTP 请求生命周期

Groove 尊崇经典的 Controller - Service - Model 三层架构设计。

#### Controller 层

Controller 层负责请求参数的校验和响应数据的封装。

#### Service 层

Service 层是业务逻辑的核心，该层的 API 应具有通用和抽象能力，以供不同业务模块调用。抽离 Service 层的目的在于模拟微服务架构调用关系，**每个 Service 是一个独立的微服务**，Service 之间互相配合互相调用。**同时，每个业务模块的 Service 应当只有权限调用自己所管辖的 Model，而不应该调用本辖区外的 Model，若要使用应通过 Service 之间的 API 调用来实现。**这样严格的代码纪律将会非常便于日后真正的微服务拆分。

#### Model 层

Model 层为业务中使用到的数据模型，仅声明 Model 自身数据属性，不包含任何业务逻辑。

### 数据库迁移文件

Groove 使用简单的 SQL 文件表示数据库的 DDL 操作，其命名规则为 `日期.业务名称.sql`，这样文件将会按照日期顺序排序，由于项目可能会由多人开发，这样排序可以保证在部署时能够按照正确的顺序执行 SQL，不会出现错误覆盖操作。

### DEBUG 和单元测试

Groove 自带了 vscode launch.json 文件，内部声明了 HTTP debug 启动入口，启动时会从 .env 中读取环境变量，方便 debug 调试。

Groove 提倡编写单元测试。

### 优雅关闭

虽然 Go 应用在容器中由 bin/exec 脚本启动，但可确保 Go 应用为容器主程序 (pid: 1)，可正常接收容器的 term/kill 等信号，您可自行在代码中编写优雅退出逻辑。

## Groove 快捷指令

### 本地启动 Groove App

```bash
# 本地启动 HTTP 服务
make run-http
# 本地启动 Cron 程序
make run-cron
```

### 一键生成 CRUD

```bash
# 参数说明:
#   table: 数据表名称
#   model: 生成的 Model 的名称
make api table=x_persons model=Person
```

### 发布镜像

```bash
# 参数说明:
#   version: 版本号
#   env: 发布的环境，不同环境将打包不同的 env 文件
make publish version=1.0.0 env=development
```
