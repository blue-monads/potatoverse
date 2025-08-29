default:
	just --list
start_dev_frontend:
	cd frontend && npm run dev
start_dev_backend:
	FRONTEND_DEV_SERVER="http://localhost:7779" go run -v cmd/dev/*.go
run_tests:
	cd tmp && rm -rf data.db && go run ../tests/*.go

play:
	FRONTEND_DEV_SERVER="http://localhost:7779" go run cli/play/play.go

build:
	cd frontend && npm run build
	go build -v -o tmp/turnix ./main.go

build_frontend:
	cd frontend && npm run build

build_backend:
	go build -v -o tmp/turnix ./main.go

copy_binary:
	cp tmp/turnix ~/go/bin

run_cli:
	cd tmp && FRONTEND_DEV_SERVER="http://localhost:7779"  go run ../main.go 


run_http_server:
	cd contrib && python3 -m http.server 8080

start_test_server:
	FRONTEND_DEV_SERVER="http://localhost:7779" TURNIX_DEV_MODE="true" TURNIX_DEV_ABC_SERVER="http://localhost:8080" go run -v cmd/dev/main.go

citest:
	./ci.sh