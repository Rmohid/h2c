Target: help ## Description
	@echo

OUT=h2c

.FORCE:

help: ##  This help dialog.
	@cat $(MAKEFILE_LIST) | perl -ne 's/(^\S+): .*##\s*(.+)/printf "\n %-16s %s", $$1,$$2/eg'

build: .FORCE  ## Build the binary
	go get golang.org/x/net/http2/hpack
	go get github.com/fatih/color
	go install

get: .FORCE  ## Install into existing golang setup
	export GO15VENDOREXPERIMENT=1
	go get github.com/rmohid/h2c

all: build test
	@echo 

clean: ## Remove derived files
	@echo rm $(OUT)
	go clean

test: ## Basic unit tests
	make test-client1
	
test1: 
	go install
	make test-client1

test-client1: ## Test a simple http2 connection
	$(OUT) start --dump &
	$(OUT) connect http2.akamai.com
	$(OUT) get /index.html > /dev/null
	$(OUT) disconnect
	$(OUT) stop

test-client2: ## Test a simple http connection
	$(OUT) start --dump &
	$(OUT) connect akamai.com > /dev/null
	$(OUT) get /index.html
	$(OUT) disconnect
	$(OUT) stop

test-client3: ## Test a more complex http2 connection with pushes
	pkill $(OUT) || true
	$(OUT) start --dump &
	$(OUT) connect http2.cloudflare.com > /dev/null
	$(OUT) get /  > /dev/null
	sleep 4
	$(OUT) disconnect
	$(OUT) stop

test-client4: ## Test a server push every second
	$(OUT) start --dump &
	$(OUT) connect http2.golang.org
	$(OUT) get /clockstream &
	sh -c "sleep 4; $(OUT) disconnect; $(OUT) stop" &

test-client5: ## Test a page with many images
	$(OUT) start --dump &
	$(OUT) connect http2.golang.org
	$(OUT) get /gophertiles
	$(OUT) disconnect
	$(OUT) stop

test-wiretap1: ## Test wiretap functionality
	pkill $(OUT) || true
	$(OUT) wiretap localhost:8888 http2.akamai.com &
	@echo Connect your browser to https://localhost:8888
	@read -rsp $$'Press any key to continue...\n' -n1 key
	@pkill $(OUT) 

test-wiretap2: ## Test wiretap with server push
	pkill $(OUT) || true
	$(OUT) wiretap localhost:8889 http2.cloudflare.com &
	@echo Connect your browser to https://localhost:8889
	@read -rsp $$'Press any key to continue...\n' -n1 key
	@pkill $(OUT) 
