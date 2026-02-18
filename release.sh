count=$(git log --oneline |wc -l)
git tag -a v0.0.$count -m "release"
git push origin v0.0.$count

. ./.token

goreleaser release --clean
