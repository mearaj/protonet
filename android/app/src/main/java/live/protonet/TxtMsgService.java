package live.protonet;

import android.app.NotificationChannel;
import android.app.PendingIntent;
import android.app.Service;
import android.content.Intent;
import android.os.Build;
import android.os.IBinder;
import android.util.Log;

import androidx.annotation.Nullable;
import androidx.core.app.NotificationCompat;
import androidx.core.app.NotificationManagerCompat;

import org.gioui.GioActivity;

import static android.content.ContentValues.TAG;

public class TxtMsgService extends Service {

    private static final String TXT_MSG_SERVICE_CHANNEL_ID = "TXT_MSG_SERVICE_CHANNEL_ID";
    private static final String TXT_MSG_SERVICE_CHANNEL_NAME= "TXT_MSG_SERVICE_STATUS";
    private static final int TXT_MSG_SERVICE_STATUS_ID = 1;

    @Nullable
    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        Log.d(TAG, "onStartCommand: service started");
        String title = intent.getStringExtra("title");
        String message = intent.getStringExtra("message");
        Log.d(TAG, "showNotification: called with title " + title + " message " + message);
        Intent notificationIntent = new Intent(this, TxtMsgService.class);
        PendingIntent pendingIntent =
                PendingIntent.getActivity(this, 0, new Intent(this, GioActivity.class), PendingIntent.FLAG_UPDATE_CURRENT);
        createNotificationChannel(TXT_MSG_SERVICE_CHANNEL_ID, TXT_MSG_SERVICE_CHANNEL_NAME, NotificationManagerCompat.IMPORTANCE_HIGH);

        NotificationCompat.Builder builder = new NotificationCompat.Builder(this, TXT_MSG_SERVICE_CHANNEL_ID)
                .setSmallIcon(R.mipmap.appicon)
                .setContentTitle(title)
                .setContentText(message)
                .setContentIntent(pendingIntent)
                .setPriority(NotificationCompat.PRIORITY_HIGH);

        // Notification ID cannot be 0.
        startForeground(TXT_MSG_SERVICE_STATUS_ID, builder.build());
        return START_STICKY;
    }

    @Override
    public void onCreate() {
        super.onCreate();
        Log.d(TAG, "onCreate: service created");
    }

    @Override
    public void onDestroy() {
        super.onDestroy();
        Log.d(TAG, "onDestroy: service destroyed");
    }
    private void createNotificationChannel(String id, String name, int importance) {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.O) {
            return;
        }
        NotificationChannel channel = new NotificationChannel(id, name, importance);
        NotificationManagerCompat nm = NotificationManagerCompat.from(this);
        nm.createNotificationChannel(channel);
    }
}
