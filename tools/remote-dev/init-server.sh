#!/usr/bin/env bash
# 远程云服务器初始化脚本（方案1：VS Code Remote-SSH + 原生环境）
# 说明：在新购置的 Linux 云服务器上以 root 身份执行本脚本，可一次性完成常用依赖安装、
#       Go/Node 环境配置以及开发用户目录初始化，方便通过 Remote-SSH 直接开发 Lingcdn。

set -euo pipefail

if [[ $EUID -ne 0 ]]; then
  echo "请使用 root 或 sudo 运行本脚本。" >&2
  exit 1
fi

GO_VERSION="${GO_VERSION:-1.22.3}"
NODE_MAJOR="${NODE_MAJOR:-18}"
TIMEZONE="${TIMEZONE:-Asia/Shanghai}"
DEV_USER="${REMOTE_DEV_USER:-${SUDO_USER:-lingcdn}}"
DEV_HOME="$(eval echo "~${DEV_USER}" 2>/dev/null || true)"

log() {
  echo -e "\033[1;32m[init]\033[0m $*"
}

detect_pkg_manager() {
  if command -v apt-get >/dev/null 2>&1; then
    echo "apt"
  elif command -v dnf >/dev/null 2>&1; then
    echo "dnf"
  elif command -v yum >/dev/null 2>&1; then
    echo "yum"
  else
    echo ""
  fi
}

ensure_user() {
  if id "$DEV_USER" >/dev/null 2>&1; then
    log "开发用户 ${DEV_USER} 已存在。"
  else
    log "创建开发用户 ${DEV_USER} ..."
    useradd -m -s /bin/bash "$DEV_USER"
    passwd -d "$DEV_USER"
    DEV_HOME="$(eval echo "~${DEV_USER}")"
  fi
}

install_common_packages() {
  local mgr="$1"
  log "安装常用基础包（包管理器：${mgr}）..."
  case "$mgr" in
    apt)
      export DEBIAN_FRONTEND=noninteractive
      apt-get update -y
      apt-get install -y build-essential git curl wget unzip tar htop pkg-config ca-certificates gnupg lsb-release
      ;;
    dnf|yum)
      "$mgr" install -y epel-release || true
      "$mgr" install -y gcc gcc-c++ make git curl wget unzip tar htop pkgconfig ca-certificates gnupg2
      ;;
    *)
      echo "暂不支持的包管理器，请手动安装基础依赖。" >&2
      exit 1
      ;;
  esac
}

setup_timezone() {
  log "设置系统时区为 ${TIMEZONE} ..."
  ln -sf "/usr/share/zoneinfo/${TIMEZONE}" /etc/localtime
  if command -v timedatectl >/dev/null 2>&1; then
    timedatectl set-timezone "${TIMEZONE}" || true
  fi
}

install_go() {
  if command -v go >/dev/null 2>&1 && [[ "$(go env GOVERSION || true)" == "go${GO_VERSION}" ]]; then
    log "Go ${GO_VERSION} 已安装，跳过。"
    return
  fi

  log "安装 Go ${GO_VERSION} ..."
  local tmp="/tmp/go${GO_VERSION}.linux-amd64.tar.gz"
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o "${tmp}"
  rm -rf /usr/local/go
  tar -C /usr/local -xzf "${tmp}"
  cat >/etc/profile.d/go.sh <<EOF
export PATH=\$PATH:/usr/local/go/bin
EOF
  chmod +x /etc/profile.d/go.sh
}

install_node() {
  local mgr="$1"
  if command -v node >/dev/null 2>&1 && [[ "$(node --version)" == v${NODE_MAJOR}.* ]]; then
    log "Node.js 主版本 ${NODE_MAJOR} 已存在，跳过。"
    return
  fi
  log "安装 Node.js ${NODE_MAJOR} ..."
  case "$mgr" in
    apt)
      curl -fsSL "https://deb.nodesource.com/setup_${NODE_MAJOR}.x" | bash -
      apt-get install -y nodejs
      ;;
    dnf|yum)
      curl -fsSL "https://rpm.nodesource.com/setup_${NODE_MAJOR}.x" | bash -
      "$mgr" install -y nodejs
      ;;
    *)
      echo "未检测到受支持的包管理器，无法安装 Node.js。" >&2
      return
      ;;
  esac
  npm install -g pnpm
}

configure_shell_profile() {
  log "写入 ${DEV_USER} 的 shell 配置 ..."
  cat >>"${DEV_HOME}/.bashrc" <<'EOF'
# >>> lingcdn remote-dev >>>
export GOPATH="$HOME/go"
export PATH="$PATH:$GOPATH/bin:/usr/local/go/bin"
export LINGCDN_HOME="$HOME/workspace/Lingcdn系统开发"
alias gs='git status -sb'
alias gp='git pull --rebase'
# <<< lingcdn remote-dev <<<
EOF
  chown "${DEV_USER}:${DEV_USER}" "${DEV_HOME}/.bashrc"
}

prepare_workspace() {
  local ws="${DEV_HOME}/workspace"
  mkdir -p "${ws}"
  chown -R "${DEV_USER}:${DEV_USER}" "${ws}"
  log "工作目录准备完成：${ws}"
}

main() {
  ensure_user
  local mgr
  mgr="$(detect_pkg_manager)"
  install_common_packages "${mgr}"
  setup_timezone
  install_go
  install_node "${mgr}"
  configure_shell_profile
  prepare_workspace
  log "初始化完成。请重新登录 ${DEV_USER} 以加载新的环境变量。"
}

main "$@"
