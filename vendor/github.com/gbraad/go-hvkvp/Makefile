PREFIX=hvkvp
DESCRIBE=$(git describe --tags)
VERSION=0.0.3

TARGETS=$(addprefix $(PREFIX)-, centos7 fedora26)

build: $(TARGETS)

$(PREFIX)-%: build/Dockerfile.compile-%
	mkdir -p out
	docker rmi -f $@ >/dev/null  2>&1 || true
	docker rm -f $@-extract > /dev/null 2>&1 || true
	echo "Building binaries for $@"
	docker build -t compile-$@ -f $< .
	docker create --name $@-extract compile-$@ sh
	docker cp $@-extract:/workspace/bin/hvkvp ./hvkvp
	docker rm $@-extract || true
	tar czvf ./out/$@-$(VERSION).tar.gz hvkvp
	mv ./hvkvp ./out/$@
	#docker rmi $@ || true

clean:
	rm -f ./$(PREFIX)-*
