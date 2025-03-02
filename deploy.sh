#!/bin/bash
set -euo pipefail

# Define the version (change if needed)
VERSION="1.0.0"

echo "Building ARM64 binary..."
GOARCH=arm64 go build -o llmdog_arm64 ./cmd/llmdog

echo "Building AMD64 binary..."
GOARCH=amd64 go build -o llmdog_amd64 ./cmd/llmdog

# Prepare ARM64 tarball with the binary renamed to "llmdog"
echo "Packaging ARM64 binary..."
rm -rf temp_arm64
mkdir temp_arm64
# Copy and rename the binary into the temporary folder
cp llmdog_arm64 temp_arm64/llmdog
tar -C temp_arm64 -czvf llmdog_v${VERSION}_darwin_arm64.tar.gz llmdog

# Prepare AMD64 tarball with the binary renamed to "llmdog"
echo "Packaging AMD64 binary..."
rm -rf temp_amd64
mkdir temp_amd64
cp llmdog_amd64 temp_amd64/llmdog
tar -C temp_amd64 -czvf llmdog_v${VERSION}_darwin_amd64.tar.gz llmdog

# Clean up temporary directories
rm -rf temp_arm64 temp_amd64

echo "Computing SHA256 checksums..."
ARM64_SHA256=$(shasum -a 256 llmdog_v${VERSION}_darwin_arm64.tar.gz | awk '{print $1}')
AMD64_SHA256=$(shasum -a 256 llmdog_v${VERSION}_darwin_amd64.tar.gz | awk '{print $1}')

echo "ARM64 SHA256: $ARM64_SHA256"
echo "AMD64 SHA256: $AMD64_SHA256"

echo "Generating Homebrew formula (llmdog.rb)..."
cat <<EOF > llmdog.rb
# frozen_string_literal: true
# typed: true
#
# Formula for llmdog, a tool to prepare files for LLM consumption.
class Llmdog < Formula
  desc "Prepare files for LLM consumption"
  homepage "https://github.com/doganarif/llmdog"
  license "MIT"
  version "$VERSION"

  if Hardware::CPU.arm?
    url "https://github.com/doganarif/llmdog/releases/download/v$VERSION/llmdog_v\${VERSION}_darwin_arm64.tar.gz"
    sha256 "$ARM64_SHA256"
  else
    url "https://github.com/doganarif/llmdog/releases/download/v$VERSION/llmdog_v\${VERSION}_darwin_amd64.tar.gz"
    sha256 "$AMD64_SHA256"
  end

  depends_on "go" => :build

  def install
    # The tarball contains the binary named "llmdog"
    bin.install "llmdog"
  end

  test do
    assert_match "llmdog version", shell_output("\#{bin}/llmdog --version")
  end
end
EOF

echo "Homebrew formula generated: llmdog.rb"
