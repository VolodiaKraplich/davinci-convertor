# Maintainer: Volodia Kraplich
pkgname=davinci-convert
pkgver=1.0.0
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

    install -Dm755 "bin/$pkgname" "$pkgdir/usr/bin/$pkgname"

    if [ -f README.md ]; then
        install -Dm644 README.md "$pkgdir/usr/share/doc/$pkgname/README.md"
    fi

    if [ -f LICENSE ]; then
        install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
    fi

    # Install Fish completion if generator exists
    if "$pkgdir/usr/bin/$pkgname" completion fish >/dev/null 2>&1; then
        "$pkgdir/usr/bin/$pkgname" completion fish > "$pkgname.fish"
        install -Dm644 "$pkgname.fish" "$pkgdir/usr/share/fish/vendor_completions.d/$pkgname.fish"
    fi
}
