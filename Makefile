include .env
export

start:
	go run *.go

up:
	@docker-compose up -d

down:
	@docker-compose down
