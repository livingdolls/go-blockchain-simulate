import 'package:intl/intl.dart';

/// Format angka dengan pemisah ribuan.
/// Contoh: 1234567.89 → "1,234,567.89"
String formatNumber(num value, {int decimals = 2}) {
  final formatter = NumberFormat('#,##0.${'0' * decimals}');
  return formatter.format(value);
}

/// Format USD.
/// Contoh: 1234.56 → "$1,234.56"
String formatUSD(double value) {
  return '\$${formatNumber(value)}';
}

/// Format YTE.
/// Contoh: 1234.56789012 → "1,234.56789012"
String formatYTE(double value) {
  return '${formatNumber(value, decimals: 8)} YTE';
}

/// Format persentase.
/// Contoh: 2.5 → "+2.50%" atau -1.2 → "-1.20%"
String formatPercent(double value) {
  final sign = value >= 0 ? '+' : '';
  return '$sign${value.toStringAsFixed(2)}%';
}

/// Format tanggal dari Unix timestamp.
String formatDate(int timestamp) {
  final dt = DateTime.fromMillisecondsSinceEpoch(timestamp * 1000);
  return DateFormat('dd MMM yyyy HH:mm').format(dt);
}

/// Format tanggal dari ISO string.
String formatISODate(String isoString) {
  try {
    final dt = DateTime.parse(isoString);
    return DateFormat('dd MMM yyyy HH:mm').format(dt);
  } catch (_) {
    return isoString;
  }
}

/// Format relative time.
/// Contoh: "2 menit lalu", "1 jam lalu"
String formatRelativeTime(int timestamp) {
  final dt = DateTime.fromMillisecondsSinceEpoch(timestamp * 1000);
  final now = DateTime.now();
  final diff = now.difference(dt);

  if (diff.inSeconds < 60) return '${diff.inSeconds} detik lalu';
  if (diff.inMinutes < 60) return '${diff.inMinutes} menit lalu';
  if (diff.inHours < 24) return '${diff.inHours} jam lalu';
  if (diff.inDays < 30) return '${diff.inDays} hari lalu';
  return DateFormat('dd MMM yyyy').format(dt);
}

/// Potong address Ethereum.
/// Contoh: "0x1234567890abcdef..." → "0x1234...cdef"
String shortAddress(String address, {int start = 6, int end = 4}) {
  if (address.length <= start + end + 3) return address;
  return '${address.substring(0, start)}...${address.substring(address.length - end)}';
}

/// Potong hash.
String shortHash(String hash, {int chars = 8}) {
  if (hash.length <= chars * 2) return hash;
  return '${hash.substring(0, chars)}...${hash.substring(hash.length - chars)}';
}
