# Lingcdn Docker 插件测试示例

为了方便在本地验证 VS Code Docker 插件是否正常工作，这里提供一个极简的 Go Web 服务示例。容器启动后会暴露 `8080` 端口，并提供 `/`、`/info`、`/healthz` 三个接口，适合用来验证构建、启动、日志、终端、端口转发等常见开发动作。

## 目录结构

```
tools/docker-demo
├── Dockerfile          # 构建 demo 服务镜像
├── docker-compose.yml  # 一键启动服务
├── main.go             # Go HTTP 服务示例
└── README.md
```

## 在 VS Code Docker 插件中操作

1. **开启 Docker 引擎**  
   启动 Docker Desktop（或其他 Docker 服务），在终端运行 `docker version` 确认已安装。

2. **在 VS Code 打开仓库根目录**  
   保证左侧资源管理器能看到 `tools/docker-demo` 目录。

3. **使用 Docker 插件构建镜像**  
   - 打开 VS Code 命令面板（`Ctrl+Shift+P`），执行 `Docker: Build Image...`。  
   - 选择 `tools/docker-demo/Dockerfile`，为镜像命名（例如 `lingcdn/docker-demo:local`）。  
   - 观察终端输出，确保镜像构建成功后会出现在 Docker 插件面板的 “Images” 中。

4. **通过 Docker 插件启动容器**  
   - 在资源管理器定位到 `tools/docker-demo/docker-compose.yml`，右键选择 `Compose Up`（或在 Docker 插件面板的 “Contexts/Composes” 区域执行 `Up`）。  
   - 容器启动后，将在 “Containers” 面板看到 `lingcdn-docker-demo-demo-web-1`，可直接右键查看日志或附加终端。

5. **验证结果**  
   - 打开浏览器访问 [http://localhost:8080](http://localhost:8080) 查看示例页面。  
   - 访问 [http://localhost:8080/info](http://localhost:8080/info) 可看到 JSON 输出，用于测试端口映射和热重载。  
   - `docker ps`、VS Code Docker 面板日志、`/healthz` 接口都应显示容器运行状态。

6. **清理**  
   - 在 Docker 插件面板中，对应容器执行 `Stop` 或 `Remove`。  
   - 如果使用了 compose，可右键 `docker-compose.yml` 执行 `Compose Down`。  
   - 如需重新构建，先 `Compose Down`，再重新执行步骤 3。

## 使用命令行（可选）

若想直接验证 docker 命令是否可用，可在仓库根目录执行：

```powershell
cd tools/docker-demo
docker compose up --build
```

终端出现 `demo 服务启动：http://localhost:8080` 后即可访问浏览器，使用 `Ctrl+C` 结束并自动清理。

## 常见排查思路

- 镜像构建失败：确认网络可访问 `golang:1.22-alpine`，必要时配置镜像加速；若 `go` 下载依赖慢，可提前设置代理。  
- 容器启动立即退出：查看 VS Code Docker 面板日志，确认端口未被占用，或修改 `docker-compose.yml` 中的 `PORT`。  
- 插件未显示容器：执行 `Docker: Refresh` 或手动运行 `docker ps` 排除 Docker 服务未启动的问题。
