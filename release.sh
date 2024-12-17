#!/bin/bash
# Copyright 2024 Nametag Inc.
#
# All information contained herein is the property of Nametag Inc.. The
# intellectual and technical concepts contained herein are proprietary, trade
# secrets, and/or confidential to Nametag, Inc. and may be covered by U.S.
# and Foreign Patents, patents in process, and are protected by trade secret or
# copyright law. Reproduction or distribution, in whole or in part, is
# forbidden except by express written permission of Nametag, Inc.

set -e -o pipefail

# This script copies the CLI from github.com/nametaginc/nt/cli to github.com/nametaginc/cli
# which is where we publish the source.
#
# Because the CLI depends on lots of things we don't publish, it requires some massaging.
# which is the main job of this tool.
#
# TODO: It would be a lot better if we could have more meaningful commit messages etc.
#   but that is a problem for another day. Sorry.

source_root=$(git rev-parse --show-toplevel)
dir=$(mktemp -d)
echo "dir: $dir"

git clone --bare git@github.com:nametaginc/cli "$dir/.git"
cp -r "$source_root/cli/" "$dir"
cp "$source_root"/go.{mod,sum} "$dir"

cd "$dir"
git config core.bare false

# builds internal/api/api.gen.go from the OpenAPI spec, but
# depends on the internal mechanics of generating and validating
# the spec.
rm internal/api/generate.go

# utility for managing VERSION. Not a secret or anything but there is
# no reason to have github.com/Masterminds/semver in our go.mod
rm internal/cli/version_bump.go

# tests depend on lots of internal things e.g. datatest, expect,
# etc, so we can't have them in the open source. :(
find . -name \*_test.go -type f -exec rm {} \;
find . -type d -name testdata -print0 | xargs -0 rm -rf
find . -type d -name recording -print0 | xargs -0 rm -rf

# fix imports
find . -name \*.go -type f -exec sed -i.bak 's|nametaginc/nt/cli|nametaginc/cli|g' {} \;
find . -name \*.bak -exec rm {} \;

# fix go.mod
# We start with the nt root go.mod, remove our special stuff, and
# rewrite the package name. We'll still have loads of extra dependencies,
# but `go mod tidy` can take care of that.
cat go.mod |
	sed 's|nametaginc/nt|nametaginc/cli|g' |
	grep -v -e 'github.com/bas-d/appattest' |
	cat >go.mod~
mv go.mod~ go.mod
go mod tidy

# make sure we can actually build before we commit or push anything
go run github.com/goreleaser/goreleaser/v2@latest --snapshot --clean

version=$(cat internal/cli/VERSION)
echo "version: $version"

# commit & push
git add -A
git commit -m "release version $version from upstream"
git push origin main
git tag "v$version"
git push --tags

# track the next version
(
	cd "$source_root/cli/internal/cli"
	go run version_bump.go
)

# make a release
(
	GITHUB_TOKEN=$(gh auth token) go run github.com/goreleaser/goreleaser/v2@latest release --clean
)
