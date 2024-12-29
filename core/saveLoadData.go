package core

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"

	"github.com/phantasma-io/phantasma-go/pkg/rpc"
)

type WalletSettings struct {
	AskPwd             bool     `json:"ask_pwd"` // security settings
	LgnTmeOut          int      `json:"lgn_tmeout"`
	SendOnly           bool     `json:"send_only"`
	NetworkName        string   `json:"network"` //  network settings
	ChainName          string   `json:"chain"`
	DefaultGasLimit    *big.Int `json:"default_gas_limit"`
	GasLimitSliderMax  float64  `json:"gas_limit_slider_max"`
	GasLimitSliderMin  float64  `json:"gas_limit_slider_min"`
	GasPrice           *big.Int `json:"gas_price"`
	TxExplorerLink     string   `json:"tx_explorer_link"`
	AccExplorerLink    string   `json:"acc_explorer_link"`
	RpcType            string   `json:"rpc_type"` //tried rpc.PhantasmaRpc but not worked
	CustomRpcLink      string   `json:"custom_rpc_link"`
	NetworkType        string   `json:"network_type"`
	DexSlippage        float64  `json:"dex_slippage"` // Dex settings
	DexBaseFeeLimit    *big.Int `json:"dex_base_fee_limit"`
	DexRouteEvaluation string   `json:"dex_route_evaluation"`
	DexDirectRoute     bool     `json:"dex_direct_route"`
}

var Client rpc.PhantasmaRPC
var UserSettings WalletSettings

var defaultSettings = WalletSettings{
	AskPwd:             true, //default security settings
	LgnTmeOut:          5,
	SendOnly:           false,
	NetworkName:        "mainnet", // default network settings
	ChainName:          "main",
	DefaultGasLimit:    big.NewInt(10000),
	GasLimitSliderMax:  100000,
	GasLimitSliderMin:  5000,
	GasPrice:           big.NewInt(100000),
	TxExplorerLink:     "https://explorer.phantasma.info/en/transaction?id=",
	AccExplorerLink:    "https://explorer.phantasma.info/en/address?id=",
	RpcType:            "mainnet",
	NetworkType:        "Mainnet",
	DexSlippage:        5,
	DexBaseFeeLimit:    big.NewInt(30000),
	DexRouteEvaluation: "auto",
	DexDirectRoute:     true,
}

func DefaultSettings() WalletSettings {
	return defaultSettings
}

func SaveAddressBook(adrBk addressBook, pwd string, rootPath string) error {
	filename := filepath.Join(rootPath, "data/essential/addressbook.spallet")

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
		return err
	}

	data, err := json.Marshal(adrBk)
	if err != nil {
		return err
	}

	encryptedData, err := encrypt(data, pwd) // Use password for encryption
	if err != nil {
		return err
	}

	// Save the encrypted data to the file
	return os.WriteFile(filename, []byte(encryptedData), 0600)
}

func LoadAddressBook(path, rawPassword string, rootPath string) (addressBook, error) {
	finalPath := filepath.Join(rootPath, path)
	encryptedData, err := os.ReadFile(finalPath)
	if err != nil {
		SaveAddressBook(UserAddressBook, rawPassword, rootPath)
		return addressBook{Wallets: map[string]Wallet{}}, fmt.Errorf("cant find adddressbook data and saved an empty one:\n%v", err)
	}
	decryptedData, err := decrypt(string(encryptedData), rawPassword)
	if err != nil {
		return addressBook{Wallets: map[string]Wallet{}}, err
	}
	var adrBk addressBook
	err = json.Unmarshal(decryptedData, &adrBk)
	if err != nil {
		return addressBook{Wallets: map[string]Wallet{}}, err
	}
	return adrBk, nil
}

func SaveCredentials(creds Credentials, rootPath string) error {
	filename := filepath.Join(rootPath, "data/essential/credentials.spallet")

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
		return err
	}

	data, err := json.Marshal(creds)
	if err != nil {
		return err
	}

	encryptedData, err := encrypt(data, creds.Password) // Use password for encryption
	if err != nil {
		return err
	}

	// Save the encrypted data to the file
	return os.WriteFile(filename, []byte(encryptedData), 0600)
}

func LoadCredentials(path, rawPassword string, rootPath string) (Credentials, error) {
	finalPath := filepath.Join(rootPath, path)
	encryptedData, err := os.ReadFile(finalPath)
	if err != nil {
		return Credentials{}, err
	}
	decryptedData, err := decrypt(string(encryptedData), rawPassword)
	if err != nil {
		return Credentials{}, err
	}
	var creds Credentials
	err = json.Unmarshal(decryptedData, &creds)
	if err != nil {
		return Credentials{}, err
	}
	return creds, nil
}

// checks if file exists
func FileExists(filePath string) bool {

	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func LoadSettings(path string, rootPath string) error {
	finalPath := filepath.Join(rootPath, path)
	file, err := os.Open(finalPath)
	if err != nil {
		// File doesn't exist, create with default settings
		UserSettings = defaultSettings
		applySettings()
		err = SaveSettings(rootPath)
		if err != nil {
			return err
		}

		return nil
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&UserSettings)
	if err != nil {
		UserSettings = defaultSettings
		err = SaveSettings(rootPath)
		if err != nil {
			return err
		}
	}
	applySettings()
	return nil
}
func LoadDexPools(rootPath string) error {
	path := filepath.Join(rootPath, "data/cache/dexpools.cache")
	file, err := os.Open(path)
	if err != nil {
		// File doesn't exist, create with default settings

		LatestDexPools.PoolKeyCount = 14
		LatestDexPools.PoolList = []string{
			"SOUL_KCAL",
			"BNB_SOUL",
			"RAA_SOUL",
			"RAA_KCAL",
			"BNB_KCAL",
			"GAS_SOUL",
			"ETH_SOUL"}
		err = SaveDexPools(rootPath)
		if err != nil {
			return err
		}

		return nil
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&LatestDexPools)
	if err != nil {
		// File doesn't exist, create with default settings

		LatestDexPools.PoolKeyCount = 14
		LatestDexPools.PoolList = []string{
			"SOUL_KCAL",
			"BNB_SOUL",
			"RAA_SOUL",
			"RAA_KCAL",
			"BNB_KCAL",
			"GAS_SOUL",
			"ETH_SOUL"}
		err = SaveDexPools(rootPath)
		if err != nil {
			return err
		}

		return nil
	}
	return nil
}

func SaveDexPools(rootPath string) error {
	path := filepath.Join(rootPath, "data/cache/dexpools.cache")

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(&LatestDexPools)
	if err != nil {
		return err
	}
	return nil
}

// Save settings to file
func SaveSettings(rootPath string) error {
	filename := filepath.Join(rootPath, "data/essential/settings.spallet")

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(&UserSettings)
	if err != nil {
		return err
	}

	applySettings() // Apply settings immediately after saving
	return nil
}

func applySettings() {
	switch UserSettings.RpcType {
	case "mainnet":
		Client = rpc.NewRPCMainnet()

		fmt.Println("Applied network settings: Mainnet")
	case "testnet":
		Client = rpc.NewRPCTestnet()

		fmt.Println("Applied network settings: Testnet")
	case "custom":
		Client = rpc.NewRPC(UserSettings.CustomRpcLink)

		fmt.Println("Applied network settings: Custom RPC -", UserSettings.CustomRpcLink)
	}

}

func BackupCopyFolder(source, dest string) error {
	entries, err := os.ReadDir(source)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(source, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		fileInfo, err := os.Stat(sourcePath)
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			err = os.MkdirAll(destPath, fileInfo.Mode())
			if err != nil {
				return err
			}

			err = BackupCopyFolder(sourcePath, destPath)
			if err != nil {
				return err
			}
		} else {
			err = BackupCopyFile(sourcePath, destPath)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func BackupCopyFile(source, dest string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
