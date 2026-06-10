import '../../core/api/api_client.dart';

/// Helper untuk flow transaksi blockchain:
/// 1. Generate nonce dari backend
/// 2. Bangun pesan canonical sesuai tipe transaksi
/// 3. (Frontend) User sign pesan dengan wallet
/// 4. Submit transaksi ke backend
class TransactionHelper {
  final ApiClient api;

  TransactionHelper(this.api);

  /// Generate nonce untuk address tertentu.
  Future<String> generateNonce(String address) async {
    final resp = await api.generateNonce(address);
    return resp.data['data']['nonce'] as String;
  }

  /// Bangun pesan canonical untuk SEND transaction.
  /// Format: "Send {amount} to {toAddress} nonce:{nonce}"
  static String buildSendMessage({
    required double amount,
    required String toAddress,
    required String nonce,
  }) {
    return 'Send ${amount.toStringAsFixed(2)} to $toAddress nonce:$nonce';
  }

  /// Bangun pesan canonical untuk BUY transaction.
  /// Format: " BUY {amount} nonce:{nonce}" (perhatikan spasi di awal)
  static String buildBuyMessage({
    required double amount,
    required String nonce,
  }) {
    return ' BUY ${amount.toStringAsFixed(2)} nonce:$nonce';
  }

  /// Bangun pesan canonical untuk SELL transaction.
  /// Format: " SELL {amount} nonce:{nonce}" (perhatikan spasi di awal)
  static String buildSellMessage({
    required double amount,
    required String nonce,
  }) {
    return ' SELL ${amount.toStringAsFixed(2)} nonce:$nonce';
  }

  /// Submit SEND transaction.
  Future<Map<String, dynamic>> sendTransaction({
    required String fromAddress,
    required String toAddress,
    required double amount,
    required String nonce,
    required String signature,
    String? idempotencyKey,
  }) async {
    final resp = await api.sendTransaction(
      fromAddress: fromAddress,
      toAddress: toAddress,
      amount: amount,
      nonce: nonce,
      signature: signature,
      idempotencyKey: idempotencyKey,
    );
    return resp.data['data'] as Map<String, dynamic>;
  }

  /// Submit BUY transaction.
  Future<Map<String, dynamic>> buyTransaction({
    required String address,
    required double amount,
    required String nonce,
    required String signature,
    String? idempotencyKey,
  }) async {
    final resp = await api.buyTransaction(
      address: address,
      amount: amount,
      nonce: nonce,
      signature: signature,
      idempotencyKey: idempotencyKey,
    );
    return resp.data['data'] as Map<String, dynamic>;
  }

  /// Submit SELL transaction.
  Future<Map<String, dynamic>> sellTransaction({
    required String address,
    required double amount,
    required String nonce,
    required String signature,
    String? idempotencyKey,
  }) async {
    final resp = await api.sellTransaction(
      address: address,
      amount: amount,
      nonce: nonce,
      signature: signature,
      idempotencyKey: idempotencyKey,
    );
    return resp.data['data'] as Map<String, dynamic>;
  }
}
