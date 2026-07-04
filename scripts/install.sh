#!/bin/sh
set -eu

EMISSOR_REPO="${EMISSOR_REPO:-https://github.com/gabriellsalesx/Emissor_CLI}"
EMISSOR_VERSION="${EMISSOR_VERSION:-latest}"
EMISSOR_BIN_DIR="${EMISSOR_BIN_DIR:-$HOME/.local/bin}"
APP_NAME="emissor-cli"

info() { printf '%s\n' "$*"; }
err()  { printf 'Erro: %s\n' "$*" >&2; exit 1; }

os="$(uname -s)"
case "$os" in
	Linux)  os="linux" ;;
	Darwin) os="darwin" ;;
	*) err "sistema operacional nao suportado: $os" ;;
esac

arch="$(uname -m)"
case "$arch" in
	x86_64|amd64) arch="amd64" ;;
	aarch64|arm64) arch="arm64" ;;
	*) err "arquitetura nao suportada: $arch" ;;
esac

asset="${APP_NAME}_${os}_${arch}"

if [ "$EMISSOR_VERSION" = "latest" ]; then
	url="${EMISSOR_REPO}/releases/latest/download/${asset}"
else
	url="${EMISSOR_REPO}/releases/download/${EMISSOR_VERSION}/${asset}"
fi

info "Baixando ${APP_NAME} (${os}/${arch})..."
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

if command -v curl >/dev/null 2>&1; then
	curl -fsSL "$url" -o "$tmp" || err "nao consegui baixar de $url"
elif command -v wget >/dev/null 2>&1; then
	wget -qO "$tmp" "$url" || err "nao consegui baixar de $url"
else
	err "preciso de curl ou wget instalado"
fi

mkdir -p "$EMISSOR_BIN_DIR"
install -m 0755 "$tmp" "$EMISSOR_BIN_DIR/$APP_NAME" 2>/dev/null || {
	cp "$tmp" "$EMISSOR_BIN_DIR/$APP_NAME"
	chmod 0755 "$EMISSOR_BIN_DIR/$APP_NAME"
}

info "Instalado em: $EMISSOR_BIN_DIR/$APP_NAME"

case ":$PATH:" in
	*":$EMISSOR_BIN_DIR:"*) ;;
	*)
		info ""
		info "A pasta $EMISSOR_BIN_DIR nao esta no seu PATH."
		info "Adicione esta linha ao seu ~/.bashrc ou ~/.zshrc:"
		info "  export PATH=\"$EMISSOR_BIN_DIR:\$PATH\""
		;;
esac

info ""
info "Pronto! Execute: $APP_NAME"
