import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:go_router/go_router.dart';
import 'package:uuid/uuid.dart';
import '../../core/api/api_client.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/utils/formatters.dart';
import '../../shared/widgets/app_widgets.dart';
import 'transaction_helper.dart';

/// Screen untuk membeli YTE dari sistem (BUY).
///
/// Flow sama dengan SEND tapi:
/// - Hanya butuh address (buyer) dan amount
/// - Pesan canonical: " BUY {amount} nonce:{nonce}"
/// - User membayar dalam USD, menerima YTE
class BuyScreen extends StatefulWidget {
  const BuyScreen({super.key});

  @override
  State<BuyScreen> createState() => _BuyScreenState();
}

class _BuyScreenState extends State<BuyScreen> {
  final _formKey = GlobalKey<FormState>();
  final _addressController = TextEditingController();
  final _amountController = TextEditingController();
  final _nonceController = TextEditingController();
  final _signatureController = TextEditingController();
  final _api = ApiClient();
  late final TransactionHelper _txHelper;

  bool _isLoading = false;
  bool _isSigningMode = false;
  String? _canonicalMessage;
  String? _error;
  String? _success;

  // Fee estimation
  String _priority = 'low';
  Map<String, dynamic>? _feeEstimate;

  @override
  void initState() {
    super.initState();
    _txHelper = TransactionHelper(_api);
  }

  @override
  void dispose() {
    _addressController.dispose();
    _amountController.dispose();
    _nonceController.dispose();
    _signatureController.dispose();
    super.dispose();
  }

  Future<void> _generateNonce() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final nonce =
          await _txHelper.generateNonce(_addressController.text.trim());
      _nonceController.text = nonce;

      final amount = double.parse(_amountController.text.trim());
      final msg = TransactionHelper.buildBuyMessage(
        amount: amount,
        nonce: nonce,
      );

      setState(() {
        _canonicalMessage = msg;
        _isSigningMode = true;
      });
    } on ApiException catch (e) {
      setState(() => _error = e.message);
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() {
      _isLoading = true;
      _error = null;
      _success = null;
    });

    try {
      final idempotencyKey = const Uuid().v4();
      final result = await _txHelper.buyTransaction(
        address: _addressController.text.trim(),
        amount: double.parse(_amountController.text.trim()),
        nonce: _nonceController.text,
        signature: _signatureController.text.trim(),
        idempotencyKey: idempotencyKey,
      );

      setState(() {
        _success = result['message'] as String? ??
            'Buy order berhasil dikirim dan sedang diproses';
        _isSigningMode = false;
      });
    } on ApiException catch (e) {
      setState(() => _error = e.message);
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  void _reset() {
    setState(() {
      _isSigningMode = false;
      _canonicalMessage = null;
      _error = null;
      _success = null;
      _nonceController.clear();
      _signatureController.clear();
      _feeEstimate = null;
    });
  }

  Future<void> _fetchFee() async {
    final amountText = _amountController.text.trim();
    final amount = double.tryParse(amountText);
    if (amount == null || amount <= 0) {
      setState(() => _feeEstimate = null);
      return;
    }

    try {
      final resp = await _api.estimateFee(amount: amount, priority: _priority);
      setState(() {
        _feeEstimate = resp.data['data'] as Map<String, dynamic>;
      });
    } on ApiException catch (_) {
      setState(() => _feeEstimate = null);
    }
  }

  Widget _buildFeeCard() {
    if (_feeEstimate == null) return const SizedBox.shrink();

    final baseFee = _feeEstimate!['base_fee'] as double;
    final estimatedFee = _feeEstimate!['estimated_fee'] as double;
    final congestionLevel = _feeEstimate!['congestion_level'] as String;
    final pendingCount = _feeEstimate!['pending_count'] as int;
    final congestionMult = _feeEstimate!['congestion_multiplier'] as double;

    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: AppTheme.darkSurface,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: AppTheme.darkBorder),
      ),
      child: Column(
        children: [
          _feeRow('Base Fee', formatYTE(baseFee)),
          _feeRow('Congestion', '$congestionMult x ($pendingCount pending)'),
          _feeRow('Prioritas', _priority),
          const Divider(height: 16),
          _feeRow('Estimasi Fee', formatYTE(estimatedFee), isBold: true),
          const SizedBox(height: 8),
          Text(
            'Jaringan: $congestionLevel',
            style: TextStyle(
              color: congestionLevel == 'low'
                  ? AppTheme.success
                  : AppTheme.warning,
              fontSize: 11,
            ),
          ),
        ],
      ),
    );
  }

  Widget _feeRow(String label, String value, {bool isBold = false}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label,
              style: TextStyle(
                  color: AppTheme.darkTextSecondary,
                  fontSize: 12,
                  fontWeight: isBold ? FontWeight.bold : FontWeight.normal)),
          Text(value,
              style: TextStyle(
                  fontSize: 12,
                  fontWeight: isBold ? FontWeight.bold : FontWeight.w600)),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Beli YTE')),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          AppCard(
            child: Form(
              key: _formKey,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  Text(
                    _isSigningMode
                        ? 'Langkah 2: Tandatangani pesan'
                        : 'Langkah 1: Isi detail pembelian',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 20),

                  // Address
                  TextFormField(
                    controller: _addressController,
                    readOnly: _isSigningMode,
                    decoration: const InputDecoration(
                      labelText: 'Wallet Address',
                      hintText: '0x...',
                      prefixIcon: Icon(Icons.account_balance_wallet),
                    ),
                    validator: (v) {
                      if (v == null || v.trim().isEmpty) {
                        return 'Address wajib diisi';
                      }
                      if (!v.trim().startsWith('0x') || v.trim().length != 42) {
                        return 'Format address tidak valid';
                      }
                      return null;
                    },
                  ),
                  const SizedBox(height: 16),

                  // Amount
                  TextFormField(
                    controller: _amountController,
                    readOnly: _isSigningMode,
                    keyboardType:
                        const TextInputType.numberWithOptions(decimal: true),
                    decoration: const InputDecoration(
                      labelText: 'Jumlah YTE yang ingin dibeli',
                      hintText: '5.0',
                      prefixIcon: Icon(Icons.token),
                    ),
                    onChanged: (_) {
                      if (!_isSigningMode) _fetchFee();
                    },
                    validator: (v) {
                      if (v == null || v.trim().isEmpty) {
                        return 'Jumlah wajib diisi';
                      }
                      final amount = double.tryParse(v.trim());
                      if (amount == null || amount <= 0) {
                        return 'Jumlah harus lebih besar dari 0';
                      }
                      return null;
                    },
                  ),
                  const SizedBox(height: 20),

                  // Fee estimation
                  if (!_isSigningMode && _feeEstimate != null) ...[
                    _buildFeeCard(),
                    const SizedBox(height: 20),
                  ],

                  if (!_isSigningMode) ...[
                    ElevatedButton.icon(
                      onPressed: _isLoading ? null : _generateNonce,
                      icon: _isLoading
                          ? const SizedBox(
                              height: 18,
                              width: 18,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            )
                          : const Icon(Icons.vpn_key),
                      label: const Text('Generate Nonce & Lanjut'),
                    ),
                  ] else ...[
                    // Canonical message
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: AppTheme.darkSurface,
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: AppTheme.darkBorder),
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            children: [
                              const Icon(Icons.info_outline,
                                  size: 16, color: AppTheme.primary),
                              const SizedBox(width: 6),
                              Text(
                                'Pesan yang ditandatangani:',
                                style: Theme.of(context)
                                    .textTheme
                                    .bodySmall
                                    ?.copyWith(
                                      color: AppTheme.darkTextSecondary,
                                    ),
                              ),
                              const Spacer(),
                              IconButton(
                                icon: const Icon(Icons.copy, size: 16),
                                onPressed: () {
                                  Clipboard.setData(ClipboardData(
                                      text: _canonicalMessage!));
                                  ScaffoldMessenger.of(context).showSnackBar(
                                    const SnackBar(
                                        content: Text('Pesan disalin')),
                                  );
                                },
                              ),
                            ],
                          ),
                          const SizedBox(height: 8),
                          SelectableText(
                            _canonicalMessage!,
                            style: const TextStyle(
                              fontFamily: 'monospace',
                              fontSize: 13,
                            ),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 8),

                    // Nonce
                    TextFormField(
                      controller: _nonceController,
                      readOnly: true,
                      decoration: const InputDecoration(
                        labelText: 'Nonce',
                        prefixIcon: Icon(Icons.key),
                      ),
                    ),
                    const SizedBox(height: 16),

                    // Signature
                    TextFormField(
                      controller: _signatureController,
                      decoration: const InputDecoration(
                        labelText: 'Signature',
                        hintText: '0x...',
                        prefixIcon: Icon(Icons.edit),
                      ),
                      maxLines: 3,
                      validator: (v) {
                        if (v == null || v.trim().isEmpty) {
                          return 'Signature wajib diisi';
                        }
                        if (v.trim().length != 132) {
                          return 'Signature harus 132 karakter';
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 24),

                    ElevatedButton.icon(
                      onPressed: _isLoading ? null : _submit,
                      icon: _isLoading
                          ? const SizedBox(
                              height: 18,
                              width: 18,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            )
                          : const Icon(Icons.shopping_cart),
                      label: const Text('Beli YTE'),
                    ),
                    const SizedBox(height: 8),
                    TextButton(
                      onPressed: _isLoading ? null : _reset,
                      child: const Text('Batal & Ubah Detail'),
                    ),
                  ],

                  if (_error != null) ...[
                    const SizedBox(height: 16),
                    _ErrorBanner(message: _error!),
                  ],

                  if (_success != null) ...[
                    const SizedBox(height: 16),
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: AppTheme.success.withValues(alpha: 0.1),
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(
                            color: AppTheme.success.withValues(alpha: 0.3)),
                      ),
                      child: Row(
                        children: [
                          const Icon(Icons.check_circle,
                              color: AppTheme.success, size: 18),
                          const SizedBox(width: 8),
                          Expanded(
                            child: Text(_success!,
                                style:
                                    const TextStyle(color: AppTheme.success)),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 12),
                    ElevatedButton(
                      onPressed: () => context.go('/'),
                      child: const Text('Kembali ke Dashboard'),
                    ),
                  ],
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _ErrorBanner extends StatelessWidget {
  final String message;
  const _ErrorBanner({required this.message});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: AppTheme.error.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: AppTheme.error.withValues(alpha: 0.3)),
      ),
      child: Row(
        children: [
          const Icon(Icons.error_outline, color: AppTheme.error, size: 18),
          const SizedBox(width: 8),
          Expanded(
            child: Text(message,
                style: const TextStyle(color: AppTheme.error, fontSize: 13)),
          ),
        ],
      ),
    );
  }
}
