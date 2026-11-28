#!/usr/bin/env bash

# Управление dev-серверами проекта: БД (docker), back (go) и front (vite).
# Один скрипт и для человека, и для агента: идемпотентный start/stop/status.
# Серверы стартуют через setsid (отдельная сессия/группа процессов) и живут
# независимо от терминала; pgid-файлы и логи хранятся в ./.run.

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUN_DIR="$ROOT/.run"
LOG_DIR="$RUN_DIR/logs"
BACK_LOG="$LOG_DIR/back.log"
FRONT_LOG="$LOG_DIR/front.log"
BACK_PGID_FILE="$RUN_DIR/back.pgid"
FRONT_PGID_FILE="$RUN_DIR/front.pgid"
LOCK_FILE="$RUN_DIR/dev.lock"
DB_CONTAINER="klubok-postgres"

DB_TIMEOUT=30     # секунд на ожидание готовности postgres
BACK_TIMEOUT=180  # секунд на ожидание готовности back (первый go run -race компилируется долго)
FRONT_TIMEOUT=60  # секунд на ожидание готовности front
STOP_TIMEOUT=15   # секунд на graceful остановку перед SIGKILL

# --- вывод -------------------------------------------------------------

if [[ -t 1 ]]; then
	C_GREEN=$'\033[32m'; C_YELLOW=$'\033[33m'; C_RED=$'\033[31m'; C_BLUE=$'\033[34m'; C_RESET=$'\033[0m'
else
	C_GREEN=""; C_YELLOW=""; C_RED=""; C_BLUE=""; C_RESET=""
fi

info() { printf '%s\n' "${C_BLUE}==>${C_RESET} $*"; }
ok()   { printf '%s\n' "${C_GREEN}  ✓${C_RESET} $*"; }
warn() { printf '%s\n' "${C_YELLOW}  !${C_RESET} $*"; }
err()  { printf '%s\n' "${C_RED}  ✗${C_RESET} $*" >&2; }
die()  { err "$*"; exit 1; }

usage() {
	cat <<'EOF'
Управление dev-серверами проекта Klubok (БД + back + front).

Использование: bash scripts/dev.sh <команда> [аргументы]

Команды:
  start [all|back|front]    Запустить; по умолчанию all в порядке: БД -> back -> front
  stop [all|back|front|db]  Остановить; по умолчанию all: front, back и контейнер БД
  restart [all|back|front]  Перезапустить
  status                    Показать статус БД, back и front (exit 1, если что-то не запущено)
  logs [-f] [back|front]    Логи компонента: последние 200 строк; с -f — follow

Примеры:
  bash scripts/dev.sh start
  bash scripts/dev.sh restart back
  bash scripts/dev.sh logs -f back

Из корня проекта то же самое доступно через make:
  make dev / dev-stop / dev-restart / dev-status / dev-logs
  выборочный таргет — через ARGS: make dev ARGS=back, make dev-logs ARGS="-f back"
EOF
}

# --- окружение ---------------------------------------------------------

# Неинтерактивные оболочки могут не загружать профиль пользователя, из-за
# чего в PATH нет инструментов вроде go и bun. Добавляем типовые каталоги
# их установки; в обычном терминале это no-op.
ensure_user_path() {
	local dir
	for dir in "$HOME/.local/bin" "$HOME/.bun/bin"; do
		if [[ -d "$dir" && ":$PATH:" != *":$dir:"* ]]; then
			export PATH="$dir:$PATH"
		fi
	done
}

# Загружает переменные из back/.envrc (работает и без direnv).
# strict=strict — отсутствие файла фатально (нужно для запуска back).
load_env() {
	local strict="${1:-loose}"
	if [[ -f "$ROOT/back/.envrc" ]]; then
		set -a
		# shellcheck disable=SC1091
		source "$ROOT/back/.envrc"
		set +a
	elif [[ "$strict" == "strict" ]]; then
		die "Нет файла back/.envrc — скопируй его из back/.envrc.example или выполни make init в ./back"
	else
		warn "Нет back/.envrc — порты взяты по умолчанию"
	fi
	HTTP_PORT="${HTTP_PORT:-5172}"
	WEBSOCKET_PORT="${WEBSOCKET_PORT:-5171}"
	POSTGRES_PORT="${POSTGRES_PORT:-50432}"
	POSTGRES_USER="${POSTGRES_USER:-postgres}"
	POSTGRES_DB="${POSTGRES_DB:-postgres}"
	FRONT_PORT="${FRONT_PORT:-5173}"
}

# --- проверки ----------------------------------------------------------

# 0, если HTTP-сервер ответил успешным кодом.
http_ok() {
	curl -fs -o /dev/null --max-time 2 "$1"
}

# 0, если на TCP-порту кто-то слушает.
port_busy() {
	local port="$1"
	timeout 1 bash -c "exec 3<>/dev/tcp/127.0.0.1/$port" 2>/dev/null
}

back_live_url()   { printf 'http://127.0.0.1:%s/api/health/live' "$HTTP_PORT"; }
back_ready_url()  { printf 'http://127.0.0.1:%s/api/health/ready' "$HTTP_PORT"; }
front_root_url()  { printf 'http://127.0.0.1:%s/' "$FRONT_PORT"; }

# --- pgid-файлы --------------------------------------------------------

# 0, если файл есть и группа процессов жива.
pgid_alive() {
	local file="$1" pgid
	[[ -f "$file" ]] || return 1
	pgid="$(cat "$file")"
	[[ "$pgid" =~ ^[0-9]+$ ]] || return 1
	kill -0 -- "-$pgid" 2>/dev/null
}

# Убирает протухший pgid-файл.
drop_pgid_if_stale() {
	local file="$1"
	if [[ -f "$file" ]] && ! pgid_alive "$file"; then
		rm -f "$file"
	fi
}

# Запускает команду (аргументы после трёх первых) в новой сессии из каталога $1,
# пишет pgid в файл $2 и лог в $3. exec внутри под-оболочки гарантирует,
# что $! — это pid нового лидера сессии (новой группы процессов).
launch_detached() {
	local dir="$1" pgid_file="$2" log_file="$3"
	shift 3
	: >"$log_file"
	# 9>&- не даёт серверу унаследовать fd 9 с flock, иначе лок не освободится.
	( cd "$dir" && exec setsid "$@" </dev/null >>"$log_file" 2>&1 ) 9>&- &
	echo $! >"$pgid_file"
}

# --- БД ----------------------------------------------------------------

db_running() {
	[[ "$(docker inspect -f '{{.State.Running}}' "$DB_CONTAINER" 2>/dev/null || true)" == "true" ]]
}

db_ready() {
	docker exec "$DB_CONTAINER" pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB" >/dev/null 2>&1
}

start_db() {
	if db_running; then
		ok "БД: контейнер $DB_CONTAINER уже запущен"
	else
		if ! docker inspect "$DB_CONTAINER" >/dev/null 2>&1; then
			die "БД: контейнер $DB_CONTAINER не существует — выполни make init в каталоге ./back"
		fi
		info "БД: запускаю контейнер $DB_CONTAINER"
		docker start "$DB_CONTAINER" >/dev/null
	fi
	info "БД: жду готовности postgres (до ${DB_TIMEOUT}с)..."
	local waited=0
	until db_ready; do
		(( waited < DB_TIMEOUT )) || die "БД: postgres не готов за ${DB_TIMEOUT}с — смотри docker logs $DB_CONTAINER"
		sleep 1
		waited=$((waited + 1))
	done
	ok "БД: postgres готов (порт $POSTGRES_PORT)"
}

stop_db() {
	if db_running; then
		info "БД: останавливаю контейнер $DB_CONTAINER..."
		docker stop "$DB_CONTAINER" >/dev/null
		ok "БД: контейнер остановлен"
	else
		info "БД: контейнер не запущен"
	fi
}

# --- back / front ------------------------------------------------------

wait_ready() {
	local name="$1" url="$2" pgid_file="$3" log_file="$4" timeout="$5"
	local waited=0
	until http_ok "$url"; do
		if ! pgid_alive "$pgid_file"; then
			err "$name: процесс завершился с ошибкой, последние строки лога:"
			tail -n 40 "$log_file" >&2 || true
			exit 1
		fi
		(( waited < timeout )) || {
			err "$name: не дождался готовности за ${timeout}с, последние строки лога:"
			tail -n 40 "$log_file" >&2 || true
			exit 1
		}
		sleep 2
		waited=$((waited + 2))
	done
}

start_back() {
	drop_pgid_if_stale "$BACK_PGID_FILE"
	if pgid_alive "$BACK_PGID_FILE"; then
		info "back: процесс уже запущен, жду health..."
	else
		local port
		for port in "$HTTP_PORT" "$WEBSOCKET_PORT"; do
			if port_busy "$port"; then
				die "back: порт $port занят посторонним процессом — скрипт ничего не убивает, освободи порт вручную"
			fi
		done
		load_env strict
		info "back: запускаю (make run, лог: $BACK_LOG)"
		launch_detached "$ROOT/back" "$BACK_PGID_FILE" "$BACK_LOG" make run
	fi
	wait_ready "back" "$(back_live_url)" "$BACK_PGID_FILE" "$BACK_LOG" "$BACK_TIMEOUT"
	ok "back: запущен (pgid $(cat "$BACK_PGID_FILE"), порт $HTTP_PORT)"
	if ! http_ok "$(back_ready_url)"; then
		warn "back: /api/health/ready вернул ошибку — БД недоступна"
	fi
}

start_front() {
	drop_pgid_if_stale "$FRONT_PGID_FILE"
	if pgid_alive "$FRONT_PGID_FILE"; then
		info "front: процесс уже запущен, жду ответа..."
	elif port_busy "$FRONT_PORT"; then
		die "front: порт $FRONT_PORT занят посторонним процессом — скрипт ничего не убивает, освободи порт вручную"
	else
		info "front: запускаю (make run, лог: $FRONT_LOG)"
		launch_detached "$ROOT/front" "$FRONT_PGID_FILE" "$FRONT_LOG" make run
	fi
	wait_ready "front" "$(front_root_url)" "$FRONT_PGID_FILE" "$FRONT_LOG" "$FRONT_TIMEOUT"
	ok "front: запущен (pgid $(cat "$FRONT_PGID_FILE"), порт $FRONT_PORT)"
}

stop_proc() {
	local name="$1" pgid_file="$2"
	drop_pgid_if_stale "$pgid_file"
	if ! pgid_alive "$pgid_file"; then
		info "$name: не запущен"
		return 0
	fi
	local pgid
	pgid="$(cat "$pgid_file")"
	info "$name: останавливаю (pgid $pgid, SIGTERM)..."
	kill -TERM -- "-$pgid" 2>/dev/null || true
	local waited=0
	while pgid_alive "$pgid_file" && (( waited < STOP_TIMEOUT )); do
		sleep 1
		waited=$((waited + 1))
	done
	if pgid_alive "$pgid_file"; then
		warn "$name: не завершился за ${STOP_TIMEOUT}с — отправляю SIGKILL"
		kill -KILL -- "-$pgid" 2>/dev/null || true
		sleep 1
	fi
	rm -f "$pgid_file"
	ok "$name: остановлен"
}

# --- команды -----------------------------------------------------------

print_urls() {
	printf '\n'
	ok "Всё запущено:"
	printf '   Приложение:  http://localhost:%s\n' "$FRONT_PORT"
	printf '   Диаграмма:   http://localhost:%s%s\n' "$FRONT_PORT"
	printf '   API:         http://localhost:%s/api\n' "$HTTP_PORT"
}

cmd_start() {
	local target="${1:-all}"
	# Окружение нужно сразу: pg_isready для БД и порты для проверок.
	load_env loose
	case "$target" in
		all)   start_db; start_back; start_front ;;
		back)  start_db; start_back ;;
		front) start_front ;;
		*)     die "Неизвестная цель для start: $target (допустимо: all|back|front)" ;;
	esac
	print_urls
}

cmd_stop() {
	local target="${1:-all}"
	case "$target" in
		all)   stop_proc "front" "$FRONT_PGID_FILE"; stop_proc "back" "$BACK_PGID_FILE"; stop_db ;;
		back)  stop_proc "back" "$BACK_PGID_FILE" ;;
		front) stop_proc "front" "$FRONT_PGID_FILE" ;;
		db)    stop_db ;;
		*)     die "Неизвестная цель для stop: $target (допустимо: all|back|front|db)" ;;
	esac
}

cmd_restart() {
	local target="${1:-all}"
	case "$target" in
		all|back|front) cmd_stop "$target"; cmd_start "$target" ;;
		*)              die "Неизвестная цель для restart: $target (допустимо: all|back|front)" ;;
	esac
}

cmd_status() {
	load_env loose
	local code=0

	if db_running; then
		if db_ready; then
			ok "БД:    запущена и готова (порт $POSTGRES_PORT)"
		else
			warn "БД:    контейнер запущен, но postgres не готов"; code=1
		fi
	else
		warn "БД:    не запущена"; code=1
	fi

	drop_pgid_if_stale "$BACK_PGID_FILE"
	if pgid_alive "$BACK_PGID_FILE" && http_ok "$(back_live_url)"; then
		if http_ok "$(back_ready_url)"; then
			ok "back:  запущен (порт $HTTP_PORT, БД доступна)"
		else
			warn "back:  запущен (порт $HTTP_PORT), но /api/health/ready=503 — БД недоступна"; code=1
		fi
	elif port_busy "$HTTP_PORT"; then
		warn "back:  порт $HTTP_PORT занят посторонним процессом"; code=1
	else
		warn "back:  не запущен"; code=1
	fi

	drop_pgid_if_stale "$FRONT_PGID_FILE"
	if pgid_alive "$FRONT_PGID_FILE" && http_ok "$(front_root_url)"; then
		ok "front: запущен (порт $FRONT_PORT)"
	elif port_busy "$FRONT_PORT"; then
		warn "front: порт $FRONT_PORT занят посторонним процессом"; code=1
	else
		warn "front: не запущен"; code=1
	fi

	exit "$code"
}

cmd_logs() {
	local follow=0 target="back" arg
	for arg in "$@"; do
		case "$arg" in
			-f|--follow) follow=1 ;;
			back|front)  target="$arg" ;;
			*)           die "Неизвестный аргумент logs: $arg" ;;
		esac
	done
	local log_file="$BACK_LOG"
	[[ "$target" == "front" ]] && log_file="$FRONT_LOG"
	[[ -f "$log_file" ]] || die "Лога ещё нет: $log_file (запусти start)"
	if (( follow )); then
		tail -n 200 -f "$log_file"
	else
		tail -n 200 "$log_file"
	fi
}

# --- точка входа -------------------------------------------------------

acquire_lock() {
	mkdir -p "$RUN_DIR"
	exec 9>"$LOCK_FILE"
	flock -n 9 || die "Другой экземпляр скрипта уже выполняет start/stop — подожди его завершения"
}

main() {
	ensure_user_path
	if (( $# < 1 )); then
		usage
		exit 1
	fi
	local cmd="$1"
	shift
	case "$cmd" in
		start|stop|restart)
			mkdir -p "$LOG_DIR"
			acquire_lock
			"cmd_$cmd" "$@"
			;;
		status)
			mkdir -p "$LOG_DIR"
			cmd_status "$@"
			;;
		logs)
			cmd_logs "$@"
			;;
		help|-h|--help)
			usage
			;;
		*)
			usage >&2
			die "Неизвестная команда: $cmd"
			;;
	esac
}

main "$@"
