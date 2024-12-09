.PHONY: build-bindings

PYTHON_PREFIX := $(shell python3-config --prefix)
PYTHON_VERSION := $(shell python3 -c "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')")
PYTHON_LIBNAME := python$(PYTHON_VERSION)

build-bindings:
	cd bindings && \
	GO111MODULE=on \
	CGO_ENABLED=1 \
	GOWORK=off \
	GOFLAGS=-mod=mod \
	CGO_LDFLAGS="-L$(PYTHON_PREFIX)/lib -l$(PYTHON_LIBNAME) $(shell python3-config --ldflags)" \
	CGO_CFLAGS="$(shell python3-config --cflags)" \
	gopy build -output=../../_binding -vm=python3 .