@echo off
rem Сглобява DWG-Viewer.exe (нужен е Go: https://go.dev/dl). Пускай от папката windows.
rmdir /s /q web 2>nul
mkdir web
copy ..\index.html web\ >nul
copy ..\manifest.webmanifest web\ >nul
copy ..\sw.js web\ >nul
copy ..\apple-touch-icon.png web\ >nul
xcopy ..\icons web\icons\ /e /i /q >nul
set CGO_ENABLED=0
go build -trimpath -ldflags "-s -w -H=windowsgui" -o DWG-Viewer.exe .
echo Готово: DWG-Viewer.exe
