package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	_ "signaturechecker/docs"

	"github.com/ethereum/go-ethereum/common"
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
	param  = "&apikey=G16FM9WS3JUMQ5G3KTUMYRNNRWFUJWTBA3"
)

type CheckResponse struct {
	Address  string `json:"address"`
	Standard string `json:"standard"`
	Verified bool   `json:"verified"`
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

// @host localhost:8080
// @BasePath /

func main() {
	// Connect to the Ethereum node
	client, err := ethclient.Dial("https://mainnet.infura.io/v3/0be6799a482149d8943b1406039a585c")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Set up the Gin router and serve the Swagger API documentation files at the "/docs" path
	router := gin.New()
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Redirect the "/docs" path to the documentation page
	router.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/docs")
	})
	router.Use(gin.Recovery())

	// Load the HTML templates
	router.LoadHTMLGlob("template/*.html")
	// @Summary Check the contract type for a given address
	// @Description Check if the contract at the given Ethereum address is an ERC-20 or ERC-721 contract, and whether it has been verified on Etherscan.
	// @ID check-contract-type
	// @Accept  json
	// @Produce  json
	// @Param address path string true "The Ethereum contract address to check"
	// @Success 200 {object} main.ContractTypeResponse "The contract type / Verification status for the given address"
	// @Failure 400 {object} main.ErrorResponse "Invalid address"
	// @Router /check/{address} [get]

	router.GET("/check/:address", func(c *gin.Context) {
		// Get the address from the request parameters
		address := c.Param("address")

		// Convert the address to a common.Address
		addr := common.HexToAddress(address)
		// Check the contract type and verification status for the address
		standard, verified := checker_service(addr, client)

		// Return the result as a JSON response
		c.JSON(http.StatusOK, CheckResponse{
			Address:  addr.Hex(),
			Standard: strings.ToUpper(standard),
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

// @Summary Check the contract type and verification status for an Ethereum address
// @Description Check if the contract at the given Ethereum address is an ERC-20 or ERC-721 contract, and whether it has been verified on Etherscan.
// @ID check-contract
// @Accept  json
// @Produce  json
// @Param address path string true "Ethereum address of the contract to check"
// @Success 200 {object} CheckResponse
// @Header 200 {string} Token "Contract Address"
// @Router /check/{address} [get]
func checker_service(addr common.Address, client *ethclient.Client) (string, bool) {
	isErc20, err := IsErc20(addr, client)
	if err != nil {
	}
	isErc721, err := IsErc721(addr, client)
	if err != nil {
	}

	// Add the verification checker code here
	verified, err := check_verification(addr)
	if err != nil {
	}

	if isErc20 {
		return "ERC-20", verified
	} else if isErc721 {
		return "ERC-721", verified
	}

	return "UNDEFINED", verified
}

func check_verification(addr common.Address) (bool, error) {
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
