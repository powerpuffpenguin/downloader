Target="downloader"
Docker="github.com/powerpuffpenguin/downloader"
Dir=$(cd "$(dirname $BASH_SOURCE)/.." && pwd)
Version="v0.0.1"
View=0
Platforms=(
    darwin/amd64
    windows/amd64
    linux/arm
    linux/amd64
)