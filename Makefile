Target: help ## Description
	@echo

OUT=./h2c

.FORCE:

help: ##  This help dialog.
	@cat $(MAKEFILE_LIST) | perl -ne 's/(^\S+): .*##\s*(.+)/printf "\n %-16s %s", $$1,$$2/eg'

build: .FORCE  ## Build the binary
	go build

get: .FORCE  ## Install into existing golang setup
	export GO15VENDOREXPERIMENT=1
	go get github.com/rmohid/h2c

all: build test
	@echo 

clean: ## Remove derived files
	@echo rm $(OUT)
	go clean

test: ## Basic unit tests
	make test-http2-client1
	
test1: 
	go clean
	go build
	make test-http2-client2

test-http2-client1: ## Test a simple http2 connection
	$(OUT) start --dump &
	$(OUT) connect http2.akamai.com
	$(OUT) get /index.html > /dev/null
	$(OUT) stop

test-http-client1: ## Test a simple http2 connection
	$(OUT) start --dump &
	$(OUT) connect akamai.com > /dev/null
	$(OUT) get /index.html
	$(OUT) stop

test-http2-client2: ## Test a simple http2 connection
	$(OUT) start --dump &
	$(OUT) connect http2.akamai.com
	$(OUT) stop

