count=$(git log --oneline |wc -l)
git tag -a v0.0.$count -m "release"
git push origin v0.0.$count

. ./.token

goreleaser --rm-dist

curl -XPOST -umschneider82:$PUT_BINTRAY_SECRET https://api.bintray.com/content/mschneider82/share/share/v0.0.$count/publish
