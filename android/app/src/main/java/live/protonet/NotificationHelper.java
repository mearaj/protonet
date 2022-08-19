package live.protonet;

import android.content.Context;
import android.content.Intent;
import android.app.PendingIntent;
import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.os.Build;
import android.util.Log;

import androidx.core.app.NotificationChannelCompat;
import androidx.core.app.NotificationCompat;
import androidx.core.app.NotificationManagerCompat;

import live.protonet.R;

public class NotificationHelper {
    private final static String tag = "NotificationHelper";
    public static void newChannel(Context ctx, int importance, String channelID, String name, String description) {
        try {
            NotificationChannel channel;
            if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.O) {
                channel = new NotificationChannel(channelID, name, importance);
                channel.setDescription(description);
                NotificationManager notificationManager = ctx.getSystemService(NotificationManager.class);
                notificationManager.createNotificationChannel(channel);
            } else {
                NotificationChannelCompat channelCompat = new NotificationChannelCompat.Builder(channelID,importance).build();
                NotificationManagerCompat notificationManagerCompat = NotificationManagerCompat.from(ctx);
                notificationManagerCompat.createNotificationChannel(channelCompat);
            }

        } catch (Exception e) {
            Log.e("NotificationHelper", "newChannel: ", e);
        }
    }
    public static void sendNotification(Context ctx, String channelID, int notificationID, String title, String text) throws ClassNotFoundException{
        Intent resultIntent = new Intent(ctx, Class.forName("org.gioui.GioActivity"));
        PendingIntent pending;
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.M) {
          pending = PendingIntent.getActivity(ctx, notificationID, resultIntent, PendingIntent.FLAG_IMMUTABLE);
        } else {
          pending = PendingIntent.getActivity(ctx, notificationID, resultIntent, 0);
        }
        NotificationCompat.Builder builder = new NotificationCompat.Builder(ctx, channelID)
                .setContentTitle(title)
                .setSmallIcon(R.drawable.appicon)
                .setContentText(text)
                .setContentIntent(pending)
                .setPriority(Notification.PRIORITY_DEFAULT);
        NotificationManagerCompat notificationManager = NotificationManagerCompat.from(ctx);
        notificationManager.notify(notificationID, builder.build());
    }
    public static void cancelNotification(Context ctx, int notificationID) {
        NotificationManager notificationManager = null;
        if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.M) {
            notificationManager = ctx.getSystemService(NotificationManager.class);
        }
        if (notificationManager != null) {
            notificationManager.cancel(notificationID);
        }
    }
}
