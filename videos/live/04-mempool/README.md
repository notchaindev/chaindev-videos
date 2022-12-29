# ChainDev LIVE - Ecouter la Mempool de la C-Chain Avalanche

## Initialisation du projet Go 

    go mod init mempool
    go mod tidy

## Lancer le programme

    go run mempool.go

## Build le programme et le lancer

**Pour Windows l'extention `.exe` sera ajouté au binaire**

    go build -ldflags="-s -w" -v
    ./mempool
