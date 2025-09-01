build:
	@echo 'Building...'
	@go build -o=./bin/api ./cmd/api
	@echo 'Done'

run:build
	@./bin/api
