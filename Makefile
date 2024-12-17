FLAGS=-ldflags "-s -w"
PLATFORM := $(shell uname)

ifeq ($(PLATFORM),$(filter MSYS_NT%,$(PLATFORM)))
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
run:
	build/$(EXEC_FILE)
