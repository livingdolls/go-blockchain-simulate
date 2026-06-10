import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:provider/provider.dart';
import 'package:intl/intl.dart';
import '../../core/models/notification.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/widgets/app_widgets.dart';
import 'notification_provider.dart';

/// Screen daftar notifikasi dengan filter all/unread.
class NotificationScreen extends StatelessWidget {
  const NotificationScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Notifikasi'),
        actions: [
          Consumer<NotificationProvider>(
            builder: (context, provider, _) {
              if (provider.unreadCount > 0) {
                return TextButton(
                  onPressed: provider.markAllAsRead,
                  child: const Text('Baca Semua'),
                );
              }
              return const SizedBox.shrink();
            },
          ),
          Consumer<NotificationProvider>(
            builder: (context, provider, _) {
              if (provider.notifications.isNotEmpty) {
                return IconButton(
                  icon: const Icon(Icons.delete_sweep),
                  onPressed: () => _confirmClearAll(context, provider),
                  tooltip: 'Hapus Semua',
                );
              }
              return const SizedBox.shrink();
            },
          ),
        ],
      ),
      body: Consumer<NotificationProvider>(
        builder: (context, provider, _) {
          if (provider.notifications.isEmpty) {
            return const EmptyState(
              icon: Icons.notifications_none,
              title: 'Belum ada notifikasi',
              subtitle: 'Notifikasi akan muncul di sini saat Anda melakukan transaksi atau block baru ditambang',
            );
          }

          return ListView.separated(
            itemCount: provider.notifications.length,
            separatorBuilder: (_, __) => const Divider(height: 1),
            itemBuilder: (context, index) {
              final notification = provider.notifications[index];
              return _NotificationTile(
                notification: notification,
                onTap: () {
                  provider.markAsRead(notification.id);
                  if (notification.route != '/') {
                    context.go(notification.route);
                  }
                },
                onDismiss: () => provider.remove(notification.id),
              );
            },
          );
        },
      ),
    );
  }

  void _confirmClearAll(BuildContext context, NotificationProvider provider) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Hapus Semua Notifikasi'),
        content: const Text('Apakah Anda yakin ingin menghapus semua notifikasi?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Batal'),
          ),
          ElevatedButton(
            onPressed: () {
              provider.clearAll();
              Navigator.pop(ctx);
            },
            style: ElevatedButton.styleFrom(backgroundColor: AppTheme.error),
            child: const Text('Hapus Semua'),
          ),
        ],
      ),
    );
  }
}

class _NotificationTile extends StatelessWidget {
  final AppNotification notification;
  final VoidCallback onTap;
  final VoidCallback onDismiss;

  const _NotificationTile({
    required this.notification,
    required this.onTap,
    required this.onDismiss,
  });

  @override
  Widget build(BuildContext context) {
    final icon = _iconForType(notification.type);
    final color = _colorForPriority(notification.priority);
    final timeStr = _formatTimestamp(notification.timestamp);

    return Dismissible(
      key: Key(notification.id),
      direction: DismissDirection.endToStart,
      onDismissed: (_) => onDismiss(),
      background: Container(
        alignment: Alignment.centerRight,
        padding: const EdgeInsets.only(right: 16),
        color: AppTheme.error,
        child: const Icon(Icons.delete, color: Colors.white),
      ),
      child: ListTile(
        onTap: onTap,
        leading: Container(
          width: 40,
          height: 40,
          decoration: BoxDecoration(
            color: color.withValues(alpha: 0.15),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(icon, color: color, size: 20),
        ),
        title: Row(
          children: [
            Expanded(
              child: Text(
                notification.title,
                style: TextStyle(
                  fontWeight: notification.isRead ? FontWeight.normal : FontWeight.bold,
                  fontSize: 14,
                ),
              ),
            ),
            if (!notification.isRead)
              Container(
                width: 8,
                height: 8,
                decoration: const BoxDecoration(
                  color: AppTheme.primary,
                  shape: BoxShape.circle,
                ),
              ),
          ],
        ),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              notification.message,
              style: TextStyle(
                fontSize: 12,
                color: notification.isRead
                    ? AppTheme.darkTextSecondary
                    : AppTheme.darkText,
              ),
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
            const SizedBox(height: 4),
            Text(
              timeStr,
              style: const TextStyle(
                fontSize: 11,
                color: AppTheme.darkTextSecondary,
              ),
            ),
          ],
        ),
        isThreeLine: true,
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

  Color _colorForPriority(String priority) {
    return switch (priority) {
      'high' => AppTheme.error,
      'medium' => AppTheme.warning,
      _ => AppTheme.primary,
    };
  }

  String _formatTimestamp(int timestamp) {
    final dt = DateTime.fromMillisecondsSinceEpoch(timestamp * 1000);
    final now = DateTime.now();
    final diff = now.difference(dt);

    if (diff.inMinutes < 1) return 'Baru saja';
    if (diff.inMinutes < 60) return '${diff.inMinutes} menit lalu';
    if (diff.inHours < 24) return '${diff.inHours} jam lalu';
    if (diff.inDays < 7) return '${diff.inDays} hari lalu';
    return DateFormat('dd MMM yyyy').format(dt);
  }
}
