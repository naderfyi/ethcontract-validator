package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "signaturechecker/docs"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/metachris/eth-go-bindings/erc165"
	"github.com/metachris/eth-go-bindings/erc20"
	"github.com/metachris/eth-go-bindings/erc721"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	apiURL = "https://api.etherscan.io/api?module=contract&action=getsourcecode&address="
	param  = "&apikey=Your_API_KEY"
)

type checkVerificationStatusResponse struct {
	Address  string `json:"address"`
	Verified bool   `json:"verified"`
}
type checkContractStandardResponse struct {
	Address  string `json:"address"`
	Standard string `json:"standard"`
}
type newContractResponse struct {
	Address     string `json:"address"`
	Standard    string `json:"standard"`
	Verified    bool   `json:"verified"`
	Transaction string `json:"transaction"`
	Block       uint64 `json:"block"`
}

// @title Ethereum Contract Checker API
// @version 1.0
// @Description Check if the contract at the given Ethereum address is an ERC-20 or ERC-721 contract, and whether it has been verified on Etherscan.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /

func main() {
	// Connect to the Ethereum node
	client, err := ethclient.Dial("https://mainnet.infura.io/v3/YOUR_ID")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Set up the Gin router and serve the Swagger API documentation files at the "/docs" path
	router := gin.New()
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Redirect the "/docs" path to the documentation page
	router.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/docs/index.html")
	})
	router.Use(gin.Recovery())

	// Load the HTML templates
	router.LoadHTMLGlob("template/*.html")

	router.GET("/getContractsByTime/:startTime/:endTime", func(c *gin.Context) {
		// Get the start and end times from the request parameters
		startTimeParam := c.Param("startTime")
		endTimeParam := c.Param("endTime")

		// Convert the start and end times to time.Time values
		startTime, err := time.Parse("2006-01-02 15:04", startTimeParam)
		if err != nil {
			fmt.Println(err)
			return
		}
		endTime, err := time.Parse("2006-01-02 15:04", endTimeParam)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Convert the start and end times to Unix timestamps
		startTimestamp := startTime.Unix()
		endTimestamp := endTime.Unix()

		// Use an Ethereum block explorer API to get the block numbers corresponding to the start and end timestamps
		apiURL := "https://api.etherscan.io/api?module=block&action=getblocknobytime&timestamp="

		startBlockResponse, err := http.Get(apiURL + strconv.FormatInt(startTimestamp, 10) + "&closest=before" + param)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer startBlockResponse.Body.Close()

		var startBlockResult map[string]interface{}
		err = json.NewDecoder(startBlockResponse.Body).Decode(&startBlockResult)
		if err != nil {
			fmt.Println(err)
			return
		}

		startBlock, err := strconv.ParseInt(startBlockResult["result"].(string), 10, 64)
		if err != nil {
			fmt.Println(err)
			return
		}

		endBlockResponse, err := http.Get(apiURL + strconv.FormatInt(endTimestamp, 10) + "&closest=before" + param)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer endBlockResponse.Body.Close()

		var endBlockResult map[string]interface{}
		err = json.NewDecoder(endBlockResponse.Body).Decode(&endBlockResult)
		if err != nil {
			fmt.Println(err)
			return
		}

		endBlock, err := strconv.ParseInt(endBlockResult["result"].(string), 10, 64)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Get the contracts created between the start and end blocks
		contractAddresses, err := getContracts(startBlock, endBlock, client)
		if err != nil {
			// Handle the error if there was a problem getting the contracts
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		// Return the contract addresses as a JSON response
		c.JSON(http.StatusOK, contractAddresses)
	})

	router.GET("/checkContractStandard/:address", func(c *gin.Context) {
		// Get the address from the request parameters
		address := c.Param("address")

		// Convert the address to a common.Address
		addr := common.HexToAddress(address)
		// Check the contract type of the contract
		standard := checkContractStandard(addr, client)

		// Return the result as a JSON response
		c.JSON(http.StatusOK, checkContractStandardResponse{
			Address:  addr.Hex(),
			Standard: strings.ToUpper(standard),
		})
	})

	router.GET("/getContracts/:startBlock/:endBlock", func(c *gin.Context) {
		// Get the start and end block numbers from the request parameters
		startBlockStr := c.Param("startBlock")
		endBlockStr := c.Param("endBlock")

		// Convert the block numbers to int64
		startBlock, err := strconv.ParseInt(startBlockStr, 10, 64)
		if err != nil {
			// Handle the error if the start block is not a valid number
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start block"})
		}
		endBlock, err := strconv.ParseInt(endBlockStr, 10, 64)
		if err != nil {
			// Handle the error if the end block is not a valid number
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end block"})
		}

		// Get the contracts created between the start and end blocks
		contractAddresses, err := getContracts(startBlock, endBlock, client)
		if err != nil {
			// Handle the error if there was a problem getting the contracts
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}

		// Return the contract addresses as a JSON response
		c.JSON(http.StatusOK, contractAddresses)
	})

	router.GET("/checkVerificationStatus/:address", func(c *gin.Context) {
		// Get the address from the request parameters
		address := c.Param("address")

		// Convert the address to a common.Address
		addr := common.HexToAddress(address)
		// Check the contract verification status for the address
		verified, err := checkVerificationStatus(addr)
		if err != nil {
		}

		// Return the result as a JSON response
		c.JSON(http.StatusOK, checkVerificationStatusResponse{
			Address:  addr.Hex(),
			Verified: verified,
		})
	})

	// Serve the HTML frontend at the root path
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// Redirect to 404 page for any invalid paths
	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.html", nil)
	})

	// Start the server
	router.Run()
}

func IsErc20(addr common.Address, client *ethclient.Client) (bool, error) {

	instance, err := erc20.NewErc20(addr, client)
	if err != nil {
		return false, err
	}

	_, err = instance.Name(nil)
	if err != nil {
		return false, err
	}

	_, err = instance.Symbol(nil)
	if err != nil {
		return false, err
	}

	_, err = instance.Decimals(nil)
	if err != nil {
		return false, err
	}

	_, err = instance.TotalSupply(nil)
	if err != nil {
		return false, err
	}
	return true, nil
}

func IsErc721(addr common.Address, client *ethclient.Client) (bool, error) {

	instance, err := erc721.NewErc721(addr, client)
	if err != nil {
		return false, err
	}

	isErc721, err := instance.SupportsInterface(nil, erc165.InterfaceIdErc165)
	if err != nil {
		return false, err
	}

	return isErc721, nil
}

// @Summary Check the contract verification status for an Ethereum address
// @Description Check if the contract has been verified on Etherscan.
// @ID checkVerificationStatus-contract
// @Accept  json
// @Produce  json
// @Param address path string true "Ethereum address of the contract to checkVerificationStatus"
// @Success 200 {object} checkVerificationStatusResponse
// @Header 200 {string} Token "Contract Address"
// @Router /checkVerificationStatus/{address} [get]
func checkVerificationStatus(addr common.Address) (bool, error) {
	url := apiURL + addr.Hex() + param
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("request failed: %s", resp.Status)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	res, ok := result["result"].([]interface{})
	if !ok {
		return false, fmt.Errorf("failed to parse response")
	}
	if len(res) == 0 {
		return false, nil
	}
	sourceCode, ok := res[0].(map[string]interface{})["SourceCode"].(string)
	if !ok {
		return false, fmt.Errorf("failed to parse response")
	}
	if sourceCode != "" {
		return true, nil
	}
	return false, nil
}

// @Summary Check the contract standard for an Ethereum address
// @Description Check if the contract at the given Ethereum address is an ERC-20 or ERC-721 contract
// @ID checkContractStandard
// @Accept  json
// @Produce  json
// @Param address path string true "Ethereum address of the contract to check Standard"
// @Success 200 {object} checkContractStandardResponse
// @Header 200 {string} Token "Contract Address"
// @Router /checkContractStandard/{address} [get]
func checkContractStandard(addr common.Address, client *ethclient.Client) string {
	isErc20, err := IsErc20(addr, client)
	if err != nil {
	}
	isErc721, err := IsErc721(addr, client)
	if err != nil {
	}

	if isErc20 {
		return "ERC-20"
	} else if isErc721 {
		return "ERC-721"
	}

	return "UNDEFINED"
}

// @Summary Get the new contracts deployed between a start and end block
// @Description Returns a list of new contracts deployed between a start and end block, along with their verification status, standard (ERC-20 or ERC-721), and transaction details.
// @Tags Contracts
// @Accept  json
// @Produce  json
// @Param startBlock path string true "Start block"
// @Param endBlock path string true "End block"
// @Success 200 {array} newContractResponse
// @Router /getContracts/{startBlock}/{endBlock} [get]
func getContracts(startBlock, endBlock int64, client *ethclient.Client) ([]newContractResponse, error) {
	var contractResponses []newContractResponse

	// Iterate through the blocks in the range [startBlock, endBlock]
	for i := startBlock; i <= endBlock; i++ {
		// Get the block by number
		block, err := client.BlockByNumber(context.Background(), big.NewInt(i))
		if err != nil {
			return nil, err
		}

		// Iterate through the transactions in the block
		for _, tx := range block.Transactions() {
			// Check if the transaction is a contract creation transaction
			isContract, contractAddress, err := IsContract(tx, client)
			if err != nil {
				return nil, err
			}
			if isContract {
				// Get the contract type and verification status
				standard := checkContractStandard(contractAddress, client)
				verified, err := checkVerificationStatus(contractAddress)
				if err != nil {
					return nil, err
				}

				// Create a newContractResponse struct for the contract
				contractResponse := newContractResponse{
					Address:     contractAddress.Hex(),
					Standard:    strings.ToUpper(standard),
					Verified:    verified,
					Transaction: tx.Hash().Hex(),
					Block:       block.NumberU64(),
				}

				// Add the struct to the list of contract responses
				contractResponses = append(contractResponses, contractResponse)
			}
		}
	}

	return contractResponses, nil
}

func IsContract(tx *types.Transaction, client *ethclient.Client) (bool, common.Address, error) {
	txReceipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		return false, common.Address{}, err
	}
	if tx.To() == nil && tx.Data() != nil && txReceipt.Status == 1 {
		return true, txReceipt.ContractAddress, nil
	}
	return false, common.Address{}, nil
}
