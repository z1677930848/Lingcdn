# Repository Guidelines

## 项目结构与模块划分
- `EdgeCommon/`：Go 公共库，提供模型、proto 与工具函数，API/Node/Admin 共享。
- `Lingapi-master/`：边缘控制 API 服务，入口 `cmd/edge-api`，安装器脚本在 `cmd/edge-instance-installer` 等，构建产物放 `dist/`。
- `Lingnode-master/`：节点守护与流量处理，入口 `cmd/ling-node`，核心逻辑在 `internal/*`（caches/monitor/waf 等）。
- `Lingadmin-master/`：管控端，后端 `cmd/edge-admin`、`cmd/lingcdnadmin`，静态前端在 `web/`（terser 压缩 JS），文档 `doc/`，发布包 `dist/`。
- 其他：`docs/` 记录远程开发实践；`tools/` 辅助脚本；根部 `.env.local` 示例环境变量。

## 构建、测试与本地运行
- 统一使用 Go 1.22、Node 18。运行 Go 命令前保持当前目录在仓库根，确保 `replace ../EdgeCommon` 生效。
- 公共库：`cd EdgeCommon && go test ./...`。
- API：`cd Lingapi-master && go build ./cmd/edge-api && go test ./...`；安装器切换到对应 `cmd/*`。
- 节点：`cd Lingnode-master && go build ./cmd/ling-node && go test ./...`，性能/并发改动建议追加 `go test -race ./...`。
- 管控端后端：`cd Lingadmin-master && go build ./cmd/edge-admin && go test ./...`。
- 管控端前端：`cd Lingadmin-master/web && npm install && npm run build`（`build:js` 使用 terser 压缩 `public/js/components.src.js`）。

## 代码风格与命名
- Go：提交前执行 `gofmt ./...` 或 `goimports`；函数/变量小驼峰，配置键可保留下划线；错误使用 `%w` 包装，优先显式传递 `context.Context`。
- Lint：`Lingapi-master` 提供 `.golangci.yaml`，推荐 `golangci-lint run ./...` 做静态检查；避免未使用依赖，必要时 `go mod tidy`。
- 前端：文件名用 kebab-case，公共代码放 `public/js`，注释/文案使用中文，避免提交临时/未压缩产物。

## 测试准则
- 新功能需附带包级单元测试，遵循表驱动用例，文件命名 `xxx_test.go`。
- 提交前至少执行 `go test ./...`；涉及并发/IO 的改动在 Node/API 模块补充 `-race` 或长跑用例。
- 前端改动确保 `npm run build` 通过，UI 变更提供手动验收要点或截图。

## 提交与 PR 规范
- 提交信息沿用历史前缀：`fix: ...`、`feat: ...`、`chore: ...`，单次改动聚焦一件事。
- PR 需说明变更背景、影响模块、已执行的构建/测试命令、关联 Issue/需求；前端/界面改动附前后截图或构建体积变化。
- 禁止提交真实密钥/证书；示例/脚本使用占位符，敏感配置按 `docs/remote-dev-ssh.md` 的流程初始化。

## 安全与配置提醒
- `.env.local` 仅作示例，实际密钥请放入本地或受控的秘密管理，勿入库。
- 远程调试遵循 `docs/remote-dev-ssh.md`，如需开放端口或修改防火墙，请先评估安全影响。
