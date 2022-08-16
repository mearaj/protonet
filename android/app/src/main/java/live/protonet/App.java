// Copyright (c) 2020 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package live.protonet;

import android.app.Application;
import android.content.Intent;
import android.os.Build;
import android.os.Handler;
import android.os.Looper;

import org.gioui.Gio;

public class App extends Application {
    private final static Handler mainHandler = new Handler(Looper.getMainLooper());

    @Override
    public void onCreate() {
        super.onCreate();
        // Load and initialize the Go library.
        Gio.init(this);
       if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
           startForegroundService(new Intent(this, MessageService.class));
       } else {
            startService(new Intent(this, MessageService.class));
        }
    }
}
