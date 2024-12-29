# Spallet - The Community Wallet for Phantasma Blockchain

## What is Spallet?
Spallet is a community-driven wallet designed for the Phantasma Blockchain. The name Spallet is a playful fusion of Sparky, Specky (mostly Sparky), and Wallet. I wanted a name that's catchy and fun, reflecting the spirit of this wallet.

The goal with Spallet is to bring a touch of fun and creativity to crypto wallets by incorporating small animations, humor, and more, all while reflecting the gaming-oriented nature of the chain. Although I might not be the most experienced developer, I aim to create a wallet that's both engaging and enjoyable to use.

I created Spallet partly because I was dissatisfied with Poltergeist's design and didn't want to see certain names on its license anymore. I hope Spallet can cultivate a new culture within the Phantasma community—who knows what we might achieve together!

## Disclaimer
This wallet is open-source and developed with the guidance of AI. The creator is not a security expert and will not accept any responsibility for potential losses. Use at your own risk!

*Translation of Disclaimer: I’m not a security guru, so if you lose your moon bag, please don’t sue me.*

Spallet uses SHA256 to securely store your wallet data on your hard drive. However, given my limited expertise, please exercise caution and do not rely solely on this security measure.

## Features of Spallet

1. **Bugs**: If you find a bug, consider it a feature!
2. **Nicknames and Badges**: Based on Staked Soul.
3. **Account Migration**: From the manage accounts menu.
4. **Asset Transfer**: Between your accounts.
5. **Send Assets**: To address book recipients.
6. **Master Rewards**: Collect your rewards.
7. **Crown Rewards**: Collect your rewards.
8. **Eligibility Badges**: Show off your status.
9. **Detailed Account Information**: Get all the details about your account.
10. **Chain Statistics**: Stay informed with chain stats.
11. **Staking Information**: Detailed info under the hodling tab.
12. **Adjustable Login Timeout**: Between 3-120 minutes.
13. **Send Assets**: To only known addresses.
14. **Wallet Backup/Restore**: From the restore point menu.
15. **Custom Network Settings**.
16. **No Version Number**: Just enjoy the updates!

Also, there are some other features I might have forgotten to mention!

## What Spallet Doesn't Have

1. **Phantasma Link**.
2. **NFT Pictures and Details**: Due to Go SDK limitations and my limited knowledge.
3. **Token Burning**.

And perhaps some other features I can't recall right now.

## Planned Features
I’ve got some exciting plans for Spallet, like integrating Saturn Dex. However, since this is a fun project, feel free to use it as is. Since it's open-source, you can fork it, continue its development, or even contribute to the codebase!

## Building Instructions

### Desktop

1. **Install Dependencies**
   - Ensure all dependencies required for the project are installed.

2. **Build and Run on Different Platforms**
   - **Windows:**
     - Navigate to the `desktop` folder in your console.
     - To test your changes, run:
       ```sh
       go run .
       ```
     - To build an executable, you have two options:
       - Using Fyne:
         ```sh
         fyne package -os windows -icon icon.png
         ```
       - Using Go:
         ```sh
         go build -ldflags "-H windowsgui"
         ```
   - **macOS:**
     - Navigate to the `desktop` folder in your console.
     - To test your changes, run:
       ```sh
       go run .
       ```
     - To build an executable, you have two options:
       - Using Fyne:
         ```sh
         fyne package -os darwin -icon icon.png
         ```
       - Using Go:
         ```sh
         go build -o myapp-macos
         ```
   - **Linux:**
     - Navigate to the `desktop` folder in your console.
     - To test your changes, run:
       ```sh
       go run .
       ```
     - To build an executable, you have two options:
       - Using Fyne:
         ```sh
         fyne package -os linux -icon icon.png
         ```
       - Using Go:
         ```sh
         go build -o myapp-linux
         ```

### Mobile

1. **Install Dependencies**
   - Ensure all dependencies required for the project are installed.

2. **Build and Run for Mobile**
   - **Android:**
     - Navigate to the `mobile` folder in your console.
     - To build an Android `.APK` file, run:
       ```sh
       fyne package -os android -appID com.spallet.app -icon icon.png
       ```
     - To test your changes on the desktop, run:
       ```sh
       go run .
       ```
   - **iOS:**
     - iOS builds require macOS. Follow these steps on a Mac:
       - Navigate to the `mobile` folder in your console.
       - To build an iOS app, run:
         ```sh
         fyne package -os ios -appID com.spallet.app -icon icon.png
         ```

### Notes

- Ensure you have Go and Fyne installed on your system [Getting Started with Fyne](https://docs.fyne.io/started/).
- For cross-platform builds, refer to the [Fyne cross compiling](https://docs.fyne.io/started/cross-compiling).
- On macOS and Linux, you might need to adjust executable permissions with:
  ```sh
  chmod +x myapp-macos
  chmod +x myapp-linux


