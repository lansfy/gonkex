@echo off
echo Starting
rem timeout doesn't work here
ping 127.0.0.1 -n 3 >nul
echo Finished
