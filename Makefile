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
# Target: watch
################################################################################
.PHONY: watch
watch:
	@while [[ 1=1 ]] ; do \
		watch -n 1 -g 'ls -Rlrt ./pkg/* ./cmd/*' && \
		clear && \
		make test && \
		sleep 1 ;\
	done