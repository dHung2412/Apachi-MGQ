.PHONY: run build test tidy clean

run:
	@echo "Starting server..."
	go run cmd/server/main.go

build:
	@echo "Building application..."
	go build -o bin/server cmd/server/main.go

test:
	@echo "Running tests..."
	go test -v ./...

tidy:
	@echo "Tidy up dependencies..."
	go mod tidy
	go mod download

clean:
	@echo "Cleaning up..."
	rm -rf bin/
	go clean

docker-up:
	@echo "Starting Docker services..."
	docker-compose up -d

docker-down:
	@echo "Stopping Docker services..."
	docker-compose down

docker-logs:
	docker-compose logs -f

deps:
	@echo "Installing dependencies..."
	go get github.com/labstack/echo/v4
	go get github.com/golang-jwt/jwt/v5
	go get golang.org/x/crypto/bcrypt
	go get github.com/joho/godotenv
	go get gorm.io/gorm
	go get gorm.io/driver/postgres