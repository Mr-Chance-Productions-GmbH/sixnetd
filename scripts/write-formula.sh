#!/bin/bash
# Usage: write-formula.sh VERSION SHA OUTPUT_FILE
set -e
VERSION=$1
SHA=$2
OUTPUT=$3

cat > "$OUTPUT" << FORMULA
class Sixnetd < Formula
  desc "Privileged daemon for the Sixnet VPN client"
  homepage "https://github.com/Mr-Chance-Productions-GmbH/sixnetd"
  url "https://github.com/Mr-Chance-Productions-GmbH/sixnetd/releases/download/v${VERSION}/sixnetd-${VERSION}.tar.gz"
  sha256 "${SHA}"
  version "${VERSION}"

  def install
    bin.install "sixnetd"
  end
end
FORMULA
