BUILDDIR     = .
SOURCEDIR    = .
SOURCES     := $(shell find $(SOURCEDIR) -name '*.go' | grep -v "_test.go")

FE_DIR       = frontend
FE_SOURCES  := $(shell find $(FE_DIR)/src -type f) $(FE_DIR)/package.json
FE_BUILD     = $(FE_DIR)/public/build/bundle.js

BINARYDIR=.
BINARY=rapina
WINBINARY=rapina.exe

VERSION=`git describe --tags --always`
BUILD_TIME=`date +%F`

export GO111MODULE=on
export GOFLAGS=-mod=vendor

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD_TIME}"

.DEFAULT_GOAL: $(BINARY)

$(BINARY): $(SOURCES) $(FE_BUILD)
	go build ${LDFLAGS} -o $(BINARYDIR)/$(BINARY) $(BUILDDIR)

run: $(FE_BUILD)
	go run rapina.go servidor

$(FE_BUILD): $(FE_SOURCES)
	cd $(FE_DIR) && pnpm run build || npm run build

frontend: $(FE_BUILD)

win: $(wildcard *.go) $(FE_BUILD)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc-win32 CXX=x86_64-w64-mingw32-cpp-win32 CGO_LDFLAGS="-lssp -w"  go build ${LDFLAGS} -o ${BINARYDIR}/$(WINBINARY) $(BUILDDIR)

osx:  $(SOURCES)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 CC=o64-clang CXX=o64-clang++ CGO_LDFLAGS="-w" go build ${LDFLAGS} -o ${BINARYDIR} $(BUILDDIR)

clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

.PHONY: run win clean
