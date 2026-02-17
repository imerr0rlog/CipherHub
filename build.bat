@echo off
REM CipherHub Build Script for Windows
REM Usage: build.bat [target]
REM Targets: windows, linux, darwin, all, clean

setlocal EnableDelayedExpansion

set VERSION=1.0.0
set BINARY_NAME=cipherhub
set MAIN_PATH=./cmd/cipherhub
set BUILD_DIR=bin

if not exist %BUILD_DIR% mkdir %BUILD_DIR%

if "%1"=="windows" goto build_windows
if "%1"=="linux" goto build_linux
if "%1"=="darwin" goto build_darwin
if "%1"=="all" goto build_all
if "%1"=="clean" goto clean
goto usage

:build_windows
echo Building for Windows AMD64...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o %BUILD_DIR%/%BINARY_NAME%-windows-amd64.exe %MAIN_PATH%
echo Built: %BUILD_DIR%/%BINARY_NAME%-windows-amd64.exe
goto end

:build_linux
echo Building for Linux AMD64...
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o %BUILD_DIR%/%BINARY_NAME%-linux-amd64 %MAIN_PATH%
echo Built: %BUILD_DIR%/%BINARY_NAME%-linux-amd64
goto end

:build_darwin
echo Building for macOS AMD64...
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="-s -w" -o %BUILD_DIR%/%BINARY_NAME%-darwin-amd64 %MAIN_PATH%
echo Built: %BUILD_DIR%/%BINARY_NAME%-darwin-amd64
goto end

:build_all
echo Building for all platforms...
set GOOS=windows
set GOARCH=amd64
go build -ldflags="-s -w" -o %BUILD_DIR%/%BINARY_NAME%-windows-amd64.exe %MAIN_PATH%
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o %BUILD_DIR%/%BINARY_NAME%-linux-amd64 %MAIN_PATH%
set GOOS=darwin
set GOARCH=amd64
go build -ldflags="-s -w" -o %BUILD_DIR%/%BINARY_NAME%-darwin-amd64 %MAIN_PATH%
echo All builds complete.
goto end

:clean
echo Cleaning build directory...
if exist %BUILD_DIR% rmdir /s /q %BUILD_DIR%
go clean
echo Cleaned.
goto end

:usage
echo CipherHub Build Script for Windows
echo.
echo Usage: %0 [target]
echo.
echo Targets:
echo   windows   Build for Windows AMD64
echo   linux     Build for Linux AMD64
echo   darwin    Build for macOS AMD64
echo   all       Build for all platforms
echo   clean     Clean build directory
echo.
echo Current default: Building Windows AMD64
goto build_windows

:end
endlocal
