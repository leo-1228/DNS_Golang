# !/bin/bash

echo "Building client"
set -e

sed -i 's@"gorm.io/driver/sqlite"@//"gorm.io/driver/sqlite"@g' dbs/cache.go
sed -i 's@"gorm.io/driver/sqlite"@//"gorm.io/driver/sqlite"@g' dbs/setup.go
sed -i 's@"//"github.com/glebarez/sqlite"@"github.com/glebarez/sqlite"@g' dbs/cache.go
sed -i 's@"//"github.com/glebarez/sqlite"@"github.com/glebarez/sqlite"@g' dbs/setup.go

cd cmd/client
CGO_ENABLED=1 go build
cd ..
cd ..

sed -i 's@"//gorm.io/driver/sqlite"@"gorm.io/driver/sqlite"@g' dbs/cache.go
sed -i 's@"//gorm.io/driver/sqlite"@"gorm.io/driver/sqlite"@g' dbs/setup.go
sed -i 's@"github.com/glebarez/sqlite"@//"github.com/glebarez/sqlite"@g' dbs/cache.go
sed -i 's@"github.com/glebarez/sqlite"@//"github.com/glebarez/sqlite"@g' dbs/setup.go

cd cmd/server
CGO_ENABLED=1 GOOS=linux GOARCH=arm CC="zig cc -target native-native-musl" CXX="zig cc -target native-native-musl" go build main.go
cd ..
cd ..
