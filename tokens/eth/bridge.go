package eth

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/fsn-dev/crossChain-Bridge/log"
	"github.com/fsn-dev/crossChain-Bridge/tokens"
	"github.com/fsn-dev/crossChain-Bridge/types"
)

// Bridge eth bridge
type Bridge struct {
	*tokens.CrossChainBridgeBase
	Signer types.Signer
}

// NewCrossChainBridge new bridge
func NewCrossChainBridge(isSrc bool) *Bridge {
	if isSrc {
		panic(tokens.ErrTodo)
	}
	return &Bridge{CrossChainBridgeBase: tokens.NewCrossChainBridgeBase(isSrc)}
}

// SetTokenAndGateway set token and gateway config
func (b *Bridge) SetTokenAndGateway(tokenCfg *tokens.TokenConfig, gatewayCfg *tokens.GatewayConfig) {
	b.CrossChainBridgeBase.SetTokenAndGateway(tokenCfg, gatewayCfg)
	b.VerifyChainID()
	b.VerifyTokenCofig()
	b.InitLatestBlockNumber()
}

// VerifyChainID verify chain id
func (b *Bridge) VerifyChainID() {
	tokenCfg := b.TokenConfig
	gatewayCfg := b.GatewayConfig

	networkID := strings.ToLower(tokenCfg.NetID)

	switch networkID {
	case "mainnet", "rinkeby":
	case "custom":
		return
	default:
		panic(fmt.Sprintf("unsupported ethereum network: %v", tokenCfg.NetID))
	}

	var (
		chainID *big.Int
		err     error
	)

	for {
		chainID, err = b.ChainID()
		if err == nil {
			break
		}
		log.Errorf("can not get gateway chainID. %v", err)
		log.Println("retry query gateway", gatewayCfg.APIAddress)
		time.Sleep(3 * time.Second)
	}

	panicMismatchChainID := func() {
		panic(fmt.Sprintf("gateway chainID %v is not %v", chainID, tokenCfg.NetID))
	}

	switch networkID {
	case "mainnet":
		if chainID.Uint64() != 1 {
			panicMismatchChainID()
		}
	case "rinkeby":
		if chainID.Uint64() != 4 {
			panicMismatchChainID()
		}
	default:
		panic("unsupported ethereum network")
	}

	b.Signer = types.MakeSigner("EIP155", chainID)

	log.Info("VerifyChainID succeed", "networkID", networkID, "chainID", chainID)
}

// VerifyTokenCofig verify token config
func (b *Bridge) VerifyTokenCofig() {
	tokenCfg := b.TokenConfig
	if !b.IsValidAddress(tokenCfg.DcrmAddress) {
		log.Fatal("invalid dcrm address", "address", tokenCfg.DcrmAddress)
	}
	if tokenCfg.ContractAddress != "" {
		if !b.IsValidAddress(tokenCfg.ContractAddress) {
			log.Fatal("invalid contract address", "address", tokenCfg.ContractAddress)
		}
		if !b.IsSrc {
			if err := b.VerifyMbtcContractAddress(tokenCfg.ContractAddress); err != nil {
				log.Fatal("wrong contract address", "address", tokenCfg.ContractAddress, "err", err)
			}
		} else if tokenCfg.IsErc20() {
			if err := b.VerifyErc20ContractAddress(tokenCfg.ContractAddress); err != nil {
				log.Fatal("wrong contract address", "address", tokenCfg.ContractAddress, "err", err)
			}
		} else {
			log.Fatal("unsupported type of contract address in source chain, please assign SrcToken.ID (eg. ERC20) in config file", "address", tokenCfg.ContractAddress)
		}
		log.Info("verify contract address pass", "address", tokenCfg.ContractAddress)
	}
}

// InitLatestBlockNumber init latest block number
func (b *Bridge) InitLatestBlockNumber() {
	var (
		tokenCfg   = b.TokenConfig
		gatewayCfg = b.GatewayConfig
		latest     uint64
		err        error
	)

	for {
		latest, err = b.GetLatestBlockNumber()
		if err == nil {
			tokens.SetLatestBlockHeight(latest, b.IsSrc)
			log.Info("get latst block number succeed.", "number", latest, "BlockChain", tokenCfg.BlockChain, "NetID", tokenCfg.NetID)
			break
		}
		log.Error("get latst block number failed.", "BlockChain", tokenCfg.BlockChain, "NetID", tokenCfg.NetID, "err", err)
		log.Println("retry query gateway", gatewayCfg.APIAddress)
		time.Sleep(3 * time.Second)
	}
}
