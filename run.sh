SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd $SCRIPT_DIR
docker run --rm -v .:/app -w /app golang go run generate-result.go
git add result.txt
git commit -m "auto update $(date +"%Y-%m-%d %H:%M:%S")" --author="robot <greenembrace+adlist@gmail.com>"
git push origin master
