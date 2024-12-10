package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2/dialog"
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

var client rpc.PhantasmaRPC
var userSettings WalletSettings

var defaultSettings = WalletSettings{
	AskPwd:             true, //default security settings
	LgnTmeOut:          15,
	SendOnly:           false,
	NetworkName:        "mainnet", // default network settings
	ChainName:          "main",
	DefaultGasLimit:    big.NewInt(10000),
	GasLimitSliderMax:  100000,
	GasLimitSliderMin:  10000,
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

func saveAddressBook(adrBk addressBook, pwd string) error {
	filename := "data/essential/addressbook.spallet"

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

func loadAddressBook(path, rawPassword string) (addressBook, error) {

	encryptedData, err := os.ReadFile(path)
	if err != nil {
		saveAddressBook(userAddressBook, rawPassword)
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

func saveCredentials(creds Credentials) error {
	filename := "data/essential/credentials.spallet"

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

func loadCredentials(path, rawPassword string) (Credentials, error) {

	encryptedData, err := os.ReadFile(path)
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
func fileExists(filePath string) bool {

	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func loadSettings(path string) {
	file, err := os.Open(path)
	if err != nil {
		// File doesn't exist, create with default settings
		userSettings = defaultSettings
		applySettings()
		err = saveSettings()
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to save default settings\n%v", err), mainWindowGui)
		}

		return
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&userSettings)
	if err != nil {
		userSettings = defaultSettings
		err = saveSettings()
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to load settings and saved default\n%v", err), mainWindowGui)
		}
	}
	applySettings()

}
func loadDexPools() {
	file, err := os.Open("data/cache/dexpools.cache")
	if err != nil {
		// File doesn't exist, create with default settings

		latestDexPools.PoolKeyCount = 14
		latestDexPools.PoolList = []string{
			"SOUL_KCAL",
			"BNB_SOUL",
			"RAA_SOUL",
			"RAA_KCAL",
			"BNB_KCAL",
			"GAS_SOUL",
			"ETH_SOUL"}
		err = saveDexPools()
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to save default dex pools\n%v", err), mainWindowGui)
		}

		return
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&latestDexPools)
	if err != nil {
		// File doesn't exist, create with default settings

		latestDexPools.PoolKeyCount = 14
		latestDexPools.PoolList = []string{
			"SOUL_KCAL",
			"BNB_SOUL",
			"RAA_SOUL",
			"RAA_KCAL",
			"BNB_KCAL",
			"GAS_SOUL",
			"ETH_SOUL"}
		err = saveDexPools()
		if err != nil {
			dialog.ShowError(fmt.Errorf("failed to load dex pools,saved default dex pools\n%v", err), mainWindowGui)
		}

		return
	}

}

func saveDexPools() error {
	filename := "data/cache/dexpools.cache"

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(&latestDexPools)
	if err != nil {
		return err
	}
	return nil
}

// Save settings to file
func saveSettings() error {
	filename := "data/essential/settings.spallet"

	// Ensure the directory exists
	if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = json.NewEncoder(file).Encode(&userSettings)
	if err != nil {
		return err
	}

	applySettings() // Apply settings immediately after saving
	return nil
}

func applySettings() {
	switch userSettings.RpcType {
	case "mainnet":
		client = rpc.NewRPCMainnet()

		fmt.Println("Applied network settings: Mainnet")
	case "testnet":
		client = rpc.NewRPCTestnet()

		fmt.Println("Applied network settings: Testnet")
	case "custom":
		client = rpc.NewRPC(userSettings.CustomRpcLink)

		fmt.Println("Applied network settings: Custom RPC -", userSettings.CustomRpcLink)
	}

}

func backupCopyFolder(source, dest string) error {
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

			err = backupCopyFolder(sourcePath, destPath)
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
