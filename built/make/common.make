

define generate_sbin
	@sed 's/process-name/$(1)/' built/sbin/go-monitor.sh > ${OUTPUT}/$(1)/sbin/go-monitor.sh
	@chmod +x ${OUTPUT}/$(1)/sbin/go-monitor.sh
endef

define init_app
	@mkdir -p ${OUTPUT}/$(1)/bin
	@mkdir -p ${OUTPUT}/$(1)/conf
	@mkdir -p ${OUTPUT}/$(1)/log
	@mkdir -p ${OUTPUT}/$(1)/sbin
	@mkdir -p ${OUTPUT}/$(1)/test

	@cp conf/$(env)/data.yml ${OUTPUT}/$(1)/conf
	@cp conf/$(env)/$(1).yml ${OUTPUT}/$(1)/conf/application.yml
endef

define release_app
	@mkdir -p release
	@cd ${OUTPUT} ; tar -czf ../release/$(1).tar.gz.$(DATE) $(1)
	@echo -e "\033[32m   多机部署 1. ansible-playbook deploy/$(1).yaml --extra-vars \"package=$(1).tar.gz.$(DATE)\" \033[0m"
	@echo -e "\033[32m   直接运行 2. cd $(OUTPUT)/$(1);./bin/$(1) --config conf/application.yml \033[0m"
	@echo -e "\033[32m   本地部署 3. mv $(OUTPUT)/$(1) /usr/local/vntop/ \033[0m"
endef

define upload
	@echo "curl xb-svr-ftp.oss-cn-hongkong.aliyuncs.com  -T ./release/$(1).tar.gz.$(DATE)"
	curl xb-svr-ftp.oss-cn-hongkong.aliyuncs.com  -T ./release/$(1).tar.gz.$(DATE)
	@echo "eg: 手工部署到server与测试方式1: "
	@echo "cd /usr/local/xbnet/server;wget xb-svr-ftp.oss-cn-hongkong.aliyuncs.com/$(1).tar.gz.$(DATE); tar zxvf $(1).tar.gz.$(DATE);"
	@echo "./charge_server/sbin/go-monitor.sh restart; sleep 3; ./charge_server/bin/httptest"
	@echo ""
	@echo "eg: 手工部署server与测试方式2（前置条件：update_server.sh部署到服务器）"
	@echo ""
	@echo "cd /usr/local/xbnet/server; sh ./update_server.sh $(1).tar.gz.$(DATE)"
	@echo ""
endef