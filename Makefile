BUILDDIR     = cmd/*
SOURCEDIR    = .
SOURCES     := $(shell find $(SOURCEDIR) -name '*.go' | grep -v "_test.go")

BINARYDIR=.
BINARY=rapinav2
WINBINARY=rapinav2.exe

VERSION=`git describe --tags --always`
BUILD_TIME=`date +%F`

export GO111MODULE=on

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-w -s -X main.version=${VERSION} -X main.build=${BUILD_TIME}"

.DEFAULT_GOAL: $(BINARY)

$(BINARY): $(SOURCES)
	go build ${LDFLAGS} -o $(BINARYDIR)/$(BINARY) $(BUILDDIR)

win: $(wildcard *.go)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc-win32 CXX=x86_64-w64-mingw32-cpp-win32 CGO_LDFLAGS="-lssp -w"  go build ${LDFLAGS} -o ${BINARYDIR}/$(WINBINARY) $(BUILDDIR)

osx:  $(SOURCES)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 CC=o64-clang CXX=o64-clang++ CGO_LDFLAGS="-w" go build ${LDFLAGS} -o ${BINARYDIR} $(BUILDDIR)

clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

.PHONY: run win clean
