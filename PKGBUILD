# Maintainer: Volodia Kraplich
pkgname=davinci-convert
pkgver=1.1.0
pkgrel=1
pkgdesc="Smart, high-performance tool to prepare media for DaVinci Resolve"
arch=('x86_64' 'aarch64')
url="https://github.com/VolodiaKraplich/davinci-convert"
license=('MIT')
depends=('ffmpeg')
makedepends=('go' 'make' 'upx')
options=('!debug')

build() {
    cd "$startdir"
    make build
}

package() {
    cd "$startdir"

    # Install binary
    install -Dm755 "bin/$pkgname" "$pkgdir/usr/bin/$pkgname"

    # Install documentation
    if [ -f README.md ]; then
        install -Dm644 README.md "$pkgdir/usr/share/doc/$pkgname/README.md"
    fi

    # Install license
    if [ -f LICENSE ]; then
        install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
    fi

    # Generate and install shell completions
    # Bash completion
    if "$pkgdir/usr/bin/$pkgname" completion bash >/dev/null 2>&1; then
        "$pkgdir/usr/bin/$pkgname" completion bash > "$pkgname.bash"
        install -Dm644 "$pkgname.bash" "$pkgdir/usr/share/bash-completion/completions/$pkgname"
    fi

    # Zsh completion
    if "$pkgdir/usr/bin/$pkgname" completion zsh >/dev/null 2>&1; then
        "$pkgdir/usr/bin/$pkgname" completion zsh > "$pkgname.zsh"
        install -Dm644 "$pkgname.zsh" "$pkgdir/usr/share/zsh/site-functions/_$pkgname"
    fi

    # Fish completion
    if "$pkgdir/usr/bin/$pkgname" completion fish >/dev/null 2>&1; then
        "$pkgdir/usr/bin/$pkgname" completion fish > "$pkgname.fish"
        install -Dm644 "$pkgname.fish" "$pkgdir/usr/share/fish/vendor_completions.d/$pkgname.fish"
    fi
}
