BINARY_NAME_CLI=pics
BINARY_NAME_UI=pics-ui
GORELEASER := go run github.com/goreleaser/goreleaser/v2@latest

# Binary versions for UI embedding
EXIFTOOL_VERSION=13.45
JPEGOPTIM_VERSION=1.5.6

.PHONY: build
build: build-cli

.PHONY: build-cli
build-cli:
	go build -o $(BINARY_NAME_CLI) ./apps/cli

.PHONY: build-ui
build-ui:
	cd apps/ui && wails build

.PHONY: build-all
build-all: build-cli build-ui

.PHONY: dev-ui
dev-ui:
	cd apps/ui && wails dev

.PHONY: run
# Example: make run ARGS="parse /source /target"
run:
	go run ./apps/cli $(ARGS)

.PHONY: test
test:
	go test -v ./...

.PHONY: clean
clean:
	rm -f $(BINARY_NAME_CLI)
	rm -rf apps/ui/build
	rm -rf dist/

.PHONY: download-binaries-windows
download-binaries-windows:
	@echo "Downloading Windows binaries..."
	@mkdir -p build/resources/windows
	@TMPDIR=$$(mktemp -d /tmp/pics-binaries-windows.XXXXXX) && \
		cd $$TMPDIR && \
		curl -L -o exiftool.zip "https://sourceforge.net/projects/exiftool/files/exiftool-$(EXIFTOOL_VERSION)_64.zip/download" && \
		unzip -q exiftool.zip && \
		chmod -R u+w . && \
		cp "exiftool-$(EXIFTOOL_VERSION)_64/exiftool(-k).exe" $(CURDIR)/build/resources/windows/exiftool.exe && \
		curl -L -o jpegoptim.zip "https://github.com/XhmikosR/jpegoptim-windows/releases/download/$(JPEGOPTIM_VERSION)-rel1/jpegoptim-$(JPEGOPTIM_VERSION)-rel1-win64-msvc-2022-mozjpeg331-static-ltcg.zip" && \
		unzip -q jpegoptim.zip && \
		chmod -R u+w . && \
		cp jpegoptim-*/jpegoptim.exe $(CURDIR)/build/resources/windows/jpegoptim.exe && \
		cd /tmp && rm -rf $$TMPDIR
	@echo "✓ Windows binaries downloaded to build/resources/windows/"

.PHONY: download-binaries-darwin
download-binaries-darwin:
	@echo "Downloading macOS binaries..."
	@mkdir -p build/resources/darwin
	@TMPDIR=$$(mktemp -d /tmp/pics-binaries-darwin.XXXXXX) && \
		cd $$TMPDIR && \
		curl -L -o exiftool.tar.gz "https://exiftool.org/Image-ExifTool-$(EXIFTOOL_VERSION).tar.gz" && \
		tar -xzf exiftool.tar.gz && \
		chmod -R u+w . && \
		cp Image-ExifTool-$(EXIFTOOL_VERSION)/exiftool $(CURDIR)/build/resources/darwin/exiftool && \
		chmod +x $(CURDIR)/build/resources/darwin/exiftool && \
		curl -L -o jpegoptim.zip "https://github.com/tjko/jpegoptim/releases/download/v$(JPEGOPTIM_VERSION)/jpegoptim-$(JPEGOPTIM_VERSION)-x64-osx.zip" && \
		unzip -q jpegoptim.zip && \
		chmod -R u+w . && \
		cp jpegoptim $(CURDIR)/build/resources/darwin/jpegoptim && \
		chmod +x $(CURDIR)/build/resources/darwin/jpegoptim && \
		cd /tmp && rm -rf $$TMPDIR
	@echo "✓ macOS binaries downloaded to build/resources/darwin/"

.PHONY: download-binaries-linux
download-binaries-linux:
	@echo "Downloading Linux binaries..."
	@mkdir -p build/resources/linux
	@TMPDIR=$$(mktemp -d /tmp/pics-binaries-linux.XXXXXX) && \
		cd $$TMPDIR && \
		curl -L -o exiftool.tar.gz "https://exiftool.org/Image-ExifTool-$(EXIFTOOL_VERSION).tar.gz" && \
		tar -xzf exiftool.tar.gz && \
		chmod -R u+w . && \
		cp Image-ExifTool-$(EXIFTOOL_VERSION)/exiftool $(CURDIR)/build/resources/linux/exiftool && \
		chmod +x $(CURDIR)/build/resources/linux/exiftool && \
		curl -L -o jpegoptim.zip "https://github.com/tjko/jpegoptim/releases/download/v$(JPEGOPTIM_VERSION)/jpegoptim-$(JPEGOPTIM_VERSION)-x64-linux.zip" && \
		unzip -q jpegoptim.zip && \
		chmod -R u+w . && \
		cp jpegoptim $(CURDIR)/build/resources/linux/jpegoptim && \
		chmod +x $(CURDIR)/build/resources/linux/jpegoptim && \
		cd /tmp && rm -rf $$TMPDIR
	@echo "✓ Linux binaries downloaded to build/resources/linux/"

.PHONY: download-binaries
download-binaries: download-binaries-windows download-binaries-darwin download-binaries-linux
	@echo ""
	@echo "All binaries downloaded successfully!"
	@echo "Windows: build/resources/windows/{exiftool.exe, jpegoptim.exe}"
	@echo "macOS:   build/resources/darwin/{exiftool, jpegoptim}"
	@echo "Linux:   build/resources/linux/{exiftool, jpegoptim}"

.PHONY: release-snapshot
release-snapshot:
	$(GORELEASER) release --snapshot --clean

.PHONY: release-test
release-test:
	$(GORELEASER) check

# Infrastructure targets
.PHONY: infra-deploy
infra-deploy:
	$(MAKE) -C infra deploy

.PHONY: infra-empty-bucket
infra-empty-bucket:
	$(MAKE) -C infra empty-bucket

.PHONY: infra-delete
infra-delete:
	$(MAKE) -C infra delete

.PHONY: infra-status
infra-status:
	$(MAKE) -C infra status

.PHONY: infra-outputs
infra-outputs:
	$(MAKE) -C infra outputs

.PHONY: infra-bucket-name
infra-bucket-name:
	$(MAKE) -C infra bucket-name

.PHONY: infra-validate
infra-validate:
	$(MAKE) -C infra validate

.PHONY: infra-help
infra-help:
	$(MAKE) -C infra help
