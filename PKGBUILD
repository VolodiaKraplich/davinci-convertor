# Maintainer: Volodia Kraplich
pkgname=davinci-convertor
pkgver=0.2.0
pkgrel=1
pkgdesc="Smart, high-performance tool to prepare media for DaVinci Resolve"
arch=('x86_64' 'aarch64')
url="https://github.com/VolodiaKraplich/davinci-convertor"
license=('MIT')
depends=('ffmpeg')
makedepends=('rustup')
options=('!debug')

build() {
    cd "$startdir"
    cargo build --release --locked
}

package() {
    cd "$startdir"

    # Install binary
    install -Dm755 "target/release/$pkgname" "$pkgdir/usr/bin/$pkgname"

    # Install documentation
    install -Dm644 README.md "$pkgdir/usr/share/doc/$pkgname/README.md"

    # Install license
    install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"

    # --- Generate and install shell completions ---
    local bin_path="target/release/$pkgname"

    # Bash completion
    "$bin_path" completion bash > "$pkgname.bash"
    install -Dm644 "$pkgname.bash" "$pkgdir/usr/share/bash-completion/completions/$pkgname"

    # Zsh completion
    "$bin_path" completion zsh > "$pkgname.zsh"
    install -Dm644 "$pkgname.zsh" "$pkgdir/usr/share/zsh/site-functions/_$pkgname"

    # Fish completion
    "$bin_path" completion fish > "$pkgname.fish"
    install -Dm644 "$pkgname.fish" "$pkgdir/usr/share/fish/vendor_completions.d/$pkgname.fish"
}
