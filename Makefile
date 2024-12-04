FLAGS=-ldflags "-s -w"
.PHONY: build
build: build_clean build_selector
build_clean:
	rm -rf build && mkdir build
build_selector: cmd/ygg-peers-select/main.go
	go build -o build/ygg-peers-select $(FLAGS) $<
	sudo setcap cap_net_raw=+ep build/ygg-peers-select
run:
	build/peers_parse
