package live.protonet;

import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.app.PendingIntent;
import android.app.Service;
import android.content.Intent;
import android.os.Build;
import android.os.IBinder;
import android.util.Log;

import androidx.core.app.NotificationCompat;
import androidx.core.app.NotificationManagerCompat;

import org.gioui.GioActivity;

import live.protonet.R;

public class MessageService extends Service {
    private static final String PROTONET_MSG_CHANNEL_ID_STR = "PROTONET_MSG_CHANNEL_ID";
    private static final int PROTONET_MSG_CHANNEL_ID_INT = 1;
    private static final String PROTONET_MSG_CHANNEL_NAME = "Protonet Chat Service";

    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        Log.d("MessageService", "onStartCommand: service started");
        Intent notificationIntent = new Intent(this, GioActivity.class);
        PendingIntent pendingIntent;
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
            pendingIntent = PendingIntent.getActivity(this, 0, notificationIntent, PendingIntent.FLAG_IMMUTABLE);
        } else {
            pendingIntent = PendingIntent.getActivity(this, 0, notificationIntent, 0);
        }
        NotificationHelper.newChannel(
                this,
                NotificationManagerCompat.IMPORTANCE_HIGH,
                PROTONET_MSG_CHANNEL_ID_STR,
                PROTONET_MSG_CHANNEL_NAME,
                PROTONET_MSG_CHANNEL_NAME
        );
        NotificationCompat.Builder builder = new NotificationCompat.Builder(this, PROTONET_MSG_CHANNEL_ID_STR);
        Notification notification = builder.setSmallIcon(R.drawable.appicon_notification_icon)
                .setContentTitle("Protonet Chat")
                .setContentText("Protonet Chat Service")
                .setContentIntent(pendingIntent)
                .build();

        // Notification ID cannot be 0.
        startForeground(PROTONET_MSG_CHANNEL_ID_INT, notification);
        return super.onStartCommand(intent, flags, startId);
    }
}