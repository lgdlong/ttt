.PHONY: help run-workflow test clean logs show-config

# Variables
WORKFLOW ?= transcript_to_json
PROVIDER ?= gemini

# Default target
help:
	@echo "Available commands:"
	@echo ""
	@echo "Workflows:"
	@echo "  make run-workflow WORKFLOW=<name> PROVIDER=<provider>"
	@echo "    WORKFLOW: transcript_to_json (default)"
	@echo "    PROVIDER: gemini, openai (default: gemini)"
	@echo "    Example: make run-workflow WORKFLOW=transcript_to_json PROVIDER=openai"
	@echo ""
	@echo "Testing:"
	@echo "  make test                      - Test current PROVIDER"
	@echo "  make test PROVIDER=openai      - Test OpenAI API"
	@echo "  make test PROVIDER=gemini      - Test Gemini API"
	@echo "  make test-all                  - Test all providers"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean                     - Clean output and log files"
	@echo "  make logs                      - Show recent API logs"
	@echo "  make show-config               - Show current configuration"

# Workflow execution
run-workflow:
	@echo "Running $(WORKFLOW) workflow with $(PROVIDER)..."
	python -c "from workflows import run_$(WORKFLOW); import asyncio; asyncio.run(run_$(WORKFLOW)(provider='$(PROVIDER)'))"

# Testing commands
test:
	@if [ "$(PROVIDER)" = "openai" ]; then \
		echo "Testing OpenAI API..."; \
		python test_custom_openai.py; \
	elif [ "$(PROVIDER)" = "gemini" ]; then \
		echo "Testing Gemini API..."; \
		python test_providers.py; \
	else \
		echo "Unknown provider: $(PROVIDER)"; \
		exit 1; \
	fi

test-all:
	@echo "Testing all providers..."
	@$(MAKE) test PROVIDER=gemini
	@$(MAKE) test PROVIDER=openai
	@echo "All API tests completed"

# Utility commands
clean:
	@echo "Cleaning output and log files..."
	rm -f resources/transcript_to_json/output/*.json
	rm -f resources/transcript_to_json/output/*.error.txt
	rm -f logs/gemini_logs.txt
	@echo "Clean completed"

logs:
	@echo "Recent API logs (last 20 lines):"
	@tail -n 20 logs/gemini_logs.txt || echo "No logs found"

show-config:
	@echo "Current Configuration:"
	@echo "====================="
	@echo "Gemini Model: $$(grep GEMINI_MODEL .env | cut -d'=' -f2)"
	@echo "OpenAI Model: $$(grep OPENAI_MODEL .env | cut -d'=' -f2)"
	@echo "OpenAI Base URL: $$(grep OPENAI_BASE_URL .env | cut -d'=' -f2)"
	@echo "Current WORKFLOW: $(WORKFLOW)"
	@echo "Current PROVIDER: $(PROVIDER)"
