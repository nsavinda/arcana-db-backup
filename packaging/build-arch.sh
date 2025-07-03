#!/bin/bash
set -e

VERSION="${1#v}"
APPNAME="arcanadbbackup"

# Prepare PKGBUILD from template
PKG_DIR="packaging/arch"
PKGBUILD_PATH="$PKG_DIR/PKGBUILD"
sed "s/VERSION/${VERSION}/g" $PKG_DIR/PKGBUILD.template > $PKGBUILD_PATH

docker run --rm -v "$PWD":/src -w /src archlinux:base-devel bash -c "
  pacman -Syu --noconfirm go git base-devel &&
  useradd -m builder &&
  su builder -c '
    mkdir -p /home/builder/pkgbuild &&
    cp /src/$PKGBUILD_PATH /home/builder/pkgbuild/PKGBUILD &&
    cd /home/builder/pkgbuild &&
    makepkg -f --noconfirm
  ' &&
  cp /home/builder/pkgbuild/*.pkg.tar.zst /src/
"
# Move the built package to the current directory
mv ${APPNAME}-v${VERSION}-1-x86_64.pkg.tar.zst ./${APPNAME}-v${VERSION}-1-x86_64.pkg.tar.zst

echo "Arch package built: ${APPNAME}-v${VERSION}-1-x86_64.pkg.tar.zst"
