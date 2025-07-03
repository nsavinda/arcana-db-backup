#!/bin/bash
set -e

RAW_VERSION="$1"  # Example: v0.3.0
VERSION="${RAW_VERSION#v}"
APPNAME="arcanadbbackup"
BUILD_DIR="build-rpm"
TARBALL_NAME="${APPNAME}-${VERSION}.tar.gz"
RPMBUILD_DIR="/root/rpmbuild"

mkdir -p "$BUILD_DIR"

# Create source tarball
TEMP_TAR_DIR="/tmp/${APPNAME}-${VERSION}"
rm -rf "$TEMP_TAR_DIR"
mkdir -p "$TEMP_TAR_DIR"
cp -r * "$TEMP_TAR_DIR/"
tar -czf "${TARBALL_NAME}" -C /tmp "${APPNAME}-${VERSION}"

docker run --rm -v "$PWD":/src -w /src rockylinux:9 bash -c "
  dnf install -y golang rpm-build &&
  mkdir -p $RPMBUILD_DIR/{BUILD,RPMS,SOURCES,SPECS,SRPMS} &&
  cp /src/${TARBALL_NAME} $RPMBUILD_DIR/SOURCES/ &&
  sed 's/VERSION/${VERSION}/g' /src/packaging/rpm/${APPNAME}.spec > $RPMBUILD_DIR/SPECS/${APPNAME}.spec &&
  cd $RPMBUILD_DIR &&
  rpmbuild -ba SPECS/${APPNAME}.spec &&
  cp RPMS/x86_64/${APPNAME}-${VERSION}-1.el9.x86_64.rpm /src/${BUILD_DIR}/
"

echo "✅ RPM package built: ${BUILD_DIR}/${APPNAME}-${VERSION}-1.el9.x86_64.rpm"
