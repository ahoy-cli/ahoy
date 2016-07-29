set -e
NAME='ahoy'
COMMIT=$(git rev-parse --short HEAD)
VERSION=$(git describe --tag $COMMIT)
go build -ldflags "-X main.version=$VERSION" -o $NAME "$@"
echo "Built ahoy version $VERSION"
