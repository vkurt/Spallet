package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"spallet/core"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/phantasma-io/phantasma-go/pkg/blockchain"
	"github.com/phantasma-io/phantasma-go/pkg/cryptography"
	scriptbuilder "github.com/phantasma-io/phantasma-go/pkg/vm/script_builder"
)

var currentMainDialog dialog.Dialog

func showStakingPage(creds core.Credentials) {
	stakingTab.SetMinSize(fyne.NewSize(0, 525))
	var countdownForSmRw string
	var countdownForCrwn string

	stakeFeeLimit := core.UserSettings.DefaultGasLimit

	// Claiming Kcal stuff
	remainedTimeForKcalGenLabel := widget.NewLabel(fmt.Sprintf("Time until next forging Ritual:\t%v", time.Duration(core.LatestAccountData.RemainedTimeForKcalGen)*time.Second))
	remainedTimeForKcalGenLabel.Wrapping = fyne.TextWrapWord
	unclaimedBalanceLabel := widget.NewLabel(fmt.Sprintf("Earned Sparks:\n%s Kcal", core.FormatBalance(core.LatestAccountData.StakedBalances.Unclaimed, core.KcalDecimals)))
	// unclaimedBalanceLabel.Wrapping = fyne.TextWrapWord
	kcalClaimButton := widget.NewButton("Forge with Sparks", func() {

		// Usage
		askPwdDia(core.UserSettings.AskPwd, creds.Password, mainWindowGui, func(correct bool) {
			fmt.Println("result", correct)
			if !correct {
				return
			}
			// Continue with your code here })
			kcalClaimConfirmLabel := widget.NewLabel("Prepare to collect your earned sparks and enhance your forging power. Are you ready to proceed?")
			kcalClaimConfirmLabel.Wrapping = fyne.TextWrapWord
			confirmButton := widget.NewButton("This is the way", func() {
				keyPair, err := cryptography.FromWIF(creds.Wallets[creds.LastSelectedWallet].WIF)
				if err != nil {
					fyne.CurrentApp().SendNotification(&fyne.Notification{
						Title:   "Transaction Failed",
						Content: fmt.Sprintf("Invalid WIF: %v", err),
					})
					return
				}
				from := keyPair.Address().String()
				expire := time.Now().UTC().Add(time.Second * 300).Unix()
				sb := scriptbuilder.BeginScript()
				sb.AllowGas(from, cryptography.NullAddress().String(), core.UserSettings.GasPrice, stakeFeeLimit)
				sb.CallContract("stake", "Claim", from, from)
				sb.SpendGas(keyPair.Address().String())
				script := sb.EndScript()
				tx := blockchain.NewTransaction(core.UserSettings.NetworkName, core.UserSettings.ChainName, script, uint32(expire), []byte(mainPayload+" Spark Collecting"))
				tx.Sign(keyPair)
				txHex := hex.EncodeToString(tx.Bytes())
				// Start the animation
				startAnimation("forging", "Specky is forging wait a bit....")

				// Here, you can use stopChan if needed later, for example:
				// defer close(stopChan) when you need to ensure it gets closed properly.

				// Send the transaction
				sendTransaction(txHex, creds)
			})
			cancelButton := widget.NewButton("Maybe later", func() {
				currentMainDialog.Hide()
			})
			kcalDiaButtonContainer := container.New(layout.NewCenterLayout())
			diaLockWarning := widget.NewLabelWithStyle("⚠️Dont forget after forging Specky needs to rest his Soul for 24h and your clan need to wait untill it recovers.", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			diaLockWarning.Wrapping = fyne.TextWrapWord
			kcalDiaButtons := container.NewHBox(cancelButton, confirmButton)
			kcalDiaButtonContainer.Objects = []fyne.CanvasObject{kcalDiaButtons}
			kcalClaimDiaRemained := remainedTimeForKcalGenLabel
			kcalClaimDiaDailyProd := widget.NewLabel(fmt.Sprintf("After ritual Specky will create:\t%s Kcal", core.FormatBalance(core.LatestAccountData.KcalDailyProd, core.KcalDecimals)))

			kcalClaimDiaUnclaimed := widget.NewLabel(fmt.Sprintf("Earned Sparks %s Kcal", core.FormatBalance(core.LatestAccountData.StakedBalances.Unclaimed, core.KcalDecimals)))
			kcalConfirmDialog := container.NewBorder(nil, kcalDiaButtonContainer, nil, nil, container.NewVBox(kcalClaimConfirmLabel, kcalClaimDiaRemained, kcalClaimDiaDailyProd, kcalClaimDiaUnclaimed, diaLockWarning))
			d := dialog.NewCustomWithoutButtons("Forge with Sparks", kcalConfirmDialog, mainWindowGui)
			d.Resize(fyne.NewSize(660, 300))
			currentMainDialog = d
			d.Refresh()
			d.Show()

		})

	})
	if core.LatestAccountData.StakedBalances.Unclaimed.Cmp(big.NewInt(0)) > 0 {
		kcalClaimButton.Enable()
	} else {
		kcalClaimButton.Disable()
	}

	//  Staking stufff
	stakedBalancesLabel := widget.NewLabel(fmt.Sprintf("Specky's Soul stash:\n%s Soul (AKA staked Soul)", core.FormatBalance(core.LatestAccountData.StakedBalances.Amount, core.SoulDecimals)))
	stakedBalancesLabel.Wrapping = fyne.TextWrapWord
	accFreeSoulAmount := core.LatestAccountData.FungibleTokens["SOUL"].Amount
	if accFreeSoulAmount == nil {
		accFreeSoulAmount = big.NewInt(0)
	}
	stakingTimeLabel := widget.NewLabel(fmt.Sprintf("Last addition specky's Soul stash:\t%s", time.Unix(int64(core.LatestAccountData.StakedBalances.Time), 0).Format(time.RFC1123)))

	remainedTimeForUnstakeLabel := widget.NewLabel(fmt.Sprintf("Clan's Waiting Period:\t%v", time.Duration(core.LatestAccountData.RemainedTimeForUnstake)*time.Second))
	kcalBoostRateLabel := widget.NewLabel(fmt.Sprintf("Specky's motivation rate\t%v%%", core.LatestAccountData.KcalBoost))
	kcalDailyProdLabel := widget.NewLabel(fmt.Sprintf("Specky Spark Output\t%s Kcal", core.FormatBalance(core.LatestAccountData.KcalDailyProd, core.KcalDecimals)))
	soulInput := widget.NewEntry() //Input for staking/unstaking
	var amount = big.NewInt(0)

	stakeButton := widget.NewButton("Power Up", func() {
		askPwdDia(core.UserSettings.AskPwd, creds.Password, mainWindowGui, func(correct bool) {
			fmt.Println("result", correct)
			if !correct {
				return
			}
			// Continue with your code here })
			// Implement the staking logic here
			userAmount, _ := core.ConvertUserInputToBigInt(soulInput.Text, core.SoulDecimals)

			if userAmount == nil {
				dialog.ShowInformation("Dont trick us...", "Are you trying to trick us?\nHope not please check amount brother/sister.", mainWindowGui)
				return
			} else if userAmount.Cmp(big.NewInt(0)) == 0 {
				dialog.ShowInformation("Dont trick us...", "Are you trying to trick us?\nHope not! please check amount brother/sister.", mainWindowGui)
				return
			} else if userAmount.Cmp(accFreeSoulAmount) > 0 {
				dialog.ShowInformation("Dont trick us...", "Are you trying to trick us?\nHope not! please check amount brother/sister.\nBecause you dont have that amount.", mainWindowGui)
				return
			}
			amount = userAmount

			StakeSoulConfirmLabel := widget.NewLabel("Prepare to strengthen the clan and ignite the stars. By staking your Soul, you'll embark on a journey of honor and reward.\nAre you ready to power up and join the ranks of the elite?")
			StakeSoulConfirmLabel.Wrapping = fyne.TextWrapWord
			confirmButton := widget.NewButton("This is the way", func() {

				keyPair, err := cryptography.FromWIF(creds.Wallets[creds.LastSelectedWallet].WIF)
				if err != nil {
					fyne.CurrentApp().SendNotification(&fyne.Notification{
						Title:   "Transaction Failed",
						Content: fmt.Sprintf("Invalid WIF: %v", err),
					})
					return
				}

				from := keyPair.Address().String()
				expire := time.Now().UTC().Add(time.Second * 300).Unix()
				sb := scriptbuilder.BeginScript()
				sb.AllowGas(from, cryptography.NullAddress().String(), core.UserSettings.GasPrice, stakeFeeLimit)
				sb.CallContract("stake", "Stake", from, amount.String())
				sb.SpendGas(keyPair.Address().String())
				script := sb.EndScript()
				tx := blockchain.NewTransaction(core.UserSettings.NetworkName, core.UserSettings.ChainName, script, uint32(expire), []byte(mainPayload+" Soul Power Up"))
				tx.Sign(keyPair)
				txHex := hex.EncodeToString(tx.Bytes())
				// Start the animation
				startAnimation("fill", "Specky powering up its Soul...")

				// Here, you can use stopChan if needed later, for example:
				// defer close(stopChan) when you need to ensure it gets closed properly.

				// Send the transaction
				sendTransaction(txHex, creds)
			})
			cancelButton := widget.NewButton("Maybe later", func() {
				currentMainDialog.Hide()
			})
			stakeDiaButtonContainer := container.New(layout.NewCenterLayout())

			stakeDiaButtons := container.NewHBox(cancelButton, confirmButton)
			stakeDiaButtonContainer.Objects = []fyne.CanvasObject{stakeDiaButtons}

			currentSparkOutput := widget.NewLabel(fmt.Sprintf("Current Spark Output %v", core.FormatBalance(core.LatestAccountData.KcalDailyProd, core.KcalDecimals)))

			userStakeAmount, _ := core.ConvertUserInputToBigInt(soulInput.Text, core.SoulDecimals)
			extraKcalProduction := core.CalculateKcalDailyProd(core.LatestAccountData.KcalBoost, userStakeAmount, core.KcalProdRate)
			stakeAmountConfirm := widget.NewLabelWithStyle(fmt.Sprintf("You are going to power up your specky to generate %v more Kcal\nwith filling it %s Soul", core.FormatBalance(extraKcalProduction, core.KcalDecimals), core.FormatBalance(amount, core.SoulDecimals)), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			stakeAmountConfirm.Wrapping = fyne.TextWrapWord
			stakeConfirmDialog := container.NewBorder(nil, stakeDiaButtonContainer, nil, nil, container.NewVBox(StakeSoulConfirmLabel, currentSparkOutput, stakeAmountConfirm))
			d := dialog.NewCustomWithoutButtons("Power Up Specky", stakeConfirmDialog, mainWindowGui)
			d.Resize(fyne.NewSize(660, 300))
			currentMainDialog.Hide()
			currentMainDialog = d
			d.Refresh()
			d.Show()

		})
	})

	// Coollecting/unstaking stuff
	collectSoulButton := widget.NewButton("Drain Soul", func() {
		askPwdDia(core.UserSettings.AskPwd, creds.Password, mainWindowGui, func(correct bool) {
			fmt.Println("result", correct)
			if !correct {
				return
			}
			// Continue with your code here })
			userAmount, _ := core.ConvertUserInputToBigInt(soulInput.Text, core.SoulDecimals)

			if userAmount == nil {
				dialog.ShowInformation("Dont trick us...", "Are you trying to trick us?\nHope not please check amount brother/sister.", mainWindowGui)
				return
			} else if userAmount.Cmp(big.NewInt(0)) == 0 {
				dialog.ShowInformation("Dont trick us...", "Are you trying to trick us?\nHope not please check amount brother/sister.", mainWindowGui)
				return
			} else if userAmount.Cmp(core.LatestAccountData.StakedBalances.Amount) > 0 {
				dialog.ShowInformation("Dont trick us...", "Are you trying to trick us?\nHope not please check amount brother/sister.\nBecause you dont have that amount.", mainWindowGui)
				return
			}
			amount = userAmount

			collectSoulConfirmLabel := widget.NewLabel("Souldier, it’s time to reclaim your Soul from Specky, the guardian of Phantasma. By draining your Soul, you will take back your power and honor, but be aware that you will lose some special abilities granted by your staked Soul. Are you ready to retrieve what is rightfully yours and continue your journey?")
			collectSoulConfirmLabel.Wrapping = fyne.TextWrapWord
			confirmButton := widget.NewButton("This is the way", func() {
				keyPair, err := cryptography.FromWIF(creds.Wallets[creds.LastSelectedWallet].WIF)
				if err != nil {
					fyne.CurrentApp().SendNotification(&fyne.Notification{
						Title:   "Transaction Failed",
						Content: fmt.Sprintf("Invalid WIF: %v", err),
					})
					return
				}

				from := keyPair.Address().String()
				expire := time.Now().UTC().Add(time.Second * 300).Unix()
				sb := scriptbuilder.BeginScript()
				sb.AllowGas(from, cryptography.NullAddress().String(), core.UserSettings.GasPrice, stakeFeeLimit)
				sb.CallContract("stake", "Unstake", from, amount.String())
				sb.SpendGas(keyPair.Address().String())
				script := sb.EndScript()
				tx := blockchain.NewTransaction(core.UserSettings.NetworkName, core.UserSettings.ChainName, script, uint32(expire), []byte(mainPayload+" Soul Drain"))
				tx.Sign(keyPair)
				txHex := hex.EncodeToString(tx.Bytes())
				// Start the animation
				startAnimation("drain", "Draining Specky for Soul...")

				// Here, you can use stopChan if needed later, for example:
				// defer close(stopChan) when you need to ensure it gets closed properly.

				// Send the transaction
				sendTransaction(txHex, creds)

			})
			cancelButton := widget.NewButton("Maybe later", func() {
				currentMainDialog.Hide()
			})
			collectSoulDiaButtonContainer := container.New(layout.NewCenterLayout())

			collectSoulDiaButtons := container.NewHBox(cancelButton, confirmButton)
			collectSoulDiaButtonContainer.Objects = []fyne.CanvasObject{collectSoulDiaButtons}

			currentSparkOutput := widget.NewLabel(fmt.Sprintf("Current Spark Output %v", core.FormatBalance(core.LatestAccountData.KcalDailyProd, core.KcalDecimals)))

			userCollectSoulAmount, _ := core.ConvertUserInputToBigInt(soulInput.Text, core.SoulDecimals)
			lessKcalProduction := core.CalculateKcalDailyProd(core.LatestAccountData.KcalBoost, userCollectSoulAmount, core.KcalProdRate)
			collectSoulAmountConfirm := widget.NewLabelWithStyle(fmt.Sprintf("You are going to drain your specky to generate %v less Kcal\nwith draining %s Soul from it.", core.FormatBalance(lessKcalProduction, core.KcalDecimals), core.FormatBalance(amount, core.SoulDecimals)), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
			collectSoulAmountConfirm.Wrapping = fyne.TextWrapWord
			collectSoulConfirmDialog := container.NewBorder(nil, collectSoulDiaButtonContainer, nil, nil, container.NewVBox(collectSoulConfirmLabel, currentSparkOutput, collectSoulAmountConfirm))
			d := dialog.NewCustomWithoutButtons("Drain Specky for Soul", collectSoulConfirmDialog, mainWindowGui)
			d.Resize(fyne.NewSize(660, 300))
			currentMainDialog.Hide()
			currentMainDialog = d
			d.Refresh()
			d.Show()

		})
	})
	//max buttons
	collectMaxSoul := widget.NewButton("Drain Max", func() {
		soulInput.Text = core.FormatBalance(core.LatestAccountData.StakedBalances.Amount, core.SoulDecimals)
		collectSoulButton.Enable()
		if core.LatestAccountData.StakedBalances.Amount.Cmp(accFreeSoulAmount) >= 0 {
			stakeButton.Disable()
		}
		stakingTab.Refresh()
		soulInput.FocusGained()
	})

	stakeMaxSoul := widget.NewButton("Max Power Up", func() {
		{
			soulInput.Text = core.FormatBalance(accFreeSoulAmount, core.SoulDecimals)
			stakeButton.Enable()
			if accFreeSoulAmount.Cmp(core.LatestAccountData.StakedBalances.Amount) >= 0 {
				collectSoulButton.Disable()
			}
			stakingTab.Refresh()
			soulInput.FocusGained()
		}
	})

	// disabling-enabling staking buttons
	stakeButton.Disable()
	collectSoulButton.Disable()

	stakeMaxSoul.Disable()
	collectMaxSoul.Disable()

	if accFreeSoulAmount.Cmp(big.NewInt(0)) > 0 && core.LatestAccountData.StakedBalances.Amount.Cmp(big.NewInt(0)) > 0 && core.LatestAccountData.RemainedTimeForUnstake == 0 {
		stakeMaxSoul.Enable()
		collectMaxSoul.Enable()
	} else if accFreeSoulAmount.Cmp(big.NewInt(0)) > 0 {
		stakeMaxSoul.Enable()
	} else if core.LatestAccountData.StakedBalances.Amount.Cmp(big.NewInt(0)) > 0 && core.LatestAccountData.RemainedTimeForUnstake == 0 {
		collectMaxSoul.Enable()
	}

	soulInput.OnChanged = func(s string) {

		userAmount, err := core.ConvertUserInputToBigInt(s, core.SoulDecimals)
		if err != nil {
			dialog.ShowError(err, mainWindowGui)
		}
		if userAmount == nil {
			amount = big.NewInt(0)
		} else {
			amount = userAmount
		}

		if amount.Cmp(accFreeSoulAmount) <= 0 && amount.Cmp(core.LatestAccountData.StakedBalances.Amount) <= 0 && core.LatestAccountData.RemainedTimeForUnstake == 0 {

			stakeButton.Enable()
			collectSoulButton.Enable()
		} else if amount.Cmp(accFreeSoulAmount) <= 0 {
			stakeButton.Enable()
			collectSoulButton.Disable()
		} else if amount.Cmp(core.LatestAccountData.StakedBalances.Amount) <= 0 && core.LatestAccountData.RemainedTimeForUnstake == 0 {
			collectSoulButton.Enable()
			stakeButton.Disable()
		} else {
			stakeButton.Disable()
			collectSoulButton.Disable()

		}

		stakingTab.Refresh()

	}

	// staking/unstaking buttons+input group
	stakeSoulLabel := widget.NewLabelWithStyle("In the vastness of the galaxy, pledge your spirit, ignite the stars, and ascend to the rank of Soul Master. 🌌✨ Feel the power flowing through you! 💫", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	stakeSoulLabel.Wrapping = fyne.TextWrapWord
	stakeButtonsContainer := container.NewGridWithColumns(2, stakeButton, collectSoulButton)
	maxButtonsContainer := container.NewGridWithColumns(2, stakeMaxSoul, collectMaxSoul)
	stakeCancelBttn := widget.NewButton("Maybe Later", func() { currentMainDialog.Hide() })
	StakeButtonGrid := container.NewVBox(soulInput, stakeButtonsContainer, maxButtonsContainer)

	stakeContainer := container.NewBorder(stakeSoulLabel, stakeCancelBttn, nil, nil, StakeButtonGrid)

	// ****name  Registering things***** between 3-15 chracters+cannot start with numbers+no special chracters+lower case+no space
	nameButtonLabel := "Forge Your Name"
	if core.LatestAccountData.OnChainName != "anonymous" {
		nameButtonLabel = "Reforge Your Name"
	}
	registerNameEntry := widget.NewEntry()

	registerNameButton := widget.NewButton(nameButtonLabel, func() {

		// Usage
		askPwdDia(core.UserSettings.AskPwd, creds.Password, mainWindowGui, func(correct bool) {
			fmt.Println("result", correct)
			if !correct {
				return
			}
			// Continue with your code here })

			response, err := core.Client.LookupName(registerNameEntry.Text)
			if err != nil {
				dialog.ShowError(fmt.Errorf("specky encountered an error while looking availability of this name\n%s", err), mainWindowGui)
				return
			}
			nameTaken := response
			fmt.Println("nameTaken ", nameTaken)
			if len(nameTaken) > 15 {
				dialog.ShowInformation("Name already forged", "Apologies, Souldier.\n\tThe name you seek is already forged by another. Choose wisely, for each name is as unique as Beskar steel. This is the way.", mainWindowGui)
			} else if core.LatestAccountData.OnChainName == "anonymous" {

				registerNameConfirmLabel := widget.NewLabel("Congratulations, souldier.\n\tThe name you have chosen is unique and worthy. By forging this name, you solidify your identity within the clan. Proceed with honor and let your name shine across the galaxy.")
				registerNameConfirmLabel.Wrapping = fyne.TextWrapWord
				registerNameInfoLabel := widget.NewLabelWithStyle(fmt.Sprintf("Specky is ready to forge your name as '%s' ", registerNameEntry.Text), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
				confirmButton := widget.NewButton("This is the way", func() {
					keyPair, err := cryptography.FromWIF(creds.Wallets[creds.LastSelectedWallet].WIF)
					if err != nil {
						fyne.CurrentApp().SendNotification(&fyne.Notification{
							Title:   "Transaction Failed",
							Content: fmt.Sprintf("Invalid WIF: %v", err),
						})
						return
					}

					from := keyPair.Address().String()
					expire := time.Now().UTC().Add(time.Second * 300).Unix()
					sb := scriptbuilder.BeginScript()
					sb.AllowGas(from, cryptography.NullAddress().String(), core.UserSettings.GasPrice, stakeFeeLimit)
					sb.CallContract("account", "RegisterName", from, registerNameEntry.Text)
					sb.SpendGas(keyPair.Address().String())
					script := sb.EndScript()
					tx := blockchain.NewTransaction(core.UserSettings.NetworkName, core.UserSettings.ChainName, script, uint32(expire), []byte(mainPayload+" Name Forging"))
					tx.Sign(keyPair)
					txHex := hex.EncodeToString(tx.Bytes())
					// Start the animation
					startAnimation("forging", "Specky forging a name for you...")

					// Here, you can use stopChan if needed later, for example:
					// defer close(stopChan) when you need to ensure it gets closed properly.

					// Send the transaction
					sendTransaction(txHex, creds)

				})
				cancelButton := widget.NewButton("Maybe later", func() {
					currentMainDialog.Hide()
				})
				registerNameDiaButtonContainer := container.New(layout.NewCenterLayout())

				registerNameDiaButtons := container.NewHBox(cancelButton, confirmButton)
				registerNameDiaButtonContainer.Objects = []fyne.CanvasObject{registerNameDiaButtons}
				collectSoulConfirmDialog := container.NewBorder(nil, registerNameDiaButtonContainer, nil, nil, container.NewVBox(registerNameConfirmLabel, registerNameInfoLabel))
				d := dialog.NewCustomWithoutButtons("Forge a name with Specky", collectSoulConfirmDialog, mainWindowGui)
				d.Resize(fyne.NewSize(660, 300))
				currentMainDialog.Hide()
				currentMainDialog = d
				d.Refresh()
				d.Show()

			} else {
				registerNameConfirmLabel := widget.NewLabel("Congratulations, souldier.\n\tThe name you have chosen is unique and worthy. By forging this name, you solidify your identity within the clan.")
				registerNameConfirmLabel.Wrapping = fyne.TextWrapWord

				registerNameInfoLabel := widget.NewLabelWithStyle(fmt.Sprintf("Specky has already forged a name for you as '%s'.\nThough it is a challenge, Specky will reforge your name to '%s'.\nProceed with honor and let your new name shine across the galaxy.", core.LatestAccountData.OnChainName, registerNameEntry.Text), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
				registerNameInfoLabel.Wrapping = fyne.TextWrapWord
				confirmButton := widget.NewButton("This is the way", func() {
					keyPair, err := cryptography.FromWIF(creds.Wallets[creds.LastSelectedWallet].WIF)
					if err != nil {
						fyne.CurrentApp().SendNotification(&fyne.Notification{
							Title:   "Transaction Failed",
							Content: fmt.Sprintf("Invalid WIF: %v", err),
						})
						return
					}

					from := keyPair.Address().String()
					expire := time.Now().UTC().Add(time.Second * 300).Unix()
					sb := scriptbuilder.BeginScript()
					sb.AllowGas(from, cryptography.NullAddress().String(), core.UserSettings.GasPrice, stakeFeeLimit)
					sb.CallContract("account", "UnregisterName", from)
					sb.CallContract("account", "RegisterName", from, registerNameEntry.Text)
					sb.SpendGas(keyPair.Address().String())
					script := sb.EndScript()
					tx := blockchain.NewTransaction(core.UserSettings.NetworkName, core.UserSettings.ChainName, script, uint32(expire), []byte(mainPayload+" Name Reforging"))
					tx.Sign(keyPair)
					txHex := hex.EncodeToString(tx.Bytes())
					// Start the animation
					startAnimation("forging", "Specky reforging a name for you...")

					// Here, you can use stopChan if needed later, for example:
					// defer close(stopChan) when you need to ensure it gets closed properly.

					// Send the transaction
					sendTransaction(txHex, creds)
				})
				cancelButton := widget.NewButton("Maybe later", func() {
					currentMainDialog.Hide()
				})
				registerNameDiaButtonContainer := container.New(layout.NewCenterLayout())

				registerNameDiaButtons := container.NewHBox(cancelButton, confirmButton)
				registerNameDiaButtonContainer.Objects = []fyne.CanvasObject{registerNameDiaButtons}
				collectSoulConfirmDialog := container.NewBorder(nil, registerNameDiaButtonContainer, nil, nil, container.NewVBox(registerNameConfirmLabel, registerNameInfoLabel))
				d := dialog.NewCustomWithoutButtons("Reforge a name with Specky", collectSoulConfirmDialog, mainWindowGui)
				d.Resize(fyne.NewSize(660, 300))
				currentMainDialog.Hide()
				currentMainDialog = d
				d.Refresh()
				d.Show()

			}

		})
	})
	registerNameLabel := widget.NewLabelWithStyle("Welcome to the clan, Souldier! you are holding a Soul Supply, you're entitled to claim your on-chain name. Declare your name below and join the ranks!", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	registerNameLabel.Wrapping = fyne.TextWrapWord
	registerNameCancelBttn := widget.NewButton("Maybe Later", func() {
		currentMainDialog.Hide()
	})
	registerNameGrid := container.NewVBox(registerNameLabel, registerNameEntry, registerNameButton)
	registerNameContainer := container.NewBorder(nil, registerNameCancelBttn, nil, nil, registerNameGrid)

	registerNameButton.Disable()
	registerNameEntry.OnChanged = func(s string) {
		if len(s) >= 3 {
			registerNameButton.Enable()
			if !core.IsValidName(s) {
				dialog.ShowInformation("Invalid Name", "The name must:\n- Be between 3-15 characters\n- Not start with a number\n- Contain no special characters\n- Be in lower case\n- Contain no spaces", mainWindowGui)
				registerNameButton.Disable()
			}
		} else {
			registerNameButton.Disable()
		}
	}

	// *************Soulmaster things****************
	currentSoulMasterRewardAmountLabel := widget.NewLabel(fmt.Sprintf("Master's Bounty\t%.4f", core.LatestChainStatisticsData.EstSMReward))
	smRwrdButton := widget.NewButton("Honor the Masters", func() {

		smRwrdConfirmLabel := widget.NewLabel("The moment of tribute has arrived. By distributing the Master's Bounty, you recognize the dedication and strength of our Soul Masters. Are you ready to honor their achievements and share the rewards?")
		smRwrdConfirmLabel.Wrapping = fyne.TextWrapWord
		confirmButton := widget.NewButton("This is the way", func() {
			keyPair, err := cryptography.FromWIF(creds.Wallets[creds.LastSelectedWallet].WIF)
			if err != nil {
				fyne.CurrentApp().SendNotification(&fyne.Notification{
					Title:   "Transaction Failed",
					Content: fmt.Sprintf("Invalid WIF: %v", err),
				})
				return
			}
			from := keyPair.Address().String()
			expire := time.Now().UTC().Add(time.Second * 300).Unix()
			sb := scriptbuilder.BeginScript()
			sb.AllowGas(from, cryptography.NullAddress().String(), core.UserSettings.GasPrice, stakeFeeLimit)
			sb.CallContract("stake", "MasterClaim", from)
			sb.SpendGas(keyPair.Address().String())
			script := sb.EndScript()
			tx := blockchain.NewTransaction(core.UserSettings.NetworkName, core.UserSettings.ChainName, script, uint32(expire), []byte(mainPayload+" Master's Bounty Distribution"))
			tx.Sign(keyPair)
			txHex := hex.EncodeToString(tx.Bytes())
			// Start the animation
			startAnimation("forging", "Specky is forging wait a bit....")

			// Here, you can use stopChan if needed later, for example:
			// defer close(stopChan) when you need to ensure it gets closed properly.

			// Send the transaction
			sendTransaction(txHex, creds)
		})
		cancelButton := widget.NewButton("Maybe later", func() {
			currentMainDialog.Hide()
		})
		smRwrdDiaButtonContainer := container.New(layout.NewCenterLayout())

		smRwrdDiaButtons := container.NewHBox(cancelButton, confirmButton)
		smRwrdDiaButtonContainer.Objects = []fyne.CanvasObject{smRwrdDiaButtons}

		smRwrdEstAmountLabel := widget.NewLabelWithStyle(fmt.Sprintf("You are going to distribute %.4f Soul to every eligible master", core.LatestChainStatisticsData.EstSMReward), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

		SmRwrdConfirmDialog := container.NewBorder(nil, smRwrdDiaButtonContainer, nil, nil, container.NewVBox(smRwrdConfirmLabel, smRwrdEstAmountLabel))
		d := dialog.NewCustomWithoutButtons("Distribute Master's Bounty", SmRwrdConfirmDialog, mainWindowGui)
		d.Resize(fyne.NewSize(660, 300))
		currentMainDialog = d
		d.Refresh()
		d.Show()

	})
	smRwrdButton.Disable()
	lastMasterClaimTimeStamp := time.Unix(core.LatestChainStatisticsData.LastMasterClaimTimestamp, 0)
	if core.LatestAccountData.IsEligibleForCurrentSmReward { //remained time for current reward if address eligible for current
		now := time.Now()
		location := now.Location()

		// Get the first day of the next month
		firstDayNextClaim := time.Date(lastMasterClaimTimeStamp.Year(), lastMasterClaimTimeStamp.Month()+1, 1, 0, 0, 0, 0, location)
		// Calculate the last day of this month
		// fmt.Println("firstDayNextMonth", firstDayNextMonth)
		endOfClaimMonth := firstDayNextClaim.AddDate(0, 0, -1).Add(time.Hour*23 + time.Minute*59 + time.Second*59)
		countdown := endOfClaimMonth.Sub(now)

		days := countdown / (24 * time.Hour)
		hours := (countdown % (24 * time.Hour)) / time.Hour
		minutes := (countdown % time.Hour) / time.Minute
		seconds := (countdown % time.Minute) / time.Second
		if countdown <= 0 {
			smRwrdButton.Enable()
			countdownForSmRw = "The time has come"
		} else {
			countdownForSmRw = fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
		}

	} else if core.LatestAccountData.IsSoulMaster { // remained time for next months reward becaue it is not eligible for curent reward
		now := time.Now()
		// Get the first day of the month after the next month
		firstDayNextClaim := time.Date(now.Year(), now.Month()+2, 1, 0, 0, 0, 0, time.UTC)
		// Calculate the last day of the next month
		endOfNextMonth := firstDayNextClaim.AddDate(0, 0, -1).Add(time.Hour*23 + time.Minute*59 + time.Second*59)

		// Calculate the remaining time
		countdown := endOfNextMonth.Sub(now)
		days := int(countdown.Hours()) / 24
		hours := int(countdown.Hours()) % 24
		minutes := int(countdown.Minutes()) % 60
		seconds := int(countdown.Seconds()) % 60
		if countdown <= 0 {
			smRwrdButton.Enable()
			countdownForSmRw = "The time has come"
		} else {
			countdownForSmRw = fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
		}
		// fmt.Println("Time left until the end of next month:", countdownForSmRw)
	}

	// ******Crown things
	crwnRwrdButton := widget.NewButton("Forge the Crowns", func() {
		crownForgeFeeLimit := new(big.Int).Mul(stakeFeeLimit, big.NewInt(10))
		crwnRwrdConfirmLabel := widget.NewLabel("The time has come to forge the legendary Crowns, living entities imbued with ancient secrets and powers. By forging these Crowns eligible Masters will gain extraordinary abilities. Remember if they choose to unlock the Crowns’ secrets through sacrifice, the gained powers will be lost.\nAre you ready to honor our champions with these mystical gifts?\n")
		crwnRwrdConfirmLabel.Wrapping = fyne.TextWrapWord
		confirmButton := widget.NewButton("This is the way", func() {
			keyPair, err := cryptography.FromWIF(creds.Wallets[creds.LastSelectedWallet].WIF)
			if err != nil {
				dialog.ShowError(fmt.Errorf("invalid WIF: %v", err), mainWindowGui)
				return
			}

			from := keyPair.Address().String()
			expire := time.Now().UTC().Add(time.Second * 300).Unix()
			sb := scriptbuilder.BeginScript()
			sb.AllowGas(from, cryptography.NullAddress().String(), core.UserSettings.GasPrice, crownForgeFeeLimit)
			sb.CallContract("gas", "ApplyInflation", from)
			sb.SpendGas(keyPair.Address().String())
			script := sb.EndScript()
			tx := blockchain.NewTransaction(core.UserSettings.NetworkName, core.UserSettings.ChainName, script, uint32(expire), []byte(mainPayload+" Crown Forging"))
			tx.Sign(keyPair)
			txHex := hex.EncodeToString(tx.Bytes())
			// Start the animation
			startAnimation("forging", "Specky is forging wait a bit....")

			// Here, you can use stopChan if needed later, for example:
			// defer close(stopChan) when you need to ensure it gets closed properly.

			// Send the transaction
			sendTransaction(txHex, creds)
		})
		cancelButton := widget.NewButton("Maybe later", func() {
			currentMainDialog.Hide()
		})
		crownFeeWarning := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		crownFeeWarning.Wrapping = fyne.TextWrapWord
		crwnRwrdDiaButtonContainer := container.New(layout.NewCenterLayout())
		feeErr := core.CheckFeeBalance(new(big.Int).Mul(core.UserSettings.GasPrice, crownForgeFeeLimit))
		if crownForgeFeeLimit.Cmp(big.NewInt(100000)) < 0 {
			crownFeeWarning.Text = "⚠️Your Default Fee Limit is too low for forging Crowns!\nMin 10 000"
			confirmButton.Disable()
		} else if feeErr != nil {
			crownFeeWarning.Text = fmt.Sprintf("⚠️You dont have enough Kcal for Specky, forging Crows needs more energy than usual.\n%v", feeErr)
			confirmButton.Disable()
		}

		crwnRwrdDiaButtons := container.NewHBox(cancelButton, confirmButton)
		crwnRwrdDiaButtonContainer.Objects = []fyne.CanvasObject{crwnRwrdDiaButtons}

		crwnRwrdConfirmDialog := container.NewBorder(nil, crwnRwrdDiaButtonContainer, nil, nil, container.NewVBox(crwnRwrdConfirmLabel, crownFeeWarning))
		d := dialog.NewCustomWithoutButtons("Forge Crowns", crwnRwrdConfirmDialog, mainWindowGui)
		d.Resize(fyne.NewSize(660, 300))
		currentMainDialog = d
		d.Refresh()
		d.Show()

	})
	crwnRwrdButton.Disable()

	if core.LatestAccountData.IsEligibleForCurrentCrown { //REmained time for current crown if eligible for current reward
		now := time.Now()
		predettime := time.Unix(core.LatestChainStatisticsData.NextInfTimeStamp, 0)
		countdown := predettime.Sub(now)

		days := countdown / (24 * time.Hour)
		hours := (countdown % (24 * time.Hour)) / time.Hour
		minutes := (countdown % time.Hour) / time.Minute
		seconds := (countdown % time.Minute) / time.Second
		if countdown <= 0 {
			crwnRwrdButton.Enable()
			countdownForCrwn = "The time has come"
		} else {
			countdownForCrwn = fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
		}

	} else if core.LatestAccountData.IsSoulMaster { //remained time for next crown

		now := time.Now()
		predettime := time.Unix(core.LatestChainStatisticsData.NextInfTimeStamp+90*86400, 0)

		countdown := predettime.Sub(now)

		days := countdown / (24 * time.Hour)
		hours := (countdown % (24 * time.Hour)) / time.Hour
		minutes := (countdown % time.Hour) / time.Minute
		seconds := (countdown % time.Minute) / time.Second
		if countdown <= 0 {
			crwnRwrdButton.Enable()
			countdownForCrwn = "The time has come"
		} else {
			countdownForCrwn = fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
		}

	}
	remanedTimeForGetttingCrownLabel := widget.NewLabel(fmt.Sprintf("Coronation after:\t%s", countdownForCrwn))
	remainedTimeForGettingSoulMasterRewardLabel := widget.NewLabel(fmt.Sprintf("Mastery Awaiting Time:\t%s", countdownForSmRw))

	stakeUnstakeBttn := widget.NewButton("Drain/Power Up Specky", func() {
		currentMainDialog = dialog.NewCustomWithoutButtons("Drain/Power Up Specky", stakeContainer, mainWindowGui)
		currentMainDialog.Resize(fyne.NewSize(600, 340))
		currentMainDialog.Show()
	})

	registerNameMainButton := widget.NewButton(nameButtonLabel, func() {
		currentMainDialog = dialog.NewCustomWithoutButtons(nameButtonLabel, registerNameContainer, mainWindowGui)
		currentMainDialog.Resize(fyne.NewSize(600, 340))
		currentMainDialog.Show()
	})

	// Building Gui/tab
	if core.LatestAccountData.StakedBalances.Amount.Cmp(core.SoulMasterThreshold) >= 0 { // if address have soulmaster show this
		soulMasterMessage := widget.NewLabelWithStyle("Souldier, your strength and dedication have earned you a place among the elite. As a Soul Master, your commitment shines brightly, and your valor grants you the honor of Crown rewards. Hold steadfast, for your path is paved with both Mastery and Royal accolades. This is the way! 🚀👑✨", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		soulMasterMessage.Wrapping = fyne.TextWrapWord
		stakingTab.Content = container.NewVBox(
			soulMasterMessage,
			stakedBalancesLabel,
			unclaimedBalanceLabel,
			remainedTimeForKcalGenLabel,
			kcalBoostRateLabel,
			kcalDailyProdLabel,
			remainedTimeForGettingSoulMasterRewardLabel,
			currentSoulMasterRewardAmountLabel,
			remainedTimeForUnstakeLabel,
			remanedTimeForGetttingCrownLabel,
			stakingTimeLabel,
			kcalClaimButton,
			smRwrdButton,
			crwnRwrdButton,

			registerNameMainButton,
			stakeUnstakeBttn,
		)

	} else if core.LatestAccountData.StakedBalances.Amount.Cmp(big.NewInt(0)) > 0 { // if address just a staker show this
		stakerMessageLabel := widget.NewLabelWithStyle("Ascend to Soul Master status, earn the Mastery Reward, and claim your Crown. Strengthen the clan, and let your legacy shine. This is the way! 🚀✨", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		stakerMessageLabel.Wrapping = fyne.TextWrapWord
		stakingTab.Content = container.NewVBox(
			stakerMessageLabel,
			stakedBalancesLabel,
			unclaimedBalanceLabel,
			remainedTimeForKcalGenLabel,
			kcalBoostRateLabel,
			kcalDailyProdLabel,
			remainedTimeForUnstakeLabel,
			stakingTimeLabel,
			kcalClaimButton,
			registerNameMainButton,
			stakeUnstakeBttn,
		)

	} else if accFreeSoulAmount.Cmp(core.MinSoulStake) >= 0 { // if address not staker but have enough soul to stake show this
		notStakerMessage := widget.NewLabelWithStyle("Warrior, you have the strength and the Soul to join the ranks of the honored. Stake your Soul, ignite your destiny, and unlock the path to the Mastery Reward. The galaxy awaits your power and courage. This is the way! 🚀✨", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		notStakerMessage.Wrapping = fyne.TextWrapWord
		stakingTab.Content = container.NewVBox(
			notStakerMessage,
			stakeUnstakeBttn,
		)

	} else {

		soullessMessage := widget.NewLabelWithStyle("Only those with enough Soul power can join the ranks of the honored clan members. Strengthen your spirit and prepare to embark on this prestigious path. The journey to greatness awaits those who possess the courage and strength. This is the way! 🚀✨", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		soullessMessage.Wrapping = fyne.TextWrapWord
		stakingTab.Content = container.NewVBox(
			soullessMessage,
		)

	}
	feeAmount := new(big.Int).Mul(stakeFeeLimit, core.UserSettings.GasPrice)
	err := core.CheckFeeBalance(feeAmount)
	if err != nil {

		kcalClaimButton.Disable()
		smRwrdButton.Disable()
		crwnRwrdButton.Disable()

		stakeUnstakeBttn.Disable()

		registerNameMainButton.Disable()

	}
	x, y := stakingTab.Offset.Components()

	stakingTab.Refresh()
	stakingTab.Offset = fyne.NewPos(x, y)

}
