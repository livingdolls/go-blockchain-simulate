/// Model untuk notification event dari backend.
///
/// Backend mengirim via WebSocket dengan format:
/// ```json
/// {
///   "type": "notification.TRANSACTION_CONFIRMED",
///   "data": {
///     "id": "uuid",
///     "type": "TRANSACTION_CONFIRMED",
///     "priority": "high",
///     "title": "Transaksi Terkonfirmasi",
///     "message": "...",
///     "data": {...},
///     "timestamp": 1700000000,
///     "related_tx_id": 42,
///     "related_block_id": 100
///   }
/// }
/// ```
class AppNotification {
  final String id;
  final String type;
  final String priority;
  final String title;
  final String message;
  final Map<String, dynamic>? data;
  final int? relatedTxId;
  final int? relatedBlockId;
  final int timestamp;
  bool isRead;

  AppNotification({
    required this.id,
    required this.type,
    required this.priority,
    required this.title,
    required this.message,
    this.data,
    this.relatedTxId,
    this.relatedBlockId,
    required this.timestamp,
    this.isRead = false,
  });

  factory AppNotification.fromJson(Map<String, dynamic> json) {
    return AppNotification(
      id: json['id'] as String? ?? '',
      type: json['type'] as String? ?? '',
      priority: json['priority'] as String? ?? 'low',
      title: json['title'] as String? ?? '',
      message: json['message'] as String? ?? '',
      data: json['data'] as Map<String, dynamic>?,
      relatedTxId: json['related_tx_id'] as int?,
      relatedBlockId: json['related_block_id'] as int?,
      timestamp: json['timestamp'] as int? ?? 0,
      isRead: json['is_read'] as bool? ?? false,
    );
  }

  /// Parse dari WebSocket message (type + data wrapper).
  factory AppNotification.fromWsMessage(Map<String, dynamic> msg) {
    final data = msg['data'] as Map<String, dynamic>? ?? {};
    return AppNotification.fromJson(data);
  }

  bool get isHighPriority => priority == 'high';
  bool get isMediumPriority => priority == 'medium';
  bool get isLowPriority => priority == 'low';

  String get typeLabel {
    return switch (type) {
      'TRANSACTION_CONFIRMED' => 'Transaksi Terkonfirmasi',
      'TRANSACTION_SUBMITTED' => 'Transaksi Terkirim',
      'BLOCK_CONFIRMED' => 'Block Baru',
      'REWARD_EARNED' => 'Reward Diterima',
      'BALANCE_UPDATED' => 'Saldo Diperbarui',
      'TRANSACTION_BLOCK_MINED' => 'Block Ditambang',
      _ => type,
    };
  }

  String get route {
    if (relatedTxId != null) return '/transactions/$relatedTxId';
    if (relatedBlockId != null) return '/blocks/$relatedBlockId';
    return '/';
  }
}
