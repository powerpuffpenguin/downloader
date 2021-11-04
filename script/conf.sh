Target="downloader"
Docker="github.com/powerpuffpenguin/downloader"
Dir=$(cd "$(dirname $BASH_SOURCE)/.." && pwd)
Version="v1.0.0"
View=0
Platforms=(
    darwin/amd64
    windows/amd64
    linux/arm
    linux/amd64
)