package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"github.com/phantasma-io/phantasma-go/pkg/rpc"
)

type WalletSettings struct {
	Network   string `json:"network"`
	CustomRPC string `json:"custom_rpc"`
	Chain     string `json:"chain"`
	AskPwd    bool   `json:"ask_pwd"`
	LgnTmeOut int    `json:"lgn_tmeout"`
	SendOnly  bool   `json:"send_only"`
}

var userSettings WalletSettings
var defaultSettings = WalletSettings{Network: "Mainnet",
	Chain:     "main",
	AskPwd:    true,
	LgnTmeOut: 15,
	SendOnly:  false}

func saveAddressBook(adrBk addressBook, pwd string) error {
	filename := "data/essential/" + "addressbook.spallet"
	data, err := json.Marshal(adrBk)
	if err != nil {
		return err
	}
	encryptedData, err := encrypt(data, pwd) // Use password for encryption
	if err != nil {
		return err
	}
	// fmt.Printf("Saving Encrypted Data: %s\n", encryptedData)
	return os.WriteFile(filename, []byte(encryptedData), 0600)
}

func loadAddressBook(path, rawPassword string) (addressBook, error) {

	encryptedData, err := os.ReadFile(path)
	if err != nil {
		return addressBook{Wallets: map[string]Wallet{}}, err
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
	filename := "data/essential/" + "credentials.spallet"
	data, err := json.Marshal(creds)
	if err != nil {
		return err
	}
	encryptedData, err := encrypt(data, creds.Password) // Use password for encryption
	if err != nil {
		return err
	}
	// fmt.Printf("Saving Encrypted Data: %s\n", encryptedData)
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

// Helper function to load image resources

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
		err = saveSettings()
		if err != nil {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Error",
				Content: "Failed to create default settings file.",
			})
		}
		client = rpc.NewRPCMainnet()
		return
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&userSettings)
	if err != nil || userSettings.Network == "" || userSettings.Chain == "" {
		// Failed to decode or settings are empty, create with default settings
		userSettings = defaultSettings
		err = saveSettings()
		if err != nil {
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Error",
				Content: "Failed to create default settings file.",
			})
		}
		client = rpc.NewRPCMainnet()
	} else {
		applySettings()
	}
}

// Save settings to file
func saveSettings() error {
	file, err := os.Create("data/essential/settings.spallet")
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
	switch userSettings.Network {
	case "Mainnet":
		client = rpc.NewRPCMainnet()
		network = "mainnet"
		fmt.Println("Applied network settings: Mainnet")
	case "Testnet":
		client = rpc.NewRPCTestnet()
		network = "testnet"
		fmt.Println("Applied network settings: Testnet")
	case "Custom":
		client = rpc.NewRPC(userSettings.CustomRPC)
		network = "custom"
		fmt.Println("Applied network settings: Custom RPC -", userSettings.CustomRPC)
	}
	chain = userSettings.Chain
	fmt.Println("Applied chain settings:", chain, network, client)
	askPwd = userSettings.AskPwd
	lgnTmeOutMnt = userSettings.LgnTmeOut
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
