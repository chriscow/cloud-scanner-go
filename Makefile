PREFIX=/usr/local
BINDIR=${PREFIX}/bin
BLDDIR = build
BLDFLAGS=
EXT=
ifeq (${GOOS},windows)
	EXT=.exe
endif

APPS = gateway result_to_dynamo scan_radius
all: $(APPS)

$(BLDDIR)/gateway: $(wildcard apps/gateway/*.go geom/*.go scan/*.go util/*.go)
$(BLDDIR)/scan_radius: $(wildcard apps/scan_radius/*.go geom/*.go scan/*.go util/*.go)
$(BLDDIR)/result_to_dynamo: $(wildcard apps/result_to_dynamo/*.go geom/*.go scan/*.go util/*.go)

$(BLDDIR)/%:
	@mkdir -p $(dir $@)
	go build ${BLDFLAGS} -o $@ ./apps/$*

$(APPS): %: $(BLDDIR)/%

clean:
	rm -fr $(BLDDIR)