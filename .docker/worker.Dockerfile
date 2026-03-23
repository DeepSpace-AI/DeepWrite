FROM python:3.13-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# Copy project files
COPY services/worker/ .

# Install uv for fast package management
RUN pip install uv

# Install dependencies
RUN uv sync --frozen --prod

# Create non-root user
RUN useradd -m -u 1000 worker
USER worker

# Expose ports
EXPOSE 8000 8001

# Default command (all services)
CMD ["uv", "run", "python", "main.py"]
