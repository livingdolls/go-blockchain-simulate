import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../../core/models/notification.dart';

/// State management untuk notifikasi.
///
/// Menyimpan daftar notifikasi di memory + persist ke SharedPreferences
/// agar user bisa lihat history setelah refresh.
///
/// Fitur:
/// - addNotification(): tambah notifikasi baru (prepend)
/// - markAsRead(id): tandai satu notifikasi sudah dibaca
/// - markAllAsRead(): tandai semua sudah dibaca
/// - clearAll(): hapus semua notifikasi
/// - unreadCount: jumlah notifikasi belum dibaca
/// - persist ke SharedPreferences (JSON)
class NotificationProvider extends ChangeNotifier {
  List<AppNotification> _notifications = [];
  static const String _storageKey = 'notifications_v1';

  List<AppNotification> get notifications => List.unmodifiable(_notifications);

  int get unreadCount => _notifications.where((n) => !n.isRead).length;

  List<AppNotification> get unreadNotifications =>
      _notifications.where((n) => !n.isRead).toList();

  /// Tambah notifikasi baru di awal list (newest first).
  /// Dipanggil saat WebSocket menerima notification.* event.
  void addNotification(AppNotification notification) {
    // Hindari duplikat berdasarkan ID
    if (_notifications.any((n) => n.id == notification.id)) return;

    _notifications.insert(0, notification);

    // Batasi jumlah notifikasi di memory (mencegah memory leak)
    if (_notifications.length > 200) {
      _notifications = _notifications.sublist(0, 200);
    }

    notifyListeners();
    _persist();
  }

  /// Tandai satu notifikasi sudah dibaca.
  void markAsRead(String id) {
    final idx = _notifications.indexWhere((n) => n.id == id);
    if (idx >= 0 && !_notifications[idx].isRead) {
      _notifications[idx].isRead = true;
      notifyListeners();
      _persist();
    }
  }

  /// Tandai semua notifikasi sudah dibaca.
  void markAllAsRead() {
    bool changed = false;
    for (final n in _notifications) {
      if (!n.isRead) {
        n.isRead = true;
        changed = true;
      }
    }
    if (changed) {
      notifyListeners();
      _persist();
    }
  }

  /// Hapus semua notifikasi.
  void clearAll() {
    if (_notifications.isNotEmpty) {
      _notifications.clear();
      notifyListeners();
      _persist();
    }
  }

  /// Hapus satu notifikasi.
  void remove(String id) {
    _notifications.removeWhere((n) => n.id == id);
    notifyListeners();
    _persist();
  }

  /// Load notifikasi dari SharedPreferences.
  Future<void> loadFromStorage() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final json = prefs.getString(_storageKey);
      if (json != null && json.isNotEmpty) {
        final list = jsonDecode(json) as List<dynamic>;
        _notifications = list
            .map((e) => AppNotification.fromJson(e as Map<String, dynamic>))
            .toList();
        notifyListeners();
      }
    } catch (_) {
      // Ignore storage errors
    }
  }

  /// Persist ke SharedPreferences.
  Future<void> _persist() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final json = jsonEncode(_notifications.map((n) => {
            'id': n.id,
            'type': n.type,
            'priority': n.priority,
            'title': n.title,
            'message': n.message,
            'data': n.data,
            'related_tx_id': n.relatedTxId,
            'related_block_id': n.relatedBlockId,
            'timestamp': n.timestamp,
            'is_read': n.isRead,
          }).toList());
      await prefs.setString(_storageKey, json);
    } catch (_) {
      // Ignore storage errors
    }
  }
}
