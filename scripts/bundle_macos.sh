#!/usr/bin/env bash
# Assembles the macOS application bundle.
set -euo pipefail

readonly APP_NAME="Noggin MCP"
readonly BUNDLE_ID="com.andgate.guided-study"
readonly VERSION="0.1.0"
readonly MINIMUM_MACOS="13.0"
readonly EXECUTABLE="noggin-mcp"
readonly CONVERTER="pdf-converter"
readonly SOURCE_ICON="cmd/noggin-mcp/assets/tray-icon.png"
readonly BUILD_DIR="bin"
readonly WORK_DIR=".cache/bundle"

readonly BUNDLE="${BUILD_DIR}/${APP_NAME}.app"
readonly CONTENTS="${BUNDLE}/Contents"
readonly ICONSET="${WORK_DIR}/icon.iconset"

# The build must run before the bundle.
for binary in "${EXECUTABLE}" "${CONVERTER}"; do
	if [[ ! -f "${BUILD_DIR}/${binary}" ]]; then
		echo "Missing ${BUILD_DIR}/${binary}. Build first." >&2
		exit 1
	fi
done

# Start from an empty bundle.
rm -rf "${BUNDLE}" "${ICONSET}"
mkdir -p "${CONTENTS}/MacOS" "${CONTENTS}/Resources" "${ICONSET}"

# The converter must sit beside the executable.
cp "${BUILD_DIR}/${EXECUTABLE}" "${CONTENTS}/MacOS/${EXECUTABLE}"
cp "${BUILD_DIR}/${CONVERTER}" "${CONTENTS}/MacOS/${CONVERTER}"

# Render the icon sizes the Dock uses.
for entry in 16:16x16 32:16x16@2x 32:32x32 64:32x32@2x 128:128x128 256:128x128@2x 256:256x256; do
	size="${entry%%:*}"
	name="${entry##*:}"
	sips --resampleHeightWidth "${size}" "${size}" "${SOURCE_ICON}" \
		--out "${ICONSET}/icon_${name}.png" >/dev/null
done
iconutil --convert icns "${ICONSET}" --output "${CONTENTS}/Resources/icon.icns"

# Describe the bundle to Launch Services.
cat >"${CONTENTS}/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleDevelopmentRegion</key>
	<string>en</string>
	<key>CFBundleDisplayName</key>
	<string>${APP_NAME}</string>
	<key>CFBundleExecutable</key>
	<string>${EXECUTABLE}</string>
	<key>CFBundleIconFile</key>
	<string>icon</string>
	<key>CFBundleIdentifier</key>
	<string>${BUNDLE_ID}</string>
	<key>CFBundleInfoDictionaryVersion</key>
	<string>6.0</string>
	<key>CFBundleName</key>
	<string>${APP_NAME}</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleShortVersionString</key>
	<string>${VERSION}</string>
	<key>CFBundleVersion</key>
	<string>${VERSION}</string>
	<key>LSMinimumSystemVersion</key>
	<string>${MINIMUM_MACOS}</string>
	<key>NSHighResolutionCapable</key>
	<true/>
</dict>
</plist>
PLIST

# Ad-hoc signatures let the bundle launch locally.
codesign --force --sign - "${CONTENTS}/MacOS/${CONVERTER}"
codesign --force --sign - "${BUNDLE}"

echo "Built ${BUNDLE}"
