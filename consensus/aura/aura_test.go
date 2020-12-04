package aura

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	lru "github.com/hashicorp/golang-lru"
	"github.com/stretchr/testify/assert"
	"math/big"
	"strings"
	"testing"
	"time"
)

var (
	auraChainConfig *params.AuraConfig
	testBankKey, _  = crypto.GenerateKey()
	testBankAddress = crypto.PubkeyToAddress(testBankKey.PublicKey)
	auraEngine      *Aura
)

func init() {
	authority1, _ := crypto.GenerateKey()
	authority2, _ := crypto.GenerateKey()
	auraChainConfig = &params.AuraConfig{
		Period: 5,
		Epoch:  500,
		Authorities: params.ValidatorSet{
			List: []common.Address{
				testBankAddress,
				crypto.PubkeyToAddress(authority1.PublicKey),
				crypto.PubkeyToAddress(authority2.PublicKey),
			},
		},
		Difficulty: big.NewInt(int64(131072)),
		Signatures: nil,
	}

	db := rawdb.NewMemoryDatabase()
	auraEngine = New(auraChainConfig, db)
	auraEngine.validatorSet = []common.Address{crypto.PubkeyToAddress(authority1.PublicKey), crypto.PubkeyToAddress(authority2.PublicKey)}

	signerFunc := func(account accounts.Account, s string, data []byte) ([]byte, error) {
		return crypto.Sign(crypto.Keccak256(data), testBankKey)
	}
	auraEngine.Authorize(testBankAddress, signerFunc)
}

func TestAura_Finalize(t *testing.T) {
	//gethHeaderWithTransition := "f9026ea0a52747663faa49ea84f1c177f981492b45fcb62d71353a288cb9374ae86381c5a01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d4934794dfa5b73a85f758b0a53992c968ac1a0e0f51ce83a037f2e8ec3a5827d2a65ace4eccfb78d264dae5c9a413bc1e483a675be956c64aa056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421b901000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000090fffffffffffffffffffffffffffffffd1f83232e6580845fca55a39cdb830300018c4f70656e457468657265756d86312e34332e31826c69a00000000000000000000000000000000000000000000000000000000000000000880000000000000000f84c888777281300000000b841ef723ea77bd06dc7153e243636668e35cec750a1b5a63ddbc3dab2d594249f3d4de14c70428031fab1649edd526b61109a60fdf930a198acd291b2267d4eb18f00"
	gethHeaderWithTransition := "f9026ea01b8f7e78fd786ce49f5ab95a69c49a8dc9cfb45c05f3c04f69210114956a04c7a01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d4934794dfa5b73a85f758b0a53992c968ac1a0e0f51ce83a0cacf4ddf8416347cfd1b2c6e4b0cbaf45e400a89e6282d579e83a1ae561c9bf0a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421b901000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000090fffffffffffffffffffffffffffffffe1f83232e6580845fca601b9cdb830300018c4f70656e457468657265756d86312e34332e31826c69a00000000000000000000000000000000000000000000000000000000000000000880000000000000000f84c889f79281300000000b841353f1e6c9a981a1247d31e9647314621ae401d0dd8dbc7f53b880dfb86c75072366176e11eaff127345228773297dbd29cf7b60d4ce433b4462affb15a2334a201"
	//withoutStateMutation := "f9026ea01b8f7e78fd786ce49f5ab95a69c49a8dc9cfb45c05f3c04f69210114956a04c7a01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d4934794dfa5b73a85f758b0a53992c968ac1a0e0f51ce83a0cacf4ddf8416347cfd1b2c6e4b0cbaf45e400a89e6282d579e83a1ae561c9bf0a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421b901000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000090fffffffffffffffffffffffffffffffe1f83232e6580845fca601b9cdb830300018c4f70656e457468657265756d86312e34332e31826c69a00000000000000000000000000000000000000000000000000000000000000000880000000000000000f84c889f79281300000000b841353f1e6c9a981a1247d31e9647314621ae401d0dd8dbc7f53b880dfb86c75072366176e11eaff127345228773297dbd29cf7b60d4ce433b4462affb15a2334a201"
	input, err := hex.DecodeString(gethHeaderWithTransition)
	assert.Nil(t, err)
	var header *types.Header
	err = rlp.Decode(bytes.NewReader(input), &header)
	assert.Nil(t, err)

	db := rawdb.NewMemoryDatabase()
	genspec := &core.Genesis{}
	genspec.MustCommit(db)

	specificEngine := New(auraChainConfig, db)

	// With EIP
	consensusChain, err := core.NewBlockChain(
		specificEngine.db,
		nil,
		&params.ChainConfig{
			ChainID:             big.NewInt(5684),
			HomesteadBlock:      big.NewInt(0),
			DAOForkBlock:        big.NewInt(0),
			DAOForkSupport:      false,
			EIP150Block:         big.NewInt(0),
			EIP150Hash:          common.Hash{},
			EIP155Block:         big.NewInt(0),
			EIP158Block:         big.NewInt(0),
			ByzantiumBlock:      big.NewInt(0),
			ConstantinopleBlock: big.NewInt(0),
			PetersburgBlock:     big.NewInt(0),
			IstanbulBlock:       big.NewInt(0),
			MuirGlacierBlock:    big.NewInt(0),
			YoloV1Block:         nil,
			EWASMBlock:          nil,
			Ethash:              nil,
			Clique:              nil,
			Aura:                auraChainConfig,
		},
		specificEngine,
		vm.Config{},
		func(block *types.Block) bool {
			return false
		},
		nil,
	)
	assert.Nil(t, err)

	stateDbDatabase := state.NewDatabase(db)
	stateDb, err := state.New(common.Hash{}, stateDbDatabase, nil)
	assert.Nil(t, err)
	headerRootString := header.Root.String()
	remoteMerkleRoot := "0xcacf4ddf8416347cfd1b2c6e4b0cbaf45e400a89e6282d579e83a1ae561c9bf0"
	assert.Equal(t, remoteMerkleRoot, headerRootString)

	specificEngine.Finalize(consensusChain, header, stateDb, nil, nil)
	assert.NotNil(t, header)

	headerRootString = header.Root.String()
	assert.Equal(t, remoteMerkleRoot, headerRootString)
}

func TestAura_CheckStep(t *testing.T) {
	currentTime := int64(1602588556)

	t.Run("should return true with no tolerance", func(t *testing.T) {
		allowed, currentTurnTimestamp, nextTurnTimestamp := auraEngine.CheckStep(currentTime, 0)
		assert.True(t, allowed)
		// Period is 5 so next time frame started within -1 from unix time
		assert.Equal(t, currentTime-1, currentTurnTimestamp)
		// Period is 5 so next time frame starts within 4 secs from unix time
		assert.Equal(t, currentTime+4, nextTurnTimestamp)
	})

	t.Run("should return true with small tolerance", func(t *testing.T) {
		allowed, currentTurnTimestamp, nextTurnTimestamp := auraEngine.CheckStep(
			currentTime,
			time.Unix(currentTime, 25).Unix(),
		)
		assert.True(t, allowed)
		// Period is 5 so next time frame started within -1 from unix time
		assert.Equal(t, currentTime-1, currentTurnTimestamp)
		// Period is 5 so next time frame starts within 4 secs from unix time
		assert.Equal(t, currentTime+4, nextTurnTimestamp)
	})

	t.Run("should return false with no tolerance", func(t *testing.T) {
		timeToCheck := currentTime + int64(6)
		allowed, currentTurnTimestamp, nextTurnTimestamp := auraEngine.CheckStep(timeToCheck, 0)
		assert.False(t, allowed)
		assert.Equal(t, timeToCheck-2, currentTurnTimestamp)
		assert.Equal(t, timeToCheck+3, nextTurnTimestamp)
	})

	// If base unixTime is invalid fail no matter what tolerance is
	// If you start sealing before its your turn or you have missed your time frame you should resubmit work
	t.Run("should return false with tolerance", func(t *testing.T) {
		timeToCheck := currentTime + int64(5)
		allowed, currentTurnTimestamp, nextTurnTimestamp := auraEngine.CheckStep(
			timeToCheck,
			time.Unix(currentTime+80, 0).Unix(),
		)
		assert.False(t, allowed)
		assert.Equal(t, timeToCheck-1, currentTurnTimestamp)
		assert.Equal(t, timeToCheck+4, nextTurnTimestamp)
	})
}

func TestAura_CountClosestTurn(t *testing.T) {
	currentTime := int64(1602588556)

	t.Run("should return error, because validator wont be able to seal", func(t *testing.T) {
		randomValidatorKey, err := crypto.GenerateKey()
		assert.Nil(t, err)
		auraChainConfig = &params.AuraConfig{
			Period: 5,
			Epoch:  500,
			Authorities: params.ValidatorSet{
				List: []common.Address{
					crypto.PubkeyToAddress(randomValidatorKey.PublicKey),
				},
			},
			Difficulty: big.NewInt(int64(131072)),
			Signatures: nil,
		}

		db := rawdb.NewMemoryDatabase()
		modifiedAuraEngine := New(auraChainConfig, db)
		closestSealTurnStart, closestSealTurnStop, err := modifiedAuraEngine.CountClosestTurn(
			time.Now().Unix(),
			0,
		)
		assert.Equal(t, errInvalidSigner, err)
		assert.Equal(t, int64(0), closestSealTurnStart)
		assert.Equal(t, int64(0), closestSealTurnStop)
	})

	t.Run("should return current time frame", func(t *testing.T) {
		closestSealTurnStart, closestSealTurnStop, err := auraEngine.CountClosestTurn(currentTime, 0)
		assert.Nil(t, err)
		assert.Equal(t, currentTime-1, closestSealTurnStart)
		assert.Equal(t, currentTime+4, closestSealTurnStop)
	})

	t.Run("should return time frame in future", func(t *testing.T) {
		timeModified := currentTime + 5
		closestSealTurnStart, closestSealTurnStop, err := auraEngine.CountClosestTurn(timeModified, 0)
		assert.Nil(t, err)
		assert.Equal(t, timeModified+9, closestSealTurnStart)
		assert.Equal(t, timeModified+14, closestSealTurnStop)
	})
}

func TestAura_DecodeSeal(t *testing.T) {
	// Block 1 rlp data
	msg4Node0 := "f90241f9023ea02778716827366f0a5479d7a907800d183c57382fa7142b84fbb71db143cf788ca01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d493479470ad1a5fba52e27173d23ad87ad97c9bbe249abfa040cf4430ecaa733787d1a65154a3b9efb560c95d9e324a23b97f0609b539133ba056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421b901000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000090ffffffffffffffffffffffffeceb197b0183222aa980845f6880949cdb830300018c4f70656e457468657265756d86312e34332e31826c69841314e684b84179d277eb6b97d25776793c1a98639d8d41da413fba24c338ee83bff533eac3695a0afaec6df1b77a48681a6a995798964adec1bb406c91b6bbe35f115a828a4101"
	input, err := hex.DecodeString(msg4Node0)
	assert.Nil(t, err)

	var auraHeaders []*types.AuraHeader
	err = rlp.Decode(bytes.NewReader(input), &auraHeaders)
	assert.Nil(t, err)
	assert.NotEmpty(t, auraHeaders)

	for _, header := range auraHeaders {
		// excepted block 1 hash (from parity rpc)
		hashExpected := "0x4d286e4f0dbce8d54b27ea70c211bc4b00c8a89ac67f132662c6dc74d9b294e4"
		assert.Equal(t, hashExpected, header.Hash().String())
		stdHeader := header.TranslateIntoHeader()
		stdHeaderHash := stdHeader.Hash()
		assert.Equal(t, hashExpected, stdHeaderHash.String())
		if header.Number.Int64() == int64(1) {
			signatureForSeal := new(bytes.Buffer)
			encodeSigHeader(signatureForSeal, stdHeader)
			messageHashForSeal := SealHash(stdHeader).Bytes()
			hexutil.Encode(crypto.Keccak256(signatureForSeal.Bytes()))
			pubkey, err := crypto.Ecrecover(messageHashForSeal, stdHeader.Seal[1])

			assert.Nil(t, err)
			var signer common.Address
			copy(signer[:], crypto.Keccak256(pubkey[1:])[12:])
			// 0x70ad1a5fba52e27173d23ad87ad97c9bbe249abf - Block 1 miner
			assert.Equal(t, "0x70ad1a5fba52e27173d23ad87ad97c9bbe249abf", strings.ToLower(signer.Hex()))
		}
	}
}

func TestAura_WaitForNextSealerTurn(t *testing.T) {
	fixedTime := int64(1602697742)
	db := rawdb.NewMemoryDatabase()

	t.Run("Should fail, signer not in validators list", func(t *testing.T) {
		specificEngine := New(&params.AuraConfig{
			Period:      0,
			Epoch:       0,
			Authorities: params.ValidatorSet{},
			Difficulty:  nil,
			Signatures:  nil,
		}, db)
		err := specificEngine.WaitForNextSealerTurn(fixedTime)
		assert.NotNil(t, err)
		assert.Equal(t, errInvalidSigner, err)
	})

	t.Run("should sleep", func(t *testing.T) {
		timeNow := time.Now().Unix()
		closestSealTurnStart, _, err := auraEngine.CountClosestTurn(timeNow, 0)
		assert.Nil(t, err)

		if closestSealTurnStart == timeNow {
			t.Logf("Equal before start")
		}

		err = auraEngine.WaitForNextSealerTurn(timeNow)
		assert.Nil(t, err)
		assert.Equal(t, time.Now().Unix(), closestSealTurnStart)
		fmt.Printf("should wait %d secs", closestSealTurnStart-timeNow)
	})
}

func TestAura_Seal(t *testing.T) {
	// block hex comes from worker test and is extracted due to unit-level of testing Seal
	blockToSignHex := "0xf902c5f9025ca0f0513bebf98c814b3c28ff61746552f74ed65909a3ca4cc3ea5b56dc6021ee3ea01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347940000000000000000000000000000000000000000a02c6e36b7f66da996dc550a19d56c9994626304dc77e459963c1b4dde768020cda02457516422f685ff3338d36c41f3eaa26c35b53f4d485d8d93543c1c4b8bdf6ba0056b23fbba480696b65fe5a59b8f2148a1299103c4f57df839233af2cf4ca2d2b901000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000083020000018347e7c4825208845f84393fb86100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000880000000000000000c0f863f8618080825208943da0ae25cdf7004849e352ba1f8b59ea4b6ebd708203e8801ca00ab99fc4760dfddc35ebd4bf4c4be06e3a2b2d6995fa37b674142c573f7683dda008e2b5c9e9c4597b59d639d7d0aba1b0aa4ddeaf4dceb8b89b914272aa340a1ac0"
	blockBytes, err := hexutil.Decode(blockToSignHex)
	assert.Nil(t, err)
	var block types.Block
	err = rlp.DecodeBytes(blockBytes, &block)
	assert.Nil(t, err)

	// Header should not contain Signature and Step because for now it is not signed
	header := block.Header()
	assert.Empty(t, header.Seal)

	// Max timeout for next turn to start sealing
	timeout := len(auraEngine.config.Authorities.List) * int(auraEngine.config.Period)
	assert.Nil(t, err)

	// Seal the block
	chain := core.BlockChain{}
	resultsChan := make(chan *types.Block)
	stopChan := make(chan struct{})
	timeNow := time.Now().Unix()
	closestSealTurnStart, _, err := auraEngine.CountClosestTurn(timeNow, int64(timeout))
	assert.Nil(t, err)
	waitFor := closestSealTurnStart - timeNow

	if waitFor < 1 {
		waitFor = 0
	}

	t.Logf("Test is waiting for proper turn to start sealing. Waiting: %v secs", waitFor)
	time.Sleep(time.Duration(waitFor) * time.Second)
	err = auraEngine.Seal(&chain, &block, resultsChan, stopChan)

	select {
	case receivedBlock := <-resultsChan:
		assert.Nil(t, err)
		assert.IsType(t, &types.Block{}, receivedBlock)
		header := receivedBlock.Header()
		assert.Len(t, header.Seal, 2)
		signatureForSeal := new(bytes.Buffer)
		encodeSigHeader(signatureForSeal, header)
		messageHashForSeal := SealHash(header).Bytes()
		hexutil.Encode(crypto.Keccak256(signatureForSeal.Bytes()))
		pubkey, err := crypto.Ecrecover(messageHashForSeal, header.Seal[1])

		assert.Nil(t, err)
		var signer common.Address
		copy(signer[:], crypto.Keccak256(pubkey[1:])[12:])

		// Signer should be equal sealer
		assert.Equal(t, strings.ToLower(testBankAddress.String()), strings.ToLower(signer.Hex()))
	case <-time.After(time.Duration(timeout) * time.Second):
		t.Fatalf("Received timeout")

	case receivedStop := <-stopChan:
		t.Fatalf("Received stop, but did not expect this, %v", receivedStop)
	}
}

func TestAura_FromBlock(t *testing.T) {
	invalidBlockRlp := "f902acf902a7a004013562d49a87c65aea12a13f12e63381647705f8e68841126a4620ac13927ca01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347940000000000000000000000000000000000000000a040cf4430ecaa733787d1a65154a3b9efb560c95d9e324a23b97f0609b539133ba056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421b90100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008083035092837a120080845f89a639b861d883010916846765746888676f312e31352e32856c696e7578000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000a00000000000000000000000000000000000000000000000000000000000000000880000000000000000f84c88a5871b1300000000b841030916c553834125cab0ea384ab904bac2e7b7fe2a49fda62a98efb5c1b4fc2c26321fc433fe87d33285f1f696330c8cc94801483544eab72e1f289191466c5b01c0c0"
	input, err := hex.DecodeString(invalidBlockRlp)
	assert.Nil(t, err)
	var standardBlock *types.Block
	err = rlp.Decode(bytes.NewReader(input), &standardBlock)
	assert.Nil(t, err)
	assert.NotNil(t, standardBlock)

	auraBlock := &types.AuraBlock{}
	err = auraBlock.FromBlock(standardBlock)
	assert.Nil(t, err)
}

func TestHeadersFromP2PMessage(t *testing.T) {
	msg4Node0 := "f90241f9023ea02778716827366f0a5479d7a907800d183c57382fa7142b84fbb71db143cf788ca01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d493479470ad1a5fba52e27173d23ad87ad97c9bbe249abfa040cf4430ecaa733787d1a65154a3b9efb560c95d9e324a23b97f0609b539133ba056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421b901000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000090ffffffffffffffffffffffffeceb197b0183222aa980845f6880949cdb830300018c4f70656e457468657265756d86312e34332e31826c69841314e684b84179d277eb6b97d25776793c1a98639d8d41da413fba24c338ee83bff533eac3695a0afaec6df1b77a48681a6a995798964adec1bb406c91b6bbe35f115a828a4101"
	//headers := make([]*types.Header, 0)
	input, err := hex.DecodeString(msg4Node0)
	assert.Nil(t, err)
	msg := p2p.Msg{
		Code:       0x04,
		Size:       uint32(len(input)),
		Payload:    bytes.NewReader(input),
		ReceivedAt: time.Time{},
	}
	headers := HeadersFromP2PMessage(msg)
	assert.Len(t, headers, 1)

	header1 := headers[0]
	auraHeader := types.AuraHeader{}
	err = auraHeader.FromHeader(header1)
	assert.Nil(t, err)
	auraHeaders := make([]*types.AuraHeader, 1)
	auraHeaders[0] = &auraHeader
	encodedBytes, err := rlp.EncodeToBytes(auraHeaders)
	assert.Nil(t, err)
	msg1 := p2p.Msg{
		Code:       0x04,
		Size:       uint32(len(encodedBytes)),
		Payload:    bytes.NewReader(encodedBytes),
		ReceivedAt: time.Time{},
	}
	headers = make([]*types.Header, 0)
	headersFromAura := HeadersFromP2PMessage(msg1)
	assert.Len(t, headersFromAura, 1)
}

func TestAura_VerifySeal(t *testing.T) {
	// Block 1 rlp data
	msg4Node0 := "f90241f9023ea02778716827366f0a5479d7a907800d183c57382fa7142b84fbb71db143cf788ca01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d493479470ad1a5fba52e27173d23ad87ad97c9bbe249abfa040cf4430ecaa733787d1a65154a3b9efb560c95d9e324a23b97f0609b539133ba056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421a056e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421b901000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000090ffffffffffffffffffffffffeceb197b0183222aa980845f6880949cdb830300018c4f70656e457468657265756d86312e34332e31826c69841314e684b84179d277eb6b97d25776793c1a98639d8d41da413fba24c338ee83bff533eac3695a0afaec6df1b77a48681a6a995798964adec1bb406c91b6bbe35f115a828a4101"
	input, err := hex.DecodeString(msg4Node0)
	assert.Nil(t, err)
	var auraHeaders []*types.AuraHeader
	err = rlp.Decode(bytes.NewReader(input), &auraHeaders)
	assert.Nil(t, err)
	assert.NotEmpty(t, auraHeaders)
	var aura Aura
	auraConfig := &params.AuraConfig{
		Period: uint64(5),
		Authorities: params.ValidatorSet{
			List: []common.Address{
				common.HexToAddress("0x70ad1a5fba52e27173d23ad87ad97c9bbe249abf"),
				common.HexToAddress("0xafe443af9d1504de4c2d486356c421c160fdd7b1"),
			},
		},
	}
	aura.config = auraConfig
	var auraSignatures *lru.ARCCache
	auraSignatures, err = lru.NewARC(inmemorySignatures)
	assert.Nil(t, err)
	auraSignatures.Add(0, "0x6f17a2ade9f6daed3968b73514466e07e3c1fef2d6350946e1a12d2b577af0aa")
	aura.signatures = auraSignatures
	for _, header := range auraHeaders {
		// excepted block 1 hash (from parity rpc)
		hashExpected := "0x4d286e4f0dbce8d54b27ea70c211bc4b00c8a89ac67f132662c6dc74d9b294e4"
		assert.Equal(t, hashExpected, header.Hash().String())
		stdHeader := header.TranslateIntoHeader()
		stdHeaderHash := stdHeader.Hash()
		assert.Equal(t, hashExpected, stdHeaderHash.String())
		if header.Number.Int64() == int64(1) {
			signatureForSeal := new(bytes.Buffer)
			encodeSigHeader(signatureForSeal, stdHeader)
			messageHashForSeal := SealHash(stdHeader).Bytes()
			hexutil.Encode(crypto.Keccak256(signatureForSeal.Bytes()))
			pubkey, err := crypto.Ecrecover(messageHashForSeal, stdHeader.Seal[1])
			assert.Nil(t, err)
			err = aura.VerifySeal(nil, stdHeader)
			assert.Nil(t, err)
			var signer common.Address
			copy(signer[:], crypto.Keccak256(pubkey[1:])[12:])
			// 0x70ad1a5fba52e27173d23ad87ad97c9bbe249abf - Block 1 miner
			assert.Equal(t, "0x70ad1a5fba52e27173d23ad87ad97c9bbe249abf", strings.ToLower(signer.Hex()))
		}
	}
}
