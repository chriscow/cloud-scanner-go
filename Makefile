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

$(BLDDIR)/%:
	@mkdir -p $(dir $@)
	go build ${BLDFLAGS} -o $@ ./apps/$*

$(APPS): %: $(BLDDIR)/%

clean:
	rm -fr $(BLDDIR)