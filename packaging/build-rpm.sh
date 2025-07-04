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
BUILD_DIR="build-rpm"
TARBALL_NAME="${APPNAME}-${VERSION}.tar.gz"
RPMBUILD_DIR="/root/rpmbuild"  # Explicitly set to /root/rpmbuild for container
SRC_DIR="$PWD"

# Create build directory
mkdir -p "$BUILD_DIR"

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
if ! tar -czf "${BUILD_DIR}/${TARBALL_NAME}" -C /tmp "${APPNAME}-${VERSION}"; then
    echo "Error: Failed to create tarball ${BUILD_DIR}/${TARBALL_NAME}"
    exit 1
fi

# Clean up temporary directory
rm -rf "$TEMP_TAR_DIR"

# Build RPM using Docker
docker run --rm -v "$SRC_DIR:/src" -w /src rockylinux:9 bash -c "
    set -e
    dnf install -y golang rpm-build rpmdevtools &&
    rpmdev-setuptree &&
    if [ ! -d \"$RPMBUILD_DIR/SOURCES\" ]; then
        echo \"Error: $RPMBUILD_DIR/SOURCES does not exist after rpmdev-setuptree\"
        exit 1
    fi &&
    cp /src/${BUILD_DIR}/${TARBALL_NAME} $RPMBUILD_DIR/SOURCES/ &&
    sed \"s/VERSION/${VERSION}/g\" /src/packaging/rpm/${APPNAME}.spec > $RPMBUILD_DIR/SPECS/${APPNAME}.spec &&
    rpmbuild -ba $RPMBUILD_DIR/SPECS/${APPNAME}.spec &&
    cp $RPMBUILD_DIR/RPMS/x86_64/${APPNAME}-${VERSION}-1.el9.x86_64.rpm /src/${BUILD_DIR}/$APPNAME-v${VERSION}-1.el9.x86_64.rpm
"

if [ $? -eq 0 ]; then
    echo "✅ RPM package built: ${BUILD_DIR}/${APPNAME}-v${VERSION}-1.el9.x86_64.rpm"
else
    echo "Error: RPM build failed"
    exit 1
fi
