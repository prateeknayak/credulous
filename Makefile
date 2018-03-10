# This option defines which mock configuration to use -- see /etc/mock for
# the available configuration files for your system.
MOCK_CONFIG=epel-6-x86_64
SHELL=/bin/bash
DIST=$(shell grep "config_opts.*dist.*" /etc/mock/$(MOCK_CONFIG).cfg | awk '{ print $$3 }' | cut -f2 -d\' )

SRCS=$(shell ls -1 *.go | grep -v _test.go ) bash/credulous.bash_completion \
	doc/credulous.md bash/credulous.sh scripts/libgit2.pc-rhel
TESTS=credulous_test.go credentials_test.go crypto_test.go git_test.go \
	testdata/testkey testdata/testkey.pub testdata/credential.json testdata/newcreds.json

DOC=doc/credulous.md
MAN=doc/credulous.1
SPEC=rpm/credulous.spec
SPEC_TMPL=rpm/credulous.spec.tmpl
NAME=$(shell grep '^Name:' $(SPEC_TMPL) | awk '{ print $$2 }' )
# Because we run under sudo, environment variables don't make it through
BUILD_NR=$(shell cat travis_build_number)
ifeq ($(strip $(BUILD_NR)), )
BUILD_NR=unknown
endif
VERS=$(shell cat VERSION 2>/dev/null )
VERSION=$(VERS).$(BUILD_NR)
RELEASE=$(shell grep '^Release:' $(SPEC_TMPL) | awk '{ print $$2 }' | sed -e 's/%{?dist}/.$(DIST)/' )

MOCK_RESULT=/var/lib/mock/$(MOCK_CONFIG)/result

NVR=$(NAME)-$(VERSION)-$(RELEASE)
MOCK_SRPM=$(NVR).src.rpm
RPM=$(NVR).x86_64.rpm
TGZ=$(NAME)-$(VERSION).tar.gz

INSTALLABLES=credulous bash/credulous.bash_completion doc/credulous.1 bash/credulous.sh

.DEFAULT: all
.PHONY: debianpkg

all: mock

man: $(DOC)
	sed -e 's/==VERSION==/$(VERSION)/' $(DOC) | pandoc -s -w man - -o $(MAN)

osx_binaries: $(SRCS) $(TESTS)
	@echo "Building for OSX"
	go get -t
	go test
	go build

osx: man osx_binaries
	tar zcvf credulous-$(VERSION)-osx.tgz $(INSTALLABLES)

# This is a dirty hack for building on ubuntu build agents in Travis.
rpmbuild: sources
	@mkdir -p 	$(HOME)/rpmbuild/SOURCES \
			$(HOME)/rpmbuild/SRPMS \
			$(HOME)/rpmbuild/RPMS \
			$(HOME)/rpmbuild/SPECS \
			$(HOME)/rpmbuild/BUILD \
			$(HOME)/rpmbuild/BUILDROOT
	cp $(NAME)-$(VERSION).tar.gz $(HOME)/rpmbuild/SOURCES
	rpmbuild -bs --target x86_64 --nodeps rpm/credulous.spec
	rpmbuild -bb --target x86_64 --nodeps rpm/credulous.spec

# Create the source tarball with N-V prefix to match what the specfile expects
sources:
	@echo "Building for version '$(VERSION)'"
	sed -i -e 's/==VERSION==/$(VERSION)/' $(DOC)
	tar czvf $(TGZ) --transform='s|^|src/github.com/realestate-com-au/credulous/|' $(SRCS) $(TESTS)

debianpkg:
	@echo Build Debian packages
	sed -i -e 's/==VERSION==/$(VERSION)/' debian-pkg/DEBIAN/control
	sed -i -e 's/==VERSION==/$(VERSION)/' $(DOC)
	mkdir -p debian-pkg/usr/bin \
		debian-pkg/usr/share/man/man1 \
		debian-pkg/etc/bash_completion.d \
		debian-pkg/etc/profile.d
	cp $(HOME)/gopath/bin/credulous debian-pkg/usr/bin
	cp bash/credulous.sh debian-pkg/etc/profile.d
	cp bash/credulous.bash_completion debian-pkg/etc/bash_completion.d
	chmod 0755 debian-pkg/usr/bin/credulous
	pandoc -s -w man $(DOC) -o debian-pkg/usr/share/man/man1/credulous.1
	dpkg-deb --build debian-pkg
	mv debian-pkg.deb $(NAME)_$(VERSION)_amd64.deb

mock: mock-rpm
	@echo "BUILD COMPLETE; RPMS are in ."

mock-rpm: mock-srpm
	mock -r $(MOCK_CONFIG) --rebuild $(MOCK_SRPM)
	cp $(MOCK_RESULT)/$(RPM) .

mock-srpm: sources
	@echo "DIST is $(DIST)"
	@echo "RELEASE is $(RELEASE)"
	# mock -r $(MOCK_CONFIG) --init
	sed -e 's/==VERSION==/$(VERSION)/' $(SPEC_TMPL) > $(SPEC)
	mock -r $(MOCK_CONFIG) --buildsrpm --spec $(SPEC) --sources .
	rm -f $(SPEC)
	cp $(MOCK_RESULT)/$(MOCK_SRPM) .

clean:
	rm -f $(MOCK_SRPM) $(RPM) $(TGZ)

allclean:
	mock -r $(MOCK_CONFIG) --clean


C_GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*")
C_GO_TEST_PKGS = $(shell docker-compose run go list ./...)


deps:
	docker-compose run dep ensure -v

test:
	docker-compose run go test ${C_GO_TEST_PKGS}

gofmt:
	@echo "+++ Formatting code with Gofmt"
	@docker-compose run --rm gofmt -s -w ${C_GOFILES_NOVENDOR}

goimports:
	@echo "+++ Checking imports with go imports"
	@docker-compose run --rm goimports -e -l -w ${C_GOFILES_NOVENDOR}

lint:
	@echo "+++ Running gometalinter"
	@docker-compose run --rm gometalinter \
	--sort linter \
	--skip=client --skip=apis --skip=signals --skip vendor \
	--deadline 400s \
	--enable gofmt \
	--enable goimports \
	--enable gosimple \
	./... --debug
