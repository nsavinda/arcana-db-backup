#!/bin/bash
set -e

RAW_VERSION="$1"  # Example: v0.3.0
VERSION="${RAW_VERSION#v}"
APPNAME="arcanadbbackup"
BUILD_DIR="build-rpm"
RPMBUILD_DIR="/root/rpmbuild"

mkdir -p "$BUILD_DIR"

docker run --rm -v "$PWD":/src -w /src rockylinux:9 bash -c "
  dnf install -y golang rpm-build &&
  mkdir -p $RPMBUILD_DIR/{BUILD,RPMS,SOURCES,SPECS,SRPMS} &&
  cp packaging/rpm/${APPNAME}.spec $RPMBUILD_DIR/SPECS/ &&
  cp main.go example.config.yaml $RPMBUILD_DIR/SOURCES/ &&
  sed -i 's/VERSION/${VERSION}/g' $RPMBUILD_DIR/SPECS/${APPNAME}.spec &&
  cd $RPMBUILD_DIR &&
  rpmbuild -ba SPECS/${APPNAME}.spec &&
  cp RPMS/x86_64/${APPNAME}-${VERSION}-1.el9.x86_64.rpm /src/${BUILD_DIR}/
"

echo "✅ RPM package built: ${BUILD_DIR}/${APPNAME}-${VERSION}-1.el9.x86_64.rpm"
