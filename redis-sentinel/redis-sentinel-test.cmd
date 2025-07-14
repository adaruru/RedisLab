@echo off
set containers=sentinel-master sentinel-slave1 sentinel-slave2 sentinel1 sentinel2 sentinel3

for %%c in (%containers%) do (
    echo [%%c] 安裝 ntpdate 並執行時間同步
    docker exec %%c bash -c "apt update && apt install -y ntpdate && ntpdate time.google.com"
)

echo ==== sentinel-master ====
docker exec -it sentinel-master redis-cli info replication | findstr "role"

echo ==== sentinel-slave1 ====
docker exec -it sentinel-slave1 redis-cli info replication | findstr "role"

echo ==== sentinel-slave2 ====
docker exec -it sentinel-slave2 redis-cli info replication | findstr "role"

echo ==== master add Test ====
docker exec -it sentinel-master redis-cli set k1 v2

echo ==== slave1 get Test ====
docker exec -it sentinel-slave1 redis-cli get k1

echo ==== slave2 get Test ====
docker exec -it sentinel-slave2 redis-cli get k1

pause