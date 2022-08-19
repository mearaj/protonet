## Inspiration
Imagine a blockchain user wants to communicate with another blockchain user,
independent of what blockchain they belong to, in a secure and reliable way.
It's not just limited to blockchain user but any user who wants to securely communicate
with another user safely. 'Safely' here means that the message(s) should only be readable
by the sender and recipient and no other entity whether government or spy or hacker or any
sort of organization should be able to view the message.
Yes you got it right, it provides privacy and security and is decentralized.

## What it does
It allows a user to communicate with another user in a secure and private manner. 
A private key is required which can be auto generated or imported by the user. 
The app is password protected so that the user doesn't need to remember his private key(s).
The private keys are encrypted with user's password and the password is saved nowhere
on the device. This is because if the device is stolen or lost, the private keys won't be
viewable unless password is provided, and it also makes hacker difficult to hack those private keys.
The messages are encrypted with the recipient's public key, and they are signed with sender's
private key. This guarantees that only recipient can read the message, and it also verifies
the sender by verifying his signature.

## How we built it
The app primarily uses [Gioui](https://gioui.org/) for cross-platform UI and [Libp2p](https://github.com/libp2p/go-libp2p) 
for networking functionality. There are other libraries used as well. Please refer to source code for that, especially go.mod files.


## Challenges we ran into
There were some questions before the creation of this app on how
it should fit with blockchain and what should be it's place in
the blockchain world and some assumptions were made.
Integrating the chat app directly with the smart contract on blockchain
may be a bad idea because of the following reasons.
1. Assuming most of the users don't want their messages to
be viewable by anyone else and don't want their message history even on blockchain.
2. User shouldn't need to pay just for communication. Imagine a scenario where each message
requires gas fees and transaction cost. This would likely be very disgusting user experience.
3. Messaging should be instantaneous.
4. etc...

## Accomplishments that we're proud of
Currently, the app is completely workable on Windows, MacOS, Android and Iphone.
It's not workable on browser. The UI is not an issue for browser as `gioui` supports
cross-platform development but browser's functionalities has limitations over desktop/devices
functionalities. Currently, we are thinking to implement a QR code in the browser
os that the user can scan it from the mobile. This should then import the entire chat
from user's device into the browser in a safe, secure and reliable manner.

## What we learned
We learned how blockchain technology works and why it is revolutionary.
Golang is not mostly used for UI, but there are helpful libraries and frameworks
that allows to create cross-platform gui and although golang might not be used for gui, but
it is capable and can be a good choice!

## What's next for Protonet
1. Voice Messaging
2. Voice Call
3. Video Call
4. Broadcasting
5. Crypto Exchange Rates (Using Chainlink)
6. A pure decentralized exchange on polygon blockchain, unlike most that exists today which
are really not implemented in a pure way because they are still under control of
big organizations.
An example that purely defines it, is [SmartExchange](https://vbn.aau.dk/en/publications/smartexchange-decentralised-trustless-cryptocurrency-exchange).
7. An app that can provide web3 functionality since it is based upon the same technology that is used
by the blockchain.