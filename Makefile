Target: help ## Description
	@echo

OUT=h2d

.FORCE:

help: ##  This help dialog.
	@cat $(MAKEFILE_LIST) | perl -ne 's/(^\S+): .*##\s*(.+)/printf "\n %-16s %s", $$1,$$2/eg'

build: .FORCE  ## Build the binary
	go build

get: .FORCE  ## Install into existing golang setup
	export GO15VENDOREXPERIMENT=1
	go get github.com/rmohid/h2d
	go get golang.org/x/text
	go get golang.org/x/crypto

all: build test
	@echo 

clean: ## Remove derived files
	@echo rm $(OUT)
	go clean

test: ## Basic unit tests
	make test-client1
	
test1: 
	go clean
	go install
	make test-client2

test-client1: ## Test a simple http2 connection
	$(OUT) start --dump &
	$(OUT) connect http2.akamai.com
	$(OUT) get /index.html > /dev/null
	$(OUT) stop

test-client2: ## Test a simple http connection
	$(OUT) start --dump &
	$(OUT) connect akamai.com > /dev/null
	$(OUT) get /index.html
	$(OUT) stop

test-client3: 
	$(OUT) start --dump &
	$(OUT) connect http2.akamai.com
	$(OUT) stop

test-wiretap1: ## Test wiretap functionality
	$(OUT) wiretap localhost:8888 http2.akamai.com &
	@echo Connect your browser to https://localhost:8888
	@read -rsp $$'Press any key to continue...\n' -n1 key
	@pkill $(OUT) 

