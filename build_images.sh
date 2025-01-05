# rack api
docker buildx build -t rack/api:1.0 -f api/Dockerfile ./

# rack blockchain
docker buildx build -t rack/blockchain:1.0 -f blockchain/Dockerfile ./
