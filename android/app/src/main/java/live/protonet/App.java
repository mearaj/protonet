// Copyright (c) 2020 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package live.protonet;

import android.app.Application;
import android.content.ComponentName;
import android.content.Intent;
import android.os.Handler;
import android.os.Looper;
import org.gioui.Gio;

public class App extends Application {
    private final static String PEER_TAG = "peer";

    private final static Handler mainHandler = new Handler(Looper.getMainLooper());

    @Override
    public void onCreate() {
        super.onCreate();
        // Load and initialize the Go library.
        Gio.init(this);
    }

    public void startService(String title, String message) {
		Intent intent = new Intent(this, TxtMsgService.class);
		intent.putExtra("title", title);
		intent.putExtra("message", message);
		startService(intent);
    }


    public void stopService() {
		Intent intent = new Intent(this, TxtMsgService.class);
		startService(intent);
    }

}
