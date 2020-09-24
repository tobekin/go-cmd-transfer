@echo off
chcp 65001
 
set projectPath=%~dp0
set GOBIN=%projectPath%
cd /d %projectPath%
 
REM 获取模块名
for /F %%i in ('go list -m') do ( set module=%%i)
 
REM 获取生成的二进制文件名
set lj=%module%
set "lj=%lj:/= %"
for %%i in (%lj%) do set binFileName=%%i
 
echo "工程目录:" %projectPath%
echo "执行文件:" %projectPath%/%binFileName%
go install
%binFileName%