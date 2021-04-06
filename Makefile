################################################################################
# Target: test-configure
################################################################################
.PHONY: test-configure
test-configure:
	go test ./pkg/configure
################################################################################
# Target: test-record
################################################################################
.PHONY: test-record
test-record:
	go test ./pkg/record
################################################################################
# Target: test-utils
################################################################################
.PHONY: test-utils
test-utils:
	go test ./pkg/utils
################################################################################
# Target: test
################################################################################
.PHONY: test
test: test-configure test-record test-utils
################################################################################
# Target: install
################################################################################
.PHONY: install
install:
	go install ./cmd/omega