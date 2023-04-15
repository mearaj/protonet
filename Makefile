android:
	ANDROID_SDK_ROOT=${HOME}/Android/Sdk \
	gogio -x -buildmode archive -appid wallet.protonet \
	-ldflags="-v" \
	-minsdk 22 \
	-tags="android" \
	-o ./android/app/libs/protonet.aar -version 1 -target android .

.PHONY: android pkgconfig
