#! /usr/bin/make -f

.DEFAULT_GOAL := build-all
include built/make/*.make

export GOPROXY=https://goproxy.io
export OUTPUT=output
export env=dev
export module=demo
export httptest
export GO111MODULE=on
export CGO_ENABLED=0
export group=group

DATE="$(shell TZ=UTC-8 date '+%Y%m%d%H%M')"


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


dtest: build
	curl -u 'deploy:9kfdhsiw28' ftp://192.168.0.22/lvfei/ -T release/${module}.tar.gz.${DATE}
	curl -XPOST http://192.168.0.22:14000/api/deploy/dev -d'{"group":"${group}", "module":"${module}", "file":"${module}.tar.gz.${DATE}"}'

goup:
	rm -f go.sum
	sed -i '/github.com\/xbonlinenet\/goup/d' go.mod


help:
	@echo
	@echo '  Usage:'
	@echo '    make build env=<enviroment>  module=<module-name>'
	@echo
	@echo '  Enviroments:'
	@echo '    dev'
	@echo '    test'
	@echo '    product[default]'
	
