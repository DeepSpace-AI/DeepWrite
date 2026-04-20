@echo off
echo ============================================
echo AI Service Test Runner
echo ============================================
echo.

REM Check if config.yaml exists
if not exist config.yaml (
    echo ERROR: config.yaml not found!
    echo Please copy config.yaml.example to config.yaml and update settings.
    exit /b 1
)

REM Check database connection
echo [Step 1] Checking database connection...
echo Make sure PostgreSQL is running and accessible.
echo.

REM Create test data
echo [Step 2] Creating test data...
uv run python setup_test_data.py
if %errorlevel% neq 0 (
    echo Failed to create test data. Check your database connection.
    exit /b 1
)
echo.

REM Start server in background
echo [Step 3] Starting AI service...
start /b uv run uvicorn main:app --host 0.0.0.0 --port 8010
timeout /t 3 /nobreak > nul
echo Server started on http://localhost:8010
echo.

REM Run tests
echo [Step 4] Running tests...
uv run python test_service.py
echo.

echo ============================================
echo Test complete!
echo ============================================
pause