# ChainDev LIVE - Ecouter l'EventLog de la C-Chain Avalanche

## Initialisation du projet Go 

    cd filterlogs
    go mod init filterlogs
    go mod tidy

**ou**

    cd realtimeeventlogs
    go mod init realtimeeventlogs
    go mod tidy

## Lancer le programme

    go run filterlogs.go

**ou**

    go run realtimeeventlogs.go

## Build le programme et le lancer

**Pour Windows l'extention `.exe` sera ajouté au binaire**

    go build -ldflags="-s -w" -v
    ./filterlogs
    ./realtimeeventlogs
