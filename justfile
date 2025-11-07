default:
	just --list
frontend_start_dev:
	cd frontend && npm run dev
backend_start_dev:
	FRONTEND_DEV_SERVER="http://localhost:7779" go run -v cmd/dev/*.go
labs_start_dev:
	go run -v backend/labs/labs.go
run_tests:
	cd tmp && rm -rf data.db && go run ../tests/*.go

play:
	FRONTEND_DEV_SERVER="http://localhost:7779" go run cli/play/play.go

build:
	cd frontend && npm run build
	go build -v -o ./tmp/potatoverse ./main.go

build_frontend:
	cd frontend && npm run build

build_backend:
	go build -v -o tmp/potatoverse ./main.go

copy_binary:
	cp tmp/potatoverse ~/go/bin

run_cli:
	cd tmp && FRONTEND_DEV_SERVER="http://localhost:7779"  go run ../main.go 


run_http_server:
	cd contrib && python3 -m http.server 8080

start_test_server:
	FRONTEND_DEV_SERVER="http://localhost:7779" TURNIX_DEV_MODE="true" TURNIX_DEV_ABC_SERVER="http://localhost:8080" go run -v cmd/dev/main.go

citest:
	./ci.sh