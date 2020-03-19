all: help

build:
	@echo "=> building server-api"
	@./scripts/build.sh 

pkg-rpm:
	@./scripts/build.sh rpm

clean:
	@rm -rf ./bin \
	rm -rf *.rpm
.PHONY:
	all build pkg-rpm clean
help:
	@echo "Make Targets:"
	@echo " make build         - build binaries "
	@echo " make pkg-rpm       - build RPM packages "
	@echo " make clean         - Remove the generated files "