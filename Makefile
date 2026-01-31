.PHONY: up tag

up:
	git add .
	@read -p "Commit message: " msg; \
	msg=$${msg:-"update"}; \
	git commit -m "$$msg"
	git pull origin master
	git push origin master
	@echo "\n 代码提交发布..."

# 获取当前最新 Tag
CURRENT_VERSION = $(shell git describe --tags --abbrev=0 2>/dev/null || echo v1.8.0)
# 计算自增后的版本号 (增加最后一位)
NEXT_VERSION = $(shell echo $(CURRENT_VERSION) | awk -F. '{$$(NF) = $$(NF) + 1;} 1' OFS=.)

tag:
	git pull origin master
	git add .
	@read -p "Commit message: " msg; \
	msg=$${msg:-"update"}; \
	git commit -m "$$msg"; \
	git push origin master

	@echo "当前版本: $(CURRENT_VERSION)"
	@read -p "请输入新版本号 [默认自增: $(NEXT_VERSION)]: " v; \
	v_final=$${v:-$(NEXT_VERSION)}; \
	git tag $$v_final; \
	git push origin --tags; \
	echo "\n $$v_final 发布成功..."