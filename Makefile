APP_NAME := scdl
BUILD_DIR := bin
VERSION := $(shell git describe --tags 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=${VERSION} -s -w"

.PHONY: check
check: tidy fmt lint test build

.PHONY: build
build:
	@go build ${LDFLAGS} ./cmd/scdl

.PHONY: test
test:
	@go test -v ./...

.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: lint
lint:
	@golangci-lint run

.PHONY: clean
clean:
	@rm -rf ${BUILD_DIR}

# Helper to build for a specific os/arch
# Usage: make build-os-arch OS=linux ARCH=amd64
.PHONY: build-os-arch
build-os-arch:
	@echo "Building ${APP_NAME} ${VERSION} for ${OS}/${ARCH}..."
	@mkdir -p ${BUILD_DIR}
	@if [ "${OS}" = "windows" ]; then \
		GOOS=${OS} GOARCH=${ARCH} go build ${LDFLAGS} -o ${BUILD_DIR}/${APP_NAME}.exe ./cmd/scdl; \
		zip -j ${BUILD_DIR}/${APP_NAME}_${VERSION}_${OS}-${ARCH}.zip ${BUILD_DIR}/${APP_NAME}.exe; \
		rm ${BUILD_DIR}/${APP_NAME}.exe; \
	else \
		GOOS=${OS} GOARCH=${ARCH} go build ${LDFLAGS} -o ${BUILD_DIR}/${APP_NAME} ./cmd/scdl; \
		tar -czf ${BUILD_DIR}/${APP_NAME}_${VERSION}_${OS}-${ARCH}.tar.gz -C ${BUILD_DIR} ${APP_NAME}; \
		rm ${BUILD_DIR}/${APP_NAME}; \
	fi

.PHONY: release
release: clean
	@$(MAKE) build-os-arch OS=darwin ARCH=amd64
	@$(MAKE) build-os-arch OS=darwin ARCH=arm64
	@$(MAKE) build-os-arch OS=linux ARCH=386
	@$(MAKE) build-os-arch OS=linux ARCH=amd64
	@$(MAKE) build-os-arch OS=linux ARCH=arm64
	@$(MAKE) build-os-arch OS=windows ARCH=386
	@$(MAKE) build-os-arch OS=windows ARCH=amd64
	@$(MAKE) build-os-arch OS=windows ARCH=arm64
	@echo "Release builds created in ${BUILD_DIR}/"
