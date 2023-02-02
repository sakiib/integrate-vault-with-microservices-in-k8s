# !/bin/sh

# delete all the vault components
kubectl delete -f integration/vault/

# if you want to delete the pvc's too
kubectl delete pvc -n vault --all

# delete the secret containing vault unseal-keys & root-token
kubectl delete secret -n vault vault-token-keys