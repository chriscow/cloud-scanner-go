PREFIX=/usr/local
BINDIR=${PREFIX}/bin
BLDDIR = build
BLDFLAGS=
EXT=
ifeq (${GOOS},windows)
	EXT=.exe
endif

APPS = gateway persist scanner
all: $(APPS)

$(BLDDIR)/gateway: $(wildcard apps/gateway/*.go geom/*.go scan/*.go util/*.go)
$(BLDDIR)/scanner: $(wildcard apps/scanner/*.go geom/*.go scan/*.go util/*.go)
$(BLDDIR)/persist: $(wildcard apps/persist/*.go geom/*.go scan/*.go util/*.go)

$(BLDDIR)/%:
	@mkdir -p $(dir $@)
	go build ${BLDFLAGS} -o $@ ./apps/$*

$(APPS): %: $(BLDDIR)/%

clean:
	rm -fr $(BLDDIR)