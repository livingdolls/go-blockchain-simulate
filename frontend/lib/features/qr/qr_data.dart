/// Encode/decode untuk format QR code YTEPay.
///
/// Format: `ytepay:<address>[?amount=<amount>]`
///
/// Contoh:
/// - `ytepay:0xABC123...5678` (address only)
/// - `ytepay:0xABC123...5678?amount=10.5` (address + amount)
class QrData {
  static const String scheme = 'ytepay';

  final String address;
  final double? amount;

  const QrData({required this.address, this.amount});

  /// Encode ke string format `ytepay:...`
  String encode() {
    final buf = StringBuffer('$scheme:$address');
    if (amount != null) {
      buf.write('?amount=$amount');
    }
    return buf.toString();
  }

  /// Parse string dari QR scan. Return null jika format tidak valid.
  ///
  /// Validasi:
  /// - Harus dimulai dengan `ytepay:`
  /// - Address harus `0x` + 40 hex chars
  /// - Amount (jika ada) harus angka positif
  static QrData? parse(String raw) {
    final trimmed = raw.trim();

    if (!trimmed.startsWith('$scheme:')) {
      return null;
    }

    final withoutScheme = trimmed.substring(scheme.length + 1);
    if (withoutScheme.isEmpty) return null;

    String address;
    double? amount;

    final queryStart = withoutScheme.indexOf('?');
    if (queryStart >= 0) {
      address = withoutScheme.substring(0, queryStart);
      final query = withoutScheme.substring(queryStart + 1);
      amount = _parseAmountFromQuery(query);
    } else {
      address = withoutScheme;
    }

    // Validasi address format
    if (!_isValidAddress(address)) {
      return null;
    }

    return QrData(address: address, amount: amount);
  }

  /// Validasi Ethereum address: 0x + 40 hex chars.
  static bool _isValidAddress(String addr) {
    if (!addr.startsWith('0x') || addr.length != 42) return false;
    final hex = addr.substring(2);
    return RegExp(r'^[0-9a-fA-F]{40}$').hasMatch(hex);
  }

  /// Parse amount dari query string (?amount=10.5)
  static double? _parseAmountFromQuery(String query) {
    for (final part in query.split('&')) {
      final kv = part.split('=');
      if (kv.length == 2 && kv[0] == 'amount') {
        final value = double.tryParse(kv[1]);
        if (value != null && value > 0) return value;
      }
    }
    return null;
  }

  @override
  String toString() => encode();

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is QrData &&
          runtimeType == other.runtimeType &&
          address == other.address &&
          amount == other.amount;

  @override
  int get hashCode => address.hashCode ^ amount.hashCode;
}
