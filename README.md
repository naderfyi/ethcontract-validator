# Ethereum Contract Checker

The Ethereum Contract Checker is an API that checks if the contract at the given Ethereum address is an ERC-20 or ERC-721 contract, and whether it has been verified on Etherscan.

## Installation

To install this project, you need to have Go installed. If you don't have Go installed, you can find it [here](https://golang.org/dl/).

You also need to install the required Go packages:

```bash
go get github.com/ethereum/go-ethereum
go get github.com/gin-gonic/gin
go get github.com/metachris/eth-go-bindings/erc165
go get github.com/metachris/eth-go-bindings/erc20
go get github.com/metachris/eth-go-bindings/erc721
go get github.com/swaggo/files
go get github.com/swaggo/gin-swagger
```
## Usage

To start the server, run the `main.go` file:

```bash
go run main.go
```

The server will start and the API endpoints will be available at `localhost:8080`.

## API Endpoints

- `GET /getContractsByTime/:startTime/:endTime`: Get contracts created between the given start and end times.
- `GET /checkContractStandard/:address`: Check if the contract at the given address is an ERC-20 or ERC-721 contract.
- `GET /getContracts/:startBlock/:endBlock`: Get contracts created between the given start and end blocks.
- `GET /checkVerificationStatus/:address`: Check if the contract at the given address has been verified on Etherscan.

## License

This project is licensed under the Apache 2.0 License.


