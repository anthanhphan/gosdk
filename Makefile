# Variables
PKGS := $(shell go list ./... | grep -vE '/example')
REPORT_DIR := code_report
COVER_OUT := $(REPORT_DIR)/coverage.out
COVER_HTML := $(REPORT_DIR)/coverage.html
SECURITY_JSON := $(REPORT_DIR)/security-report.json

# Tools
GOLANGCI_LINT ?= golangci-lint
GOSEC ?= gosec
GOVULNCHECK ?= govulncheck
STATICCHECK ?= staticcheck
INEFFASSIGN ?= ineffassign
MISSPELL ?= misspell
GOCYCLO ?= gocyclo

.PHONY: all
all: tidy fmt imports lint vet staticcheck ineffassign misspell cyclo test_coverage security_scan vul

# ------------------------------
# Tool Installation
# ------------------------------

install_tools:
	@echo "Installing Go quality tools..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/gordonklaus/ineffassign@latest
	go install github.com/client9/misspell/cmd/misspell@latest
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	@echo "All tools installed successfully."

# ------------------------------
# Basic Checks
# ------------------------------

fmt:
	@echo "Running go fmt..."
	@go fmt ./...

imports:
	@echo "Organizing imports..."
	@goimports -w .

tidy:
	@echo "Running go mod tidy..."
	@go mod tidy
	@go mod verify


# ------------------------------
# Tests & Coverage
# ------------------------------

test_coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(REPORT_DIR)
	@go clean -testcache
	@go test $(PKGS) -coverprofile=$(COVER_OUT)
	@go tool cover -html=$(COVER_OUT) -o $(COVER_HTML)
	@echo "Coverage report generated at: $(COVER_HTML)"


# ------------------------------
# Code Quality & Linting
# ------------------------------

vet:
	@echo "Running go vet..."
	@go vet ./...

lint:
	@echo "Running golangci-lint..."
	@$(GOLANGCI_LINT) run ./...

staticcheck:
	@echo "Running staticcheck..."
	@$(STATICCHECK) ./...

ineffassign:
	@echo "Checking for ineffectual assignments..."
	@$(INEFFASSIGN) ./...

misspell:
	@echo "Checking for spelling errors..."
	@$(MISSPELL) ./...

cyclo:
	@echo "Checking cyclomatic complexity..."
	@$(GOCYCLO) -over 15 .

# ------------------------------
# Security and Vulnerability
# ------------------------------

security_scan:
	@echo "Running gosec security scan..."
	@mkdir -p $(REPORT_DIR)
	@$(GOSEC) -fmt json -out $(SECURITY_JSON) ./...
	@echo "Security report generated at: $(SECURITY_JSON)"

vul:
	@echo "Running govulncheck..."
	@$(GOVULNCHECK) ./...
