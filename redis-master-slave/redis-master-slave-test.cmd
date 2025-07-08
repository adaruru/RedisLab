@echo off

docker exec -it redis-master redis-cli BGSAVE

echo ==== redis-master ====
docker exec -it redis-master redis-cli info replication | findstr "role"

echo ==== redis-slave1 ====
docker exec -it redis-slave1 redis-cli info replication | findstr "role"

echo ==== redis-slave2 ====
docker exec -it redis-slave2 redis-cli info replication | findstr "role"

echo ==== master add Test ====
docker exec -it redis-master redis-cli set k2 v2

echo ==== slave1 get Test ====
docker exec -it redis-slave1 redis-cli get k2

echo ==== slave2 get Test ====
docker exec -it redis-slave2 redis-cli get k2

pause