# Protonet
A Cross-Platform, Multi-Communication, Serverless (Decentralized) App.

## Important
This version is a complete rewrite from scratch and currently intended for the
[Polygon Hackathon on Devpost](https://buidlit.devpost.com/?ref_feature=challenge&ref_medium=homepage-recommended-hackathons)
There's much improvement from the previous implementation.
The previous version can be found [here.](https://github.com/mearaj/protonet/tree/before-hackathon) 

## About

Protonet is a communication app, based on modern techniques, intended to be a secure way of interaction, whether Text
Messaging or Voice Calling or Video Sharing or Purchasing Crypto Currencies or Using Crypto Currencies for shopping,
should be reliable and secure. It is based upon blockchain compatible wallet. It allows a blockchain user to communicate
with another blockchain user, independent of what blockchain they belong. It also allows auto creation of blockchain
 wallet account.

## Security
Security is the primary concern of this app and hence it takes it seriously.<br>
A user password is required to access the app.<br>
The app doesn't save the user password anywhere.<br>
The user can then either copy/paste private key from clipboard or auto create a new account.<br>
The private key is encrypted with user's password and then stored in the database.<br> 
This makes sure that the original private key is never stored on the user's device and if for any reason(s),<br>
the app's database base is compromised, then the attacker will need your password to view private key(s).

## Security Notes
The app is in very early stage(alpha) and not recommended for production.<br>
If the user looses his password then there's no way he can access the account(s) in this app.<br><br>
The app clears the clipboard after private key is pasted for better security.<br>
There are some OS like Windows 10 where clipboard history can be maintained and enabled.<br>
If that's the case (or similar) with your OS, then make sure it is disabled,<br>
otherwise your private key is vulnerable to attackers, especially in copy/paste private key process.

## Research Resources
https://blog.chain.link/matic-defi-price-feeds/


## [MIT Licensed](LICENSE)
You are free to use any code from this app. You are allowed to make pull request, etc. as well. The intent of this app
was to help open source community and receive help from open source community and anyone interested and also to give a
glimpse of how powerful modern technologies. For third party libraries, please refer to their respective licenses.
Please also refer to [License](LICENSE) file.

## Technologies Glimpse

[Gioui](https://gioui.org/) a modern cross-platform UI Framework in Go language.<br>
[Libp2p](https://github.com/libp2p/go-libp2p) a modern cross-platform Networking Framework / Libraries in Go
language. <br>
There are other libraries used as well. Please refer to source code for that, especially go.mod files.

## Libraries

The app uses many third party open source libraries without which this project wouldn't be possible. For Gui, it mainly
uses [Gioui](https://gioui.org/) <br>
For networking, it mainly uses [Libp2p](https://github.com/libp2p/go-libp2p)

## Supported Platforms

Windows, Mac, Linux, Android, iOS, Modern Browsers<br>
The app is mainly tried on Linux,Android and Modern Browsers, for other platforms you may need to figure out a way.

## Prerequisites

You need to install [Go](https://golang.org/) for your platform

## Running

From commandline/terminal, cd into the root directory of this project, then make sure all the dependencies are
installed. Run `go get ./...`, followed by `go run .`

## Android Build

Make sure [AndroidStudio and AndroidSdk](https://developer.android.com/studio) is installed<br>
Run the following command inside the root directory of the project from terminal/commandline<br>
```gogio -target android .```<br>
The above command will generate protonet.live.apk, then<br>

```
adb devices
adb -s deviceIdFromAbove install protonet.live.apk
```

### Issues

* Error in ... #include<jni.h> No such file or directory Resolution
  ```CGO_CFLAGS="-I${JAVA_HOME}/include -I${JAVA_HOME}/include/linux" go get ./...```
  [Solution](https://stackoverflow.com/questions/56315690/running-go-get-github-com-libp2p-go-libp2p-results-in-error-messages)

### Deployment Refer link below and ```gogio -x -work -appid live.protonet -target android .```

[https://developer.android.com/studio/command-line/apksigner](https://developer.android.com/studio/command-line/apksigner)

# Deployment To Playstore

```
 gogio -buildmode archive -x -work -appid live.protonet -minsdk 22 -version 3 -target android
```

then delete protonet.apk, followed by

```
/pathToZipAlign/zipalign -f 4 /tmpPathFromAbove/app.ap_ protonet.apk
/pathToApkSigner/apksigner sign --ks yourkey.jks protonet.live.apk
```

# Web Assembly

go run gioui.org/cmd/gogio -target js . go get github.com/shurcooL/goexec goexec 'http.ListenAndServe(":8080",
http.FileServer(http.Dir("protonet.live")))'

## Useful References

https://github.com/golang/go/wiki/Modules#can-i-work-entirely-outside-of-vcs-on-my-local-filesystem
https://levelup.gitconnected.com/best-practices-for-webassembly-using-golang-1-15-8dfa439827b8
https://github.com/golang/go/blob/master/misc/wasm/wasm_exec.html
https://gist.github.com/SteveBate/042960baa7a4795c3565

### JNI References

[java_8_jni_type_signatures](https://docs.oracle.com/javase/8/docs/technotes/guides/jni/spec/types.html#type_signatures)

[Pick Image From Android](https://stackoverflow.com/questions/48194733/whats-the-way-to-pick-images-from-gallery-on-android-in-2018/48195899#48195899)

[Encrypt and Decrypt Text Message](https://pkg.go.dev/github.com/decred/dcrd/dcrec/secp256k1/v3#example-package-EncryptDecryptMessage)

