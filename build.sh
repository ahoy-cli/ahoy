set -e
NAME='ahoy'
COMMIT=$(git rev-parse --short HEAD)
VERSION=$(git describe --tag $COMMIT)
if [ -z "$GOPATH" ]; then
    echo " [Error] You MUST set your \$GOPATH and put this repo within it at \$GOPATH/src/github.com/ahoy-cli/ahoy to build."
    exit 1
fi

IFS=':' read -r -a gopaths <<< "$GOPATH"

dir=`pwd`
for gopath in "${gopaths[@]}"; do
    repo_path="$gopath/src/github.com/ahoy-cli/ahoy"
    if [ "$dir" == "$repo_path" ]; then
        found=true
    fi
done
if [ -z "$found" ]; then
    echo "[Error] This repo should be at one of the following paths:"
    for gopath in "${gopaths[@]}"; do
        echo "$gopath/src/github.com/ahoy-cli/ahoy"
    done
    echo "  but instead it is at:"
    echo "  $dir (Move it)"
    exit 1
fi
go build -ldflags "-X main.version=$VERSION" -o $NAME "$@"
echo "Built ahoy version $VERSION"
