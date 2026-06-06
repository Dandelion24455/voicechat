.PHONY: infra up down build run clean

infra:
	colima start --cpu 2 --memory 4
	docker compose up -d

down:
	docker compose down

build:
	cd server && GOPROXY=https://goproxy.cn,direct go build -o /tmp/voicechat-server .

run:
	cd server && go run .

clean:
	docker compose down -v
	rm -f /tmp/voicechat-server
