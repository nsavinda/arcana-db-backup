#!/bin/bash
set -e

# Input validation
RAW_VERSION="$1"
if [ -z "$RAW_VERSION" ]; then
    echo "Error: Version number is required (e.g., v0.3.0)"
    exit 1
fi

# Sanitize version
VERSION="${RAW_VERSION#v}"
if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: Invalid version format. Expected vX.Y.Z"
    exit 1
fi

APPNAME="arcanadbbackup"
PKG_DIR="packaging/arch"
PKGBUILD_PATH="$PKG_DIR/PKGBUILD"
TARBALL_NAME="${APPNAME}-${VERSION}.tar.gz"
SRC_DIR="$PWD"

# Create source tarball
TEMP_TAR_DIR="/tmp/${APPNAME}-${VERSION}"
rm -rf "$TEMP_TAR_DIR"
mkdir -p "$TEMP_TAR_DIR"

# Copy necessary files
if ! cp -r config database encryption storage *.go go.mod go.sum LICENSE README.md example.config.yaml "$TEMP_TAR_DIR/"; then
    echo "Error: Failed to copy files to $TEMP_TAR_DIR"
    exit 1
fi

# Create tarball
if ! tar -czf "${TARBALL_NAME}" -C /tmp "${APPNAME}-${VERSION}"; then
    echo "Error: Failed to create tarball ${TARBALL_NAME}"
    exit 1
fi

# Prepare PKGBUILD from template
if ! sed "s/VERSION/${VERSION}/g" "$PKG_DIR/PKGBUILD.template" > "$PKGBUILD_PATH"; then
    echo "Error: Failed to generate PKGBUILD"
    exit 1
fi

# Build package using Docker
docker run --rm -v "$SRC_DIR:/src" -w /src archlinux:base-devel bash -c "
    set -e
    pacman -Syu --noconfirm go git base-devel &&
    useradd -m builder &&
    chown -R builder:builder /src &&
    su builder -c '
        mkdir -p /home/builder/pkgbuild &&
        cp /src/$PKGBUILD_PATH /home/builder/pkgbuild/PKGBUILD &&
        cp /src/$TARBALL_NAME /home/builder/pkgbuild/ &&
        cd /home/builder/pkgbuild &&
        makepkg -f --noconfirm
    ' &&
    cp /home/builder/pkgbuild/${APPNAME}-${VERSION}-1-x86_64.pkg.tar.zst /src/
"

# Move the built package
if [ -f "${APPNAME}-${VERSION}-1-x86_64.pkg.tar.zst" ]; then
    mv "${APPNAME}-${VERSION}-1-x86_64.pkg.tar.zst" "${APPNAME}-v${VERSION}-1-x86_64.pkg.tar.zst"
    echo "✅ Arch package built: ${APPNAME}-v${VERSION}-1-x86_64.pkg.tar.zst"
else
    echo "Error: Package not found"
    exit 1
fi

# Clean up
rm -f "$PKGBUILD_PATH" "$TARBALL_NAME"
