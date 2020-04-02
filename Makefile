#! /usr/bin/make -f

.DEFAULT_GOAL := build-all
include built/make/*.make

export GOPROXY=https://goproxy.io
export DATE=$(shell TZ=UTC-8 date '+%Y%m%d%H%M')
export OUTPUT=output
export env=product
export httptest
export module
export GO111MODULE=on
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64


build: init
	python built/go/env_check.py 1.12
	$(call init_app,${module})
	$(call generate_sbin,${module})
	go build -i -o ${OUTPUT}/${module}/bin/${module}-$(DATE) ./pkg/cmd/${module}
ifeq ($(httptest),1)
	go test -c -o ${OUTPUT}/${module}/bin/httptest ./pkg/cmd/${module}/httptest
endif
	cd ${OUTPUT}/${module}/bin && ln -sf ${module}-$(DATE) ${module}

	$(call release_app,${module})

run:
	cd ${OUTPUT}/${module} && ./bin/${module} --config conf/application.yml


clean:
	@rm -rf ${OUTPUT}
	@rm -rf release

help:
	@echo
	@echo '  Usage:'
	@echo '    make build env=<enviroment>  module=<module-name>'
	@echo
	@echo '  Enviroments:'
	@echo '    dev'
	@echo '    test'
	@echo '    product[default]'
	
