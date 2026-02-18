#!/bin/bash
set -e

count=$(git log --oneline | wc -l)
tag="v0.0.$count"

git tag -a "$tag" -m "release"
git push
git push --tags

echo "Tag $tag pushed. GitHub Actions will build and release."
