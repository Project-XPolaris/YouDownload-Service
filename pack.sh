go build -tags libsqlite3 main.go
rm -r pack-output
mkdir pack-output
cp ./main ./pack-output/youdownload
cp -a ./pack/. ./pack-output