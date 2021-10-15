package jni

import (
	"gioui.org/app"
	"git.wow.st/gmp/jni"
	log "github.com/sirupsen/logrus"
)

func ShareStringWith(value string) {
	jvm := jni.JVMFor(app.JavaVM())
	appCtx := jni.Object(app.AppContext())
	err := jni.Do(jvm, func(env jni.Env) (err error) {
		loader := jni.ClassLoaderFor(env, appCtx)
		intentClass, err := jni.LoadClass(env, loader, "android/content/Intent")
		if err != nil {
			log.Println("error is ", err)
		}
		intentInitID := jni.GetMethodID(env, intentClass, "<init>", "()V")
		intent, err := jni.NewObject(env, intentClass, intentInitID)
		if err != nil {
			log.Println("error is ", err)
		}
		actionJString := jni.Value(jni.JavaString(env, "android.intent.action.SEND"))
		setActionID := jni.GetMethodID(env, intentClass, "setAction",
			"(Ljava/lang/String;)Landroid/content/Intent;")
		_, err = jni.CallObjectMethod(env, intent, setActionID, actionJString)
		if err != nil {
			log.Println("error is ", err)
		}
		extraTextJString := jni.Value(jni.JavaString(env, "android.intent.extra.TEXT"))
		sharedTextJString := jni.Value(jni.JavaString(env, value))
		putExtraID := jni.GetMethodID(env, intentClass, "putExtra",
			"(Ljava/lang/String;Ljava/lang/String;)Landroid/content/Intent;")
		_, err = jni.CallObjectMethod(env, intent, putExtraID, extraTextJString, sharedTextJString)
		if err != nil {
			log.Println("error is ", err)
		}
		typeJString := jni.Value(jni.JavaString(env, "text/plain"))
		setTypeID := jni.GetMethodID(env, intentClass, "setType",
			"(Ljava/lang/String;)Landroid/content/Intent;")
		_, err = jni.CallObjectMethod(env, intent, setTypeID, typeJString)
		methodFlag := jni.GetMethodID(env, intentClass, "setFlags", "(I)Landroid/content/Intent;")
		intentActivity, err := jni.CallObjectMethod(env, intent, methodFlag, 268435456)
		log.Println("result CallObjectMethod methodFlag err is", err)
		startActivityID := jni.GetMethodID(env, jni.GetObjectClass(env, appCtx), "startActivity",
			"(Landroid/content/Intent;)V",
		)
		err = jni.CallVoidMethod(env, appCtx, startActivityID, jni.Value(intentActivity))
		return
	})
	log.Println(err)
	return
}

func OpenImage() {
	jvm := jni.JVMFor(app.JavaVM())
	appCtx := jni.Object(app.AppContext())
	err := jni.Do(jvm, func(env jni.Env) (err error) {
		loader := jni.ClassLoaderFor(env, appCtx)
		intentClass, err := jni.LoadClass(env, loader, "android/content/Intent")
		intentInitID := jni.GetMethodID(env, intentClass, "<init>", "()V")
		intent, err := jni.NewObject(env, intentClass, intentInitID)
		actionJString := jni.Value(jni.JavaString(env, "android.intent.action.OPEN_DOCUMENT"))
		setActionID := jni.GetMethodID(env, intentClass, "setAction",
			"(Ljava/lang/String;)Landroid/content/Intent;")
		err = jni.CallVoidMethod(env, intent, setActionID, actionJString)
		addCategoryString := jni.Value(jni.JavaString(env, "android.intent.category.OPENABLE"))
		addCategoryID := jni.GetMethodID(env, intentClass, "addCategory",
			"(Ljava/lang/String;)Landroid/content/Intent;")
		err = jni.CallVoidMethod(env, intent, addCategoryID, addCategoryString)
		typeJString := jni.Value(jni.JavaString(env, "image/*"))
		setTypeID := jni.GetMethodID(env, intentClass, "setType",
			"(Ljava/lang/String;)Landroid/content/Intent;")
		err = jni.CallVoidMethod(env, intent, setTypeID, typeJString)
		methodFlag := jni.GetMethodID(env, intentClass, "setFlags", "(I)Landroid/content/Intent;")
		intentActivity, err := jni.CallObjectMethod(env, intent, methodFlag, 268435456)
		log.Println("result CallObjectMethod methodFlag err is", err)
		startActivityID := jni.GetMethodID(env, jni.GetObjectClass(env, appCtx), "startActivity",
			"(Landroid/content/Intent;)V",
		)
		err = jni.CallVoidMethod(env, appCtx, startActivityID, jni.Value(intentActivity))
		log.Printf("Resultis InsideOpenImage %v\nResultis InsideOpenImage %v\n%vResultis InsideOpenImage \n", jni.Value(intentActivity), intentActivity, methodFlag)
		return
	})

	log.Println(err)
	return
}

func ShowNotification(title string, message string) {
	jvm := jni.JVMFor(app.JavaVM())
	appCtx := jni.Object(app.AppContext())
	err := jni.Do(jvm, func(env jni.Env) (err error) {
		//loader := jni.ClassLoaderFor(env, appCtx)
		app := jni.GetObjectClass(env, appCtx)
		startServID := jni.GetMethodID(env, app, "startService", "(Ljava/lang/String;Ljava/lang/String;)V")
		err = jni.CallVoidMethod(env, appCtx, startServID, jni.Value(jni.JavaString(env, title)), jni.Value(jni.JavaString(env, message)))
		if err != nil {
			log.Println("error is ", err)
		}
		return
		//txtMsgSrvCls, err := jni.LoadClass(env, loader, "live/protonet/TxtMsgService")
		//if err != nil {
		//	log.Println("error is ", err)
		//}
		//txtMsgSrvInitId := jni.GetMethodID(env, txtMsgSrvCls, "<init>", "()V")
		//service, err := jni.NewObject(env, txtMsgSrvCls,txtMsgSrvInitId)
		//if err != nil {
		//	log.Println("error is ", err)
		//}
		//update := jni.GetMethodID(env, txtMsgSrvCls, "showNotification", "(Ljava/lang/String;Ljava/lang/String;)V")
		//jtitle := jni.JavaString(env, title)
		//jmessage := jni.JavaString(env, message)
		//return jni.CallVoidMethod(env, service, update, jni.Value(jtitle), jni.Value(jmessage))
	})
	log.Println("ShowNotification......", err)
	return
}
