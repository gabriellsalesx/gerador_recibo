#!/bin/sh
set -eu

EMISSOR_BIN_DIR="${EMISSOR_BIN_DIR:-$HOME/.local/bin}"
APP_NAME="emissor-cli"
bin="$EMISSOR_BIN_DIR/$APP_NAME"

if [ -f "$bin" ]; then
	rm -f "$bin"
	printf 'Binario removido: %s\n' "$bin"
else
	printf 'Binario nao encontrado em %s (talvez ja tenha sido removido).\n' "$bin"
fi

config_dir="${XDG_CONFIG_HOME:-$HOME/.config}/emissor"
if [ -d "$config_dir" ]; then
	rm -rf "$config_dir"
	printf 'Configuracao e metadados removidos: %s\n' "$config_dir"
fi

printf '\nSeus documentos em ~/Documentos/Emissor NAO foram apagados.\n'
printf 'Se quiser remove-los, apague a pasta manualmente.\n'
