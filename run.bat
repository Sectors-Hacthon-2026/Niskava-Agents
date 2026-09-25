@echo off
setlocal

if not exist "bin\niskava.exe" (
    echo Niskava Windows executable not found. Compiling bin\niskava.exe...
    if not exist "bin" mkdir bin
    go build -o bin\niskava.exe .\cmd\niskava
    if errorlevel 1 (
        echo Failed to build niskava.exe
        exit /b %errorlevel%
    )
)

bin\niskava.exe %*
