package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"spallet/core"
	"strings"
	"time"
	"unicode/utf8"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type PaginatedResult[T any] struct {
	Page       uint `json:"page"`
	PageSize   uint `json:"pageSize"`
	Total      uint `json:"total"`
	TotalPages uint `json:"totalPages"`
	Result     T    `json:"result"`
}

type AddressTransactionsResult struct {
	Address string              `json:"address"`
	Txs     []TransactionResult `json:"txs"`
}

type TransactionResult struct {
	Hash         string            `json:"hash"`
	ChainAddress string            `json:"chainAddress"`
	Timestamp    uint              `json:"timestamp"`
	BlockHeight  int               `json:"blockHeight"`
	BlockHash    string            `json:"blockHash"`
	Script       string            `json:"script"`
	Payload      string            `json:"payload"`
	Events       []EventResult     `json:"events"`
	State        string            `json:"state"`
	Result       string            `json:"result"`
	Fee          string            `json:"fee"`
	Signatures   []SignatureResult `json:"signatures"`
	Expiration   uint              `json:"expiration"`
}

type EventResult struct {
	Address  string `json:"address"`
	Contract string `json:"contract"`
	Kind     string `json:"kind"`
	Data     string `json:"data"`
}

type SignatureResult struct {
	PubKey    string `json:"pubKey"`
	Signature string `json:"signature"`
}

func historyTabGui(address string, pageForFetch, pageSizeForFetch uint) {

	accountTxes, err := core.Client.GetAddressTransactions(address, int(pageForFetch), int(pageSizeForFetch))
	if err != nil {
		fmt.Println("Error fetching transactions:", err)
		return
	}

	// Marshal JSON data directly from the accountTxes
	jsonTxs, err := json.Marshal(accountTxes)
	if err != nil {
		fmt.Println("Error marshalling transactions:", err)
		return
	}

	// Unmarshal the JSON data into the PaginatedResult
	var paginatedResult PaginatedResult[AddressTransactionsResult]
	err = json.Unmarshal(jsonTxs, &paginatedResult)
	if err != nil {
		fmt.Println("Error unmarshalling transactions:", err)
		return
	}

	transactions := paginatedResult.Result.Txs

	if len(transactions) == 0 {
		historyTab.Content = container.NewCenter(widget.NewLabelWithStyle("This account dont have any activity", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))
	} else {
		displayTransactions(transactions, address)
	}

}

func displayTransactions(transactions []TransactionResult, address string) {

	groupedTransactions := make(map[string][]TransactionResult)
	var sortedDates []string
	today := time.Now().Format("02/01/2006")
	yesterday := time.Now().AddDate(0, 0, -1).Format("02/01/2006")

	for _, tx := range transactions {
		date := time.Unix(int64(tx.Timestamp), 0).Format("02/01/2006")
		if date == today {
			date = "Today"
		} else if date == yesterday {
			date = "Yesterday"
		}
		groupedTransactions[date] = append(groupedTransactions[date], tx)
	}

	for date := range groupedTransactions {
		sortedDates = append(sortedDates, date)
	}
	sort.SliceStable(sortedDates, func(i, j int) bool {
		if sortedDates[i] == "Today" {
			return true
		}
		if sortedDates[j] == "Today" {
			return false
		}
		if sortedDates[i] == "Yesterday" {
			return true
		}
		if sortedDates[j] == "Yesterday" {
			return false
		}
		date1, _ := time.Parse("02/01/2006", sortedDates[i])
		date2, _ := time.Parse("02/01/2006", sortedDates[j])
		return date1.After(date2)
	})

	var transactionItems []fyne.CanvasObject

	for _, date := range sortedDates {
		txGroup := groupedTransactions[date]
		dateLabel := widget.NewLabelWithStyle(date, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		dateContainer := container.NewVBox(dateLabel, widget.NewSeparator())
		transactionItems = append(transactionItems, dateContainer)

		for _, tx := range txGroup {
			var txType string
			var hasIncoming, hasOutgoing, hasExchange, isTokenMint, isLiquidity, isKcalClaim, isSoulStake, isSoulUnstake, isSoulMasterReward bool
			// fmt.Printf("-Events for %v\n", tx.Hash)
			for _, event := range tx.Events {
				// if tx.Hash == "BB889CF449627622AB6B3F75667BDF54A6A1C2D07AD38C63BEA2DEEA286302EF" {
				// 	eventData := hexToASCII(event.Data)
				// 	fmt.Printf("Event\n\tkind %v\n\tcontract %v\n\taddress %v\n\tdata %v\n", event.Kind, event.Contract, event.Address, eventData)
				// }

				// if event.Kind == "ExecutionFailure" {

				// 	isFaulty = true

				// }

				if event.Kind == "TokenSend" && event.Address == address {
					hasOutgoing = true
				} else if event.Kind == "TokenReceive" && event.Address == address {
					hasIncoming = true

				} else if event.Kind == "TokenMint" && event.Contract == "SATRN" && event.Address == address { // not sure it can catch liq mintings but at least tried
					isLiquidity = true
					// hasExchange = false
					// // fmt.Println("found liq mint tx")
					// break
				} else if event.Kind == "TokenMint" && event.Contract == "stake" {
					isKcalClaim = true
					// isTokenMint = false
					// break

				} else if event.Kind == "TokenMint" {
					isTokenMint = true
				} else if event.Kind == "TokenStake" && event.Contract == "SATRN" { // if event contract is Saturn i am acceping it as Exchange tx but not always correct
					hasExchange = true

				} else if event.Kind == "TokenStake" && event.Contract == "stake" {
					isSoulStake = true
				} else if event.Kind == "TokenClaim" && event.Contract == "stake" {
					isSoulUnstake = true
				} else if event.Kind == "MasterClaim" {
					isSoulMasterReward = true
				}

				// eventData := hexToASCII(event.Data)

				// eventDataParts := strings.Split(eventData, " ")
				// fmt.Println("event kind: ", event.Kind)
				// fmt.Println("event contract: ", event.Contract)
				// fmt.Println("event address: ", event.Address)
				// for _, data := range eventDataParts {

				// 	if strings.Contains(data, "SATRN") {
				// 		hasExchange = true
				// 	}

				// 	fmt.Println("Log data: ", data)
				// }

			}

			if hasExchange && !isLiquidity {
				txType = "Exchange"
			} else if hasIncoming {
				txType = "Incoming"
			} else if hasOutgoing {
				txType = "Outgoing"
			} else if isLiquidity {
				txType = "Liquidity"
			} else if isTokenMint && !isKcalClaim {
				txType = "Minting"
			} else if isKcalClaim && !isSoulMasterReward {
				txType = "Collecting"
			} else if isSoulStake {
				txType = "Staking"
			} else if isSoulUnstake && !isSoulMasterReward {
				txType = "Unstaking"
			} else if isSoulMasterReward {
				txType = "SmReward"
			} else {
				txType = "Unknown"
			}

			txState := "Failed"
			if strings.ToUpper(tx.State) == "HALT" {
				txState = "Success"
			}

			timestamp := time.Unix(int64(tx.Timestamp), 0)
			shortTxHash := tx.Hash[:30] + "..."
			payload := hexToASCII(tx.Payload)
			if len(payload) > 15 {
				payload = payload[:12] + "..."
			}
			button := widget.NewButton(fmt.Sprintf("%s - %s - %s - %s\n%s", timestamp.Format("15:04:05"), payload, txState, txType, shortTxHash), func(txHash string) func() {
				return func() {
					explorerURL := fmt.Sprintf("%s%s", core.UserSettings.TxExplorerLink, txHash)
					parsedURL, err := url.Parse(explorerURL)
					if err != nil {
						fmt.Println("Failed to parse URL:", err)
						return
					}
					err = fyne.CurrentApp().OpenURL(parsedURL)
					if err != nil {
						fmt.Println("Failed to open URL:", err)
					}
				}
			}(tx.Hash))

			fmt.Printf("Transaction: %s - %s\n", shortTxHash, txType)
			transactionItems = append(transactionItems, container.NewVBox(button))
		}
	}

	pageContainer := container.NewVScroll(container.NewVBox(transactionItems...))
	historyHeader := widget.NewLabelWithStyle("Recent activity", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	historyHeaderHeight := historyHeader.Size().Height
	pageContainer.SetMinSize(fyne.NewSize(mainWindowLyt.MinSize().Width, mainWindowLyt.Selected().Content.MinSize().Height-historyHeaderHeight))

	historyTab.Content = container.NewBorder(historyHeader, nil, nil, nil, pageContainer)
	historyTab.Content.Refresh()

}
func hexToASCII(hexStr string) string {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		fmt.Println("Error decoding hex string:", err)
		return ""
	}

	// Convert hex bytes to ASCII with spaces for unreadable characters
	var result string
	for _, b := range bytes {
		if utf8.Valid([]byte{b}) && (b >= 32 && b <= 126) { // printable ASCII range
			result += string(b)
		} else {
			result += " " // Add space for unreadable characters
		}
	}
	return result
}
