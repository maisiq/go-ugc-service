#!/bin/bash

set -e

mongod --replSet rs0 --bind_ip_all &

until mongosh > /dev/null; do
    echo "❗ Mongosh не отвечает. Повторная проверка через 5 секунд..."
    sleep 5
done

echo "rs.initiate()" | mongosh
echo "🟢 Replica set is ready..."

wait -n
