#!/usr/bin/env bash

set -Eeuo pipefail

readonly SERVICE_DIR="/home/work/service/nutriLens"
readonly CODE_DIR="/home/work/codeBase/nutriLens"
readonly SERVER_DIR="${CODE_DIR}/server"
readonly CONFIG_FILE="${SERVICE_DIR}/config.yaml"
readonly CONTAINER_NAME="nutrilens-server"
readonly IMAGE_NAME="nutrilens-server"
readonly UPLOAD_VOLUME="server_uploads"
readonly CONTAINER_CONFIG="/app/config/config.yaml"
readonly GOPROXY_VALUE="https://goproxy.cn,direct"
readonly HEALTH_URL="http://127.0.0.1:8080/health"

log() {
  printf '[deploy] %s\n' "$*"
}

fail() {
  printf '[deploy] ERROR: %s\n' "$*" >&2
  exit 1
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || fail "缺少命令：$1"
}

container_exists() {
  docker container inspect "${CONTAINER_NAME}" >/dev/null 2>&1
}

show_status() {
  docker ps -a --filter "name=^/${CONTAINER_NAME}$" \
    --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}'
}

show_logs() {
  log "最近 50 行服务日志："
  docker logs --tail 50 "${CONTAINER_NAME}" 2>&1 || true
}

validate_build_inputs() {
  require_command git
  require_command docker
  require_command curl
  [[ -d "${CODE_DIR}/.git" ]] || fail "代码仓库不存在：${CODE_DIR}"
  [[ -f "${SERVER_DIR}/Dockerfile" ]] || fail "Dockerfile 不存在：${SERVER_DIR}/Dockerfile"
  [[ -f "${CONFIG_FILE}" ]] || fail "配置文件不存在或不是普通文件：${CONFIG_FILE}"
  [[ -r "${CONFIG_FILE}" ]] || fail "配置文件不可读：${CONFIG_FILE}"
}

wait_until_healthy() {
  local attempt
  for attempt in {1..20}; do
    if [[ "$(docker inspect --format '{{.State.Running}}' "${CONTAINER_NAME}" 2>/dev/null || true)" == "true" ]]; then
      if curl --fail --silent --show-error --max-time 2 "${HEALTH_URL}" >/dev/null 2>&1; then
        return 0
      fi
    fi
    sleep 1
  done
  return 1
}

restore_previous_container() {
  local backup_name="$1"
  log "新容器启动失败，正在恢复旧容器"
  if container_exists; then
    docker rm -f "${CONTAINER_NAME}" >/dev/null 2>&1 || true
  fi
  if docker container inspect "${backup_name}" >/dev/null 2>&1; then
    docker rename "${backup_name}" "${CONTAINER_NAME}"
    docker start "${CONTAINER_NAME}" >/dev/null
  fi
}

build_and_deploy() {
  validate_build_inputs

  local timestamp image backup_name=""
  timestamp="$(date '+%Y-%m-%d-%H%M%S')"
  image="${IMAGE_NAME}:${timestamp}"

  log "拉取最新代码：${CODE_DIR}"
  git -C "${CODE_DIR}" pull --ff-only

  log "构建镜像：${image}"
  docker build \
    --build-arg "GOPROXY=${GOPROXY_VALUE}" \
    --tag "${image}" \
    "${SERVER_DIR}"

  log "验证容器内 app 用户可以读取挂载配置"
  docker run --rm \
    --entrypoint /bin/sh \
    --volume "${CONFIG_FILE}:${CONTAINER_CONFIG}:ro" \
    "${image}" \
    -c "test -f '${CONTAINER_CONFIG}' && test -r '${CONTAINER_CONFIG}'" \
    || fail "容器无法读取 ${CONFIG_FILE}，请检查文件类型和权限"

  if container_exists; then
    backup_name="${CONTAINER_NAME}-rollback-${timestamp}"
    log "停止旧容器并暂存为：${backup_name}"
    docker stop "${CONTAINER_NAME}" >/dev/null
    docker rename "${CONTAINER_NAME}" "${backup_name}"
  fi

  log "启动新容器：${CONTAINER_NAME}"
  if ! docker run -d \
    --name "${CONTAINER_NAME}" \
    --restart unless-stopped \
    --network host \
    --env APP_ENV=production \
    --env "CONFIG_PATH=${CONTAINER_CONFIG}" \
    --volume "${CONFIG_FILE}:${CONTAINER_CONFIG}:ro" \
    --volume "${UPLOAD_VOLUME}:/app/uploads" \
    "${image}" >/dev/null; then
    restore_previous_container "${backup_name}"
    fail "新容器创建失败"
  fi

  if ! wait_until_healthy; then
    show_logs
    restore_previous_container "${backup_name}"
    fail "新容器未能稳定运行，已尝试恢复旧版本"
  fi

  if [[ -n "${backup_name}" ]]; then
    docker rm "${backup_name}" >/dev/null
  fi

  log "部署完成：${image}"
  show_status
  show_logs
}

restart_service() {
  require_command docker
  require_command curl
  container_exists || fail "容器不存在：${CONTAINER_NAME}"
  log "重启容器：${CONTAINER_NAME}"
  docker restart "${CONTAINER_NAME}" >/dev/null
  wait_until_healthy || {
    show_logs
    fail "容器重启失败"
  }
  show_status
  show_logs
}

stop_service() {
  require_command docker
  if ! container_exists; then
    log "容器不存在，无需停止：${CONTAINER_NAME}"
    return
  fi
  log "停止容器：${CONTAINER_NAME}"
  docker stop "${CONTAINER_NAME}" >/dev/null
  show_status
}

usage() {
  cat <<'EOF'
用法：
  ./deploy.sh build     拉取代码、构建时间版本镜像并部署
  ./deploy.sh restart   重启当前容器
  ./deploy.sh stop      停止当前容器
EOF
}

main() {
  case "${1:-}" in
    build)
      build_and_deploy
      ;;
    restart)
      restart_service
      ;;
    stop)
      stop_service
      ;;
    -h|--help|help)
      usage
      ;;
    *)
      usage >&2
      exit 2
      ;;
  esac
}

main "$@"
