build-frontend:
	cd frontend && npm install && npm run build

build-backend:
	mkdir -p bin
	cd backend && CGO_ENABLED=0 go build -a -tags netgo,osusergo -ldflags '-w -extldflags "-static"' -trimpath -o seif && mv seif ../bin

build-backend-nostatic:
	mkdir -p bin
	cd backend && CGO_ENABLED=0 go build -trimpath -o seif && mv seif ../bin

build:
	make build-frontend
	make build-backend

zbuild:
	make build
	strip bin/seif
	upx --ultra-brute bin/seif

run-devel:
	make build-frontend
	cd backend && go run main.go --db seif.db

update:
	cd frontend && npm update
	cd backend && go get -u && go mod tidy

clean:
	rm -rf frontend/node_modules
	rm -rf bin
	rm -f backend/seif.db
	rm -f seif.db