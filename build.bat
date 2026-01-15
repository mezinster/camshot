@echo off
echo Building CamShot Puzzle Maker...

REM Download dependencies
echo Downloading dependencies...
go mod tidy

REM Build for Windows
echo Compiling for Windows...
go build -ldflags="-H windowsgui" -o camshot.exe .

if %ERRORLEVEL% EQU 0 (
    echo.
    echo Build successful! Run camshot.exe to start the application.
) else (
    echo.
    echo Build failed. Please check the error messages above.
)

pause
