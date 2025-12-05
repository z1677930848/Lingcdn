#!/usr/bin/env bash
set -euo pipefail

# ---- Config (override via env) ----
REPO_PATH="${REPO_PATH:-/www/wwwroot/docmost-main}"
APP_URL="${APP_URL:-https://docs.lingcdn.cloud}"
DB_NAME="${DB_NAME:-docmost}"
DB_USER="${DB_USER:-docmost}"
DB_PASS="${DB_PASS:-ChangeMe123!}"   # avoid quotes/specials
DB_HOST="${DB_HOST:-127.0.0.1}"
DB_PORT="${DB_PORT:-5432}"
REDIS_URL="${REDIS_URL:-redis://127.0.0.1:6379}"
SKIP_BUILD="${SKIP_BUILD:-false}"
SKIP_MIGRATE="${SKIP_MIGRATE:-false}"
NODE_MAJOR="${NODE_MAJOR:-18}"
GIT_URL="${GIT_URL:-https://github.com/docmost/docmost.git}"

# ---- Require root (for apt/systemctl/psql) ----
if [ "${EUID:-$(id -u)}" -ne 0 ]; then
  echo "Run as root: sudo $0" >&2; exit 1
fi

# ---- Helpers ----
new_secret() { hexdump -vn32 -e '/1 "%02x"' /dev/urandom; }
run_psql() { sudo -u postgres psql -v ON_ERROR_STOP=1 "$@"; }

# ---- Base packages ----
apt-get update
apt-get install -y curl ca-certificates gnupg lsb-release software-properties-common build-essential
if ! command -v git >/dev/null; then
  apt-get install -y git
fi

# ---- Node & pnpm (skip if Node major already high enough) ----
install_node=true
if command -v node >/dev/null; then
  NODE_VER=$(node -v | sed 's/^v//')
  NODE_MAJOR_INST=${NODE_VER%%.*}
  if [ "$NODE_MAJOR_INST" -ge "$NODE_MAJOR" ]; then
    install_node=false
  fi
fi
if [ "$install_node" = true ]; then
  curl -fsSL "https://deb.nodesource.com/setup_${NODE_MAJOR}.x" | bash -
  apt-get install -y nodejs
fi
if ! command -v pnpm >/dev/null; then
  npm install -g pnpm
fi

# ---- PostgreSQL (skip install if psql exists) ----
if command -v psql >/dev/null; then
  echo "PostgreSQL detected, skip install"
  systemctl enable --now postgresql || true
else
  if [ ! -f /etc/apt/sources.list.d/pgdg.list ]; then
    echo "deb http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list
    curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc | gpg --dearmor -o /etc/apt/trusted.gpg.d/pgdg.gpg
    apt-get update
  fi
  apt-get install -y postgresql-16
  systemctl enable --now postgresql
fi

# ---- Redis (skip if running; otherwise fix user and install) ----
redis_ready=false
if command -v redis-server >/dev/null && systemctl is-active --quiet redis-server; then
  redis_ready=true
fi
if [ "$redis_ready" = true ]; then
  echo "Redis already running, skip install"
else
  if id redis >/dev/null 2>&1; then
    deluser redis 2>/dev/null || true
    delgroup redis 2>/dev/null || true
  fi
  adduser --system --group --no-create-home --home /var/lib/redis --shell /usr/sbin/nologin redis 2>/dev/null || true
  # ensure runtime/data/log dirs exist to avoid pidfile/dir errors
  mkdir -p /var/run/redis /var/lib/redis /var/log/redis
  chown -R redis:redis /var/run/redis /var/lib/redis /var/log/redis 2>/dev/null || true
  chmod 755 /var/run/redis
  chmod 750 /var/lib/redis
  apt-get install -y redis-server
  systemctl enable --now redis-server || systemctl restart redis-server
fi

# ---- Clone repo if missing ----
if [ ! -d "$REPO_PATH/.git" ]; then
  rm -rf "$REPO_PATH"
  git clone "$GIT_URL" "$REPO_PATH"
fi

# ---- Prepare repo ----
mkdir -p "$REPO_PATH"
cd "$REPO_PATH"

if [ ! -f package.json ]; then
  echo "Repo not found or clone failed in $REPO_PATH" >&2; exit 1
fi

# ---- .env setup ----
if [ ! -f .env ]; then cp .env.example .env; fi
SECRET=$(new_secret)
DATABASE_URL="postgresql://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}/${DB_NAME}?schema=public"
tmp_env=$(mktemp)
while IFS= read -r line; do
  line=$(printf '%s\n' "$line" | \
    sed -e "s/^APP_URL=.*/APP_URL=${APP_URL//\//\\/}/" \
        -e "s/^PORT=.*/PORT=3000/" \
        -e "s/^APP_SECRET=.*/APP_SECRET=${SECRET}/" \
        -e "s/^DATABASE_URL=.*/DATABASE_URL=${DATABASE_URL//\//\\/}/" \
        -e "s/^REDIS_URL=.*/REDIS_URL=${REDIS_URL//\//\\/}/")
  printf '%s\n' "$line" >> "$tmp_env"
done < .env
mv "$tmp_env" .env

# ---- Create DB user/db if missing ----
run_psql -tc "SELECT 1 FROM pg_roles WHERE rolname='${DB_USER}'" | grep -q 1 || run_psql -c "CREATE ROLE \"${DB_USER}\" LOGIN PASSWORD '${DB_PASS}';"
run_psql -tc "SELECT 1 FROM pg_database WHERE datname='${DB_NAME}'" | grep -q 1 || run_psql -c "CREATE DATABASE \"${DB_NAME}\" OWNER \"${DB_USER}\";"
run_psql -d postgres -c "GRANT ALL PRIVILEGES ON DATABASE \"${DB_NAME}\" TO \"${DB_USER}\";"

# ---- Install node deps ----
pnpm install --frozen-lockfile

# ---- Migrate DB ----
if [ "$SKIP_MIGRATE" != "true" ]; then
  pnpm --filter ./apps/server run migration:latest
fi

# ---- Build ----
if [ "$SKIP_BUILD" != "true" ]; then
  pnpm run build
fi

echo "Done! Start commands (run as normal user inside $REPO_PATH):"
echo "  pnpm run start    # HTTP service (uses PORT in .env)"
echo "  pnpm run collab   # collaboration process"
echo "Dev mode: pnpm run dev"
echo "ENV used:"
echo "  APP_URL=$APP_URL"
echo "  DATABASE_URL=$DATABASE_URL"
echo "  REDIS_URL=$REDIS_URL"
