import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import '../../core/models/notification.dart';
import '../../core/theme/app_theme.dart';
import '../../features/notifications/notification_provider.dart';

/// Bell icon dengan unread count badge.
/// Tap -> navigasi ke /notifications.
class NotificationBell extends StatelessWidget {
  const NotificationBell({super.key});

  @override
  Widget build(BuildContext context) {
    return Consumer<NotificationProvider>(
      builder: (context, provider, _) {
        return Stack(
          clipBehavior: Clip.none,
          children: [
            IconButton(
              icon: const Icon(Icons.notifications_outlined),
              onPressed: () => context.go('/notifications'),
              tooltip: 'Notifikasi',
            ),
            if (provider.unreadCount > 0)
              Positioned(
                right: 4,
                top: 4,
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 5, vertical: 1),
                  decoration: BoxDecoration(
                    color: AppTheme.error,
                    borderRadius: BorderRadius.circular(10),
                    border: Border.all(
                      color: AppTheme.darkBg,
                      width: 1.5,
                    ),
                  ),
                  constraints: const BoxConstraints(
                    minWidth: 18,
                    minHeight: 18,
                  ),
                  child: Text(
                    provider.unreadCount > 99 ? '99+' : '${provider.unreadCount}',
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 10,
                      fontWeight: FontWeight.bold,
                    ),
                    textAlign: TextAlign.center,
                  ),
                ),
              ),
          ],
        );
      },
    );
  }
}

/// Toast/snackbar untuk notifikasi baru.
/// Dipanggil dari dashboard saat WebSocket menerima notification event.
void showNotificationToast(BuildContext context, AppNotification notification) {
  final color = notification.isHighPriority
      ? AppTheme.error
      : notification.isMediumPriority
          ? AppTheme.warning
          : AppTheme.primary;

  ScaffoldMessenger.of(context).showSnackBar(
    SnackBar(
      content: Row(
        children: [
          Icon(
            _iconForType(notification.type),
            color: color,
            size: 20,
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  notification.title,
                  style: const TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 13,
                  ),
                ),
                Text(
                  notification.message,
                  style: const TextStyle(fontSize: 12),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
              ],
            ),
          ),
        ],
      ),
      backgroundColor: AppTheme.darkCard,
      behavior: SnackBarBehavior.floating,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(8),
        side: BorderSide(color: color.withValues(alpha: 0.3)),
      ),
      duration: Duration(
        seconds: notification.isHighPriority ? 5 : 3,
      ),
      action: SnackBarAction(
        label: 'Lihat',
        textColor: color,
        onPressed: () {
          if (notification.route != '/') {
            GoRouter.of(context).go(notification.route);
          }
        },
      ),
    ),
  );
}

IconData _iconForType(String type) {
  return switch (type) {
    'TRANSACTION_CONFIRMED' => Icons.check_circle,
    'TRANSACTION_SUBMITTED' => Icons.send,
    'BLOCK_CONFIRMED' => Icons.view_module,
    'REWARD_EARNED' => Icons.token,
    'BALANCE_UPDATED' => Icons.account_balance_wallet,
    _ => Icons.notifications,
  };
}
