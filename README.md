# Secret Pusher

Helper for pre-filling Hashicorp Vault with secrets from YAML file

## Arguments

| Argument | Environment variable | Default | Description |
|---|---|---|---|
| -secrets | SECRETS | ./secrets.yaml | YAML file with secrets |
| -mount | MOUNT | secret | The path to the KV mount |

## Hashicorp Vault config
Vault client configures with environment variables
- VAULT_ADDR - Vault cluster address; Default https://127.0.0.1:8200
- VAULT_TOKEN - Vault cluster secret token; Required
