.PHONY: run test lint format

run:
	uvicorn app.main:app --reload

test:
	pytest

test-cov:
	pytest --cov=app --cov-report=html

lint:
	ruff check app/ tests/

format:
	ruff format app/ tests/
