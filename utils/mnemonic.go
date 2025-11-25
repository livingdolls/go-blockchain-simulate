package utils

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

func GenerateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128) // 128 bits for 12-word mnemonic

	if err != nil {
		return "", err
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", err
	}

	return mnemonic, nil
}

func ValidateMnemonic(mnemonic string) bool {
	return bip39.IsMnemonicValid(mnemonic)
}

// DeriveMasterKey from mnemonic (BIP39 -> seed -> BIP32 master)
func DeriveMasterKey(mnemonic, password string) (*bip32.Key, error) {
	seed := bip39.NewSeed(mnemonic, password)
	masterKey, err := bip32.NewMasterKey(seed)
	if err != nil {
		return nil, err
	}

	return masterKey, nil
}

// DeriveBIP44Account0Child m/44'/60'/0'/0/0
func DeriveChildForEth(master *bip32.Key) (*bip32.Key, error) {
	purpose, err := master.NewChildKey(44 + bip32.FirstHardenedChild)
	if err != nil {
		return nil, fmt.Errorf("derive purpose: %w", err)
	}

	coinType, err := purpose.NewChildKey(60 + bip32.FirstHardenedChild) // 60 for Ethereum
	if err != nil {
		return nil, fmt.Errorf("derive coin type: %w", err)
	}

	account, err := coinType.NewChildKey(0 + bip32.FirstHardenedChild)
	if err != nil {
		return nil, fmt.Errorf("derive account: %w", err)
	}

	change, err := account.NewChildKey(0)
	if err != nil {
		return nil, fmt.Errorf("derive change: %w", err)
	}

	addressIndex, err := change.NewChildKey(0)

	if err != nil {
		return nil, fmt.Errorf("derive address index: %w", err)
	}

	return addressIndex, nil
}

// PubKeyHexFromBIP32Key returns uncompressed public key hex (65 bytes, 0x04||X||Y)
func PubKeyHexFromBIP32Key(key *bip32.Key) string {
	pub := key.PublicKey().Key
	return hex.EncodeToString(pub)
}

// EthAddressFromPublicKeyBytes (pubKey must be uncompressed: 65 bytes, starting with 0x04)
func EthAddressFromPublicKeyBytes(pubKey []byte) (string, error) {
	if len(pubKey) == 0 {
		return "", fmt.Errorf("public key is empty")
	}

	// if pubkey is compressed (33 bytes), decompress it
	var uncompressed []byte
	if len(pubKey) == 65 && pubKey[0] == 0x04 {
		uncompressed = pubKey
	} else {
		// try parse via ecdsa
		// go-ethereum crypto.UnmarshalPubkey expects uncompressed or compressed? It expects uncompressed bytes (65) or serialized?
		// attempt to use crypto.unmarshalPubkey for generic handling
		pk, err := crypto.UnmarshalPubkey(pubKey)
		if err == nil {
			uncompressed = crypto.FromECDSAPub(pk)
		} else {
			// fallback if its 64 bytes without (0x04), prepend 0x04
			if len(pubKey) == 64 {
				uncompressed = append([]byte{0x04}, pubKey...)
			} else {
				return "", fmt.Errorf("unsuported pubkey format: %d", len(pubKey))
			}
		}
	}

	hash := crypto.Keccak256(uncompressed[1:])
	// take last 20 bytes
	addr := hash[len(hash)-20:]
	return "0x" + hex.EncodeToString(addr), nil
}

// PrivateKeyHexFromBIP32 returns private key in hex (32 bytes)
func PrivateKeyHexFromBIP32(key *bip32.Key) string {
	return hex.EncodeToString(key.Key)
}

// convert compressed public key to uncompressed
func EnsureUncompressedPubKey(pubKey []byte) ([]byte, error) {
	if len(pubKey) == 65 && pubKey[0] == 0x04 {
		return pubKey, nil
	}

	pk, err := crypto.DecompressPubkey(pubKey)
	if err != nil {
		return nil, err
	}

	return crypto.FromECDSAPub(pk), nil
}

// Convenience: generate wallet from mnemonic and return mnemonic, privHex, pubHex, address
func GenerateWalletFromMnemonic(mnemonic, passphrase string) (mnemonicOut, privHex, pubHex, address string, err error) {
	if !ValidateMnemonic(mnemonic) {
		err = fmt.Errorf("invalid mnemonic")
		return
	}

	master, err := DeriveMasterKey(mnemonic, passphrase)
	if err != nil {
		return
	}

	child, err := DeriveChildForEth(master)
	if err != nil {
		return
	}

	privHex = PrivateKeyHexFromBIP32(child)
	pubKey := child.PublicKey().Key
	uncompressedPubKey, err := EnsureUncompressedPubKey(pubKey)
	if err != nil {
		return
	}
	pubHex = hex.EncodeToString(uncompressedPubKey)
	addr, err := EthAddressFromPublicKeyBytes(uncompressedPubKey)
	if err != nil {
		return
	}

	return mnemonic, privHex, pubHex, addr, nil
}

func ValidatePrivateKeyMatchesAddress(privateKeyHex, address string) (bool, error) {
	pkBytes, err := hex.DecodeString(privateKeyHex)

	if err != nil {
		return false, fmt.Errorf("err pkBytes %w", err)
	}

	// parse private key to ecdsa
	privKey, err := crypto.ToECDSA(pkBytes)
	if err != nil {
		return false, fmt.Errorf("err to ecdsa %w", err)
	}

	pubKey := privKey.Public().(*ecdsa.PublicKey)

	// dervie address
	deriveAddr := crypto.PubkeyToAddress(*pubKey).Hex()

	//normalize address case
	if strings.EqualFold(deriveAddr, address) {
		return true, nil
	}

	return false, nil
}
