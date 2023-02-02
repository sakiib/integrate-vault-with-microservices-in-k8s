# !/bin/sh

max=100000
for i in `seq 1 $max`
do
    echo "$i"
    #vault kv get kv/"key-$i"
    vault kv put kv/"key-$i" "key-$i"="$(openssl rand -base64  92186)"
done
