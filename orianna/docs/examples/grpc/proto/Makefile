.PHONY: generate clean

generate:
	@echo "Generating Go code from proto files..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		user.proto
	@echo "Done! Generated files:"
	@ls -la *.pb.go 2>/dev/null || echo "No .pb.go files found"

clean:
	@echo "Cleaning generated files..."
	rm -f *.pb.go
	@echo "Done!"

help:
	@echo "Available targets:"
	@echo "  generate - Generate Go code from proto files"
	@echo "  clean    - Remove generated .pb.go files"
	@echo "  help     - Show this help message"
