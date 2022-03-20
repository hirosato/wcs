
#!/bin/sh
env=$1
echo "**********************************************"
echo "* Building Lambda for '$env' "
echo "***********************************************"
if [ -z "$env" ]
then
    echo "Environment Must not be Empty"
    echo "Usage:"
    echo "sh build.sh <env>"
else
echo "1-Cleaning old builds"
rm bin/main bin/main.zip bin/db.sqlite
echo "2-Building application"
cd src
CGO_ENABLED=1 GOARCH=amd64 GOOS=linux go build -o ../bin/main main.go
cd ../bin
cp ../sqlite/db.sqlite db.sqlite
zip main.zip main db.sqlite
cd ../
echo "4-Deploying to Lambda"
sh push.sh $env
fi