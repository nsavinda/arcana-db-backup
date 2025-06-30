#!/bin/bash
set -e

APPNAME="arcanadbbackup"
VERSION="${1:-0.1.0}" # remove "v" prefix for Debian packaging
VERSION="${VERSION#v}" # Ensure version does not start with 'v'
ARCH="${2:-amd64}"

if [[ "$ARCH" == "amd64" ]]; then
  GOARCH=amd64
elif [[ "$ARCH" == "arm64" ]]; then
  GOARCH=arm64
else
  echo "Unsupported architecture: $ARCH"
  exit 1
fi

echo "Building $APPNAME version $VERSION for $ARCH..."

# 1. Clean previous builds
rm -rf dist
mkdir -p dist/$APPNAME/usr/local/bin
mkdir -p dist/$APPNAME/etc/$APPNAME

# 2. Build Go binary
GOOS=linux GOARCH=$GOARCH go build -o dist/$APPNAME/usr/local/bin/$APPNAME main.go

# 3. Copy config (edit as needed)
cp example.config.yaml dist/$APPNAME/etc/$APPNAME/config.yaml

# 4. Set up DEBIAN control files
mkdir -p dist/$APPNAME/DEBIAN
sed "s/VERSION/$VERSION/g; s/ARCH/$ARCH/g" packaging/debian/control > dist/$APPNAME/DEBIAN/control

# 5. Add optional maintainer scripts if present
for script in postinst prerm; do
  if [ -f packaging/debian/$script ]; then
    cp packaging/debian/$script dist/$APPNAME/DEBIAN/
    chmod 755 dist/$APPNAME/DEBIAN/$script
  fi
done

# 6. Build the .deb
dpkg-deb --build dist/$APPNAME

# 7. Move and name artifact
DEB_NAME=${APPNAME}_${VERSION}_${ARCH}.deb
mv dist/$APPNAME.deb $DEB_NAME
echo "Built: $DEB_NAME"
