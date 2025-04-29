# sudoku-api

# DB migrte
# golang-migrate
migrate -path db/migrations -database 'mysql://root:root@tcp(localhost:3307)/sudoku-db' force 1
migrate -path db/migrations -database 'mysql://root:root@tcp(localhost:3307)/sudoku-db' up 1

## docker db
## start
```
docker-compose up -d
```
## stop
```
docker-compose stop mysql


# build run
cd src/Web.Host
go mod tidy
go build ./*.go
go run ./*.go

## memo
```
creatAnswerの実行ファイルの変更でおかしくなる。・