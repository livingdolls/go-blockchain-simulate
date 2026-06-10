import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:go_router/go_router.dart';
import 'package:uuid/uuid.dart';
import '../../core/api/api_client.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/utils/formatters.dart';
import '../../shared/widgets/app_widgets.dart';
import '../../shared/widgets/fee_estimation_card.dart';
import '../qr/qr_data.dart';
import '../qr/qr_scanner_screen.dart';
import 'transaction_helper.dart';

/// Screen untuk mengirim transaksi transfer (SEND).
///
/// Flow:
/// 1. User isi from_address, to_address, amount
/// 2. Generate nonce dari backend
/// 3. Tampilkan pesan canonical yang perlu ditandatangani
/// 4. User paste signature (dari wallet)
/// 5. Submit ke backend
class SendScreen extends StatefulWidget {
  const SendScreen({super.key});

  @override
  State<SendScreen> createState() => _SendScreenState();
}

class _SendScreenState extends State<SendScreen> {
  final _formKey = GlobalKey<FormState>();
  final _fromController = TextEditingController();
  final _toController = TextEditingController();
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
  bool _isFeeLoading = false;

  @override
  void initState() {
    super.initState();
    _txHelper = TransactionHelper(_api);
  }

  @override
  void dispose() {
    _fromController.dispose();
    _toController.dispose();
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
          await _txHelper.generateNonce(_fromController.text.trim());
      _nonceController.text = nonce;

      final amount = double.parse(_amountController.text.trim());
      final msg = TransactionHelper.buildSendMessage(
        amount: amount,
        toAddress: _toController.text.trim(),
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
      final result = await _txHelper.sendTransaction(
        fromAddress: _fromController.text.trim(),
        toAddress: _toController.text.trim(),
        amount: double.parse(_amountController.text.trim()),
        nonce: _nonceController.text,
        signature: _signatureController.text.trim(),
        idempotencyKey: idempotencyKey,
      );

      setState(() {
        _success = result['message'] as String? ??
            'Transaksi berhasil dikirim dan sedang diproses';
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

    setState(() => _isFeeLoading = true);

    try {
      final resp = await _api.estimateFee(amount: amount, priority: _priority);
      setState(() {
        _feeEstimate = resp.data['data'] as Map<String, dynamic>;
        _isFeeLoading = false;
      });
    } on ApiException catch (_) {
      setState(() {
        _feeEstimate = null;
        _isFeeLoading = false;
      });
    }
  }

  Future<void> _scanQr() async {
    final result = await Navigator.push<QrData>(
      context,
      MaterialPageRoute(builder: (_) => const QrScannerScreen()),
    );

    if (result != null && mounted) {
      setState(() {
        _toController.text = result.address;
        if (result.amount != null) {
          _amountController.text = result.amount!.toString();
        }
      });

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              result.amount != null
                  ? 'QR di-scan: ${shortAddress(result.address)} (${result.amount} YTE)'
                  : 'QR di-scan: ${shortAddress(result.address)}',
            ),
            duration: const Duration(seconds: 2),
          ),
        );
      }
    }
  }

  Widget _buildFeeEstimation() {
    if (_isFeeLoading) {
      return const Padding(
        padding: EdgeInsets.symmetric(vertical: 8),
        child: Row(
          children: [
            SizedBox(
              width: 16,
              height: 16,
              child: CircularProgressIndicator(strokeWidth: 2),
            ),
            SizedBox(width: 8),
            Text('Menghitung fee...', style: TextStyle(fontSize: 12)),
          ],
        ),
      );
    }

    if (_feeEstimate == null) return const SizedBox.shrink();

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        // Priority selector
        Row(
          children: [
            const Text('Prioritas: ',
                style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
            ...['low', 'medium', 'high'].map((p) {
              final label = switch (p) {
                'low' => 'Rendah',
                'medium' => 'Sedang',
                'high' => 'Tinggi',
                _ => p,
              };
              final isSelected = _priority == p;
              return Padding(
                padding: const EdgeInsets.only(left: 4),
                child: ChoiceChip(
                  label: Text(label, style: const TextStyle(fontSize: 12)),
                  selected: isSelected,
                  onSelected: (_) {
                    setState(() => _priority = p);
                    _fetchFee();
                  },
                  selectedColor: AppTheme.primary.withValues(alpha: 0.3),
                ),
              );
            }),
          ],
        ),
        const SizedBox(height: 12),
        FeeEstimationCard(feeData: _feeEstimate!, priority: _priority),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Kirim YTE')),
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
                        : 'Langkah 1: Isi detail transaksi',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 20),

                  // From address
                  TextFormField(
                    controller: _fromController,
                    readOnly: _isSigningMode,
                    decoration: const InputDecoration(
                      labelText: 'Dari Address',
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

                  // To address
                  TextFormField(
                    controller: _toController,
                    readOnly: _isSigningMode,
                    decoration: InputDecoration(
                      labelText: 'Ke Address',
                      hintText: '0x...',
                      prefixIcon: const Icon(Icons.send),
                      suffixIcon: _isSigningMode || kIsWeb
                          ? null
                          : IconButton(
                              icon: const Icon(Icons.qr_code_scanner),
                              onPressed: _scanQr,
                              tooltip: 'Scan QR code',
                            ),
                    ),
                    validator: (v) {
                      if (v == null || v.trim().isEmpty) {
                        return 'Address tujuan wajib diisi';
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
                      labelText: 'Jumlah (YTE)',
                      hintText: '10.5',
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

                  // Fee estimation (hanya tampil di step 1)
                  if (!_isSigningMode) ...[
                    _buildFeeEstimation(),
                    const SizedBox(height: 20),
                  ],

                  if (!_isSigningMode) ...[
                    ElevatedButton.icon(
                      onPressed: _isLoading ? null : _generateNonce,
                      icon: _isLoading
                          ? const SizedBox(
                              height: 18,
                              width: 18,
                              child:
                                  CircularProgressIndicator(strokeWidth: 2),
                            )
                          : const Icon(Icons.vpn_key),
                      label: const Text('Generate Nonce & Lanjut'),
                    ),
                  ] else ...[
                    // Canonical message (read-only, copyable)
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

                    // Nonce (read-only)
                    TextFormField(
                      controller: _nonceController,
                      readOnly: true,
                      decoration: const InputDecoration(
                        labelText: 'Nonce',
                        prefixIcon: Icon(Icons.key),
                      ),
                    ),
                    const SizedBox(height: 16),

                    // Signature input
                    TextFormField(
                      controller: _signatureController,
                      decoration: const InputDecoration(
                        labelText: 'Signature',
                        hintText: '0x... (tempel hasil sign dari wallet)',
                        prefixIcon: Icon(Icons.edit),
                      ),
                      maxLines: 3,
                      validator: (v) {
                        if (v == null || v.trim().isEmpty) {
                          return 'Signature wajib diisi';
                        }
                        if (v.trim().length != 132) {
                          return 'Signature harus 132 karakter (0x + 130 hex)';
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 24),

                    // Submit button
                    ElevatedButton.icon(
                      onPressed: _isLoading ? null : _submit,
                      icon: _isLoading
                          ? const SizedBox(
                              height: 18,
                              width: 18,
                              child:
                                  CircularProgressIndicator(strokeWidth: 2),
                            )
                          : const Icon(Icons.send),
                      label: const Text('Kirim Transaksi'),
                    ),
                    const SizedBox(height: 8),
                    TextButton(
                      onPressed: _isLoading ? null : _reset,
                      child: const Text('Batal & Ubah Detail'),
                    ),
                  ],

                  // Error
                  if (_error != null) ...[
                    const SizedBox(height: 16),
                    ErrorBanner(message: _error!),
                  ],

                  // Success
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
