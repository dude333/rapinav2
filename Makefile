BUILDDIR     := cmd/*.go
SOURCEDIR    := .
SOURCES      := $(shell find "$(SOURCEDIR)" \( -name '*.css' -o -name '*.js' -o -name '*.html' -o -name '*.go' \) -a -not -name '*_test.go')


BINARYDIR    := .
BINARY       := rapinav2
WINBINARY    := rapinav2.exe

VERSION      := $(shell git describe --tags --always)
BUILD_TIME   := $(shell date +%F)

export GO111MODULE=on

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS      := -ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD_TIME}"

# Auto-detect compiler toolchain
COMPILER_CMD :=
ifneq ($(shell which gcc 2>/dev/null),)
	COMPILER_CMD = CC=gcc CXX=g++
else ifneq ($(shell which zig 2>/dev/null),)
	COMPILER_CMD = CC="zig cc" CXX="zig c++"
endif

.DEFAULT_GOAL:= $(BINARY)

$(BINARY): $(SOURCES)
	$(COMPILER_CMD) CGO_ENABLED=1 go build $(LDFLAGS) -o $(BINARYDIR)/$(BINARY) $(BUILDDIR)

win: $(SOURCES)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc-win32 CXX=x86_64-w64-mingw32-cpp-win32 CGO_LDFLAGS="-lssp -w" go build $(LDFLAGS) -o $(BINARYDIR)/$(WINBINARY) $(BUILDDIR)

osx: $(SOURCES)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 CC=o64-clang CXX=o64-clang++ CGO_LDFLAGS="-w" go build $(LDFLAGS) -o $(BINARYDIR) $(BUILDDIR)

clean:
	rm -f $(BINARY) $(WINBINARY)

.PHONY: run win osx clean
