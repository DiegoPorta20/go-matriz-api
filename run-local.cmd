@echo off
rem Carga el .env y arranca go-api.
rem Existe porque go-api toma su configuracion del entorno, igual que en Lambda, y cmd.exe no
rem tiene un equivalente de "set -a; . ./.env; set +a".

setlocal
pushd "%~dp0"

if not exist ".env" (
  echo No hay .env en %~dp0
  echo Copia .env.example a .env y rellena los valores.
  popd
  exit /b 1
)

rem eol=# ignora los comentarios; tokens=1,* preserva los "=" que aparezcan dentro del valor.
for /f "usebackq eol=# tokens=1,* delims==" %%a in (".env") do set "%%a=%%b"

go run ./cmd/api

popd
