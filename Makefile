PREFIX=/usr/local
BINDIR=${PREFIX}/bin
BLDDIR = build

all: gateway scanner persist qos

gateway:
	go build -o $(BLDDIR)/gateway ./apps/gateway/.

scanner:
	go build -o $(BLDDIR)/scanner ./apps/scanner/.

persist:
	go build -o $(BLDDIR)/persist ./apps/persist/.

qos:
	go build -o $(BLDDIR)/qos ./apps/qos/.

clean:
	rm -fr $(BLDDIR)