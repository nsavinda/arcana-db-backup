#!/bin/bash
set -e

APPNAME="arcanadbbackup"
VERSION="${1:-0.1.0}"
VERSION="${VERSION#v}" # strip "v" if prefixed

WORKDIR=$(pwd)
BUILDROOT="$WORKDIR/build-rpm"
SRCDIR="$BUILDROOT/${APPNAME}-v${VERSION}"

# Cleanup
rm -rf "$BUILDROOT"
mkdir -p "$SRCDIR"

# Copy source files
cp main.go example.config.yaml README.md LICENSE "$SRCDIR/"
# If needed: cp -r config/ database/ encryption/ storage/ "$SRCDIR/"

# Create source tar.gz
tar czf "$BUILDROOT/${APPNAME}-v${VERSION}.tar.gz" -C "$BUILDROOT" "${APPNAME}-v${VERSION}"

# Prepare spec
SPECFILE="$BUILDROOT/${APPNAME}.spec"
sed "s/VERSION/v${VERSION}/g" packaging/rpm/arcanadbbackup.spec.template > "$SPECFILE"

# Build in Docker (CentOS)
docker run --rm -v "$BUILDROOT":/build -w /build centos:7 bash -c "
  yum install -y rpm-build golang &&
  mkdir -p /root/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS} &&
  cp /build/${APPNAME}-v${VERSION}.tar.gz /root/rpmbuild/SOURCES/ &&
  cp /build/${APPNAME}.spec /root/rpmbuild/SPECS/ &&
  cd /root &&
  rpmbuild -ba rpmbuild/SPECS/${APPNAME}.spec &&
  cp rpmbuild/RPMS/x86_64/${APPNAME}-v${VERSION}-1.el7.x86_64.rpm /build/
"

echo "✅ RPM built: $BUILDROOT/${APPNAME}-v${VERSION}-1.el7.x86_64.rpm"
