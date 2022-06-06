@ECHO OFF

REM This is the parent directory of the directory containing this script (resolves to :install_root/Puppet)
SET ADWATCHD_DIR=%~dp0

REM Add the adwatchd bindir to the PATH
SET PATH=%ADWATCHD_DIR%;%PATH%

REM Display Ruby version
adwatchd.exe --help