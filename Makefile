FLAGS=-ldflags "-s -w"
PLATFORM := $(shell uname)

ifeq ($(PLATFORM),$(findsting $(PLATFORM),MSYS_NT))
EXEC_FILE=ygg-peers-select.exe
else
EXEC_FILE=ygg-peers-select
endif

.PHONY: build
build: build_clean build_selector
build_clean:
	rm -rf build/* && (mkdir build || true)
build_selector: cmd/ygg-peers-select/main.go
	go build -o build/$(EXEC_FILE) $(FLAGS) $<
ifeq ($(PLATFORM),Linux)
	sudo setcap cap_net_raw=+ep build/$(EXEC_FILE)
endif
run:
	build/$(EXEC_FILE)
