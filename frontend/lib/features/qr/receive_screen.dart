import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/utils/formatters.dart';
import '../../shared/widgets/app_widgets.dart';
import '../../shared/widgets/qr_display_widget.dart';
import 'qr_data.dart';

/// Screen untuk menerima YTE. Menampilkan QR code berisi wallet address
/// dan optional amount yang bisa di-scan oleh pengirim.
///
/// Flow:
/// 1. Tampilkan QR code default (address only)
/// 2. User bisa input amount → QR update ke format `ytepay:address?amount=X`
/// 3. Pengirim scan QR → auto-fill form kirim
class ReceiveScreen extends StatefulWidget {
  final String address;
  final String? username;

  const ReceiveScreen({
    super.key,
    required this.address,
    this.username,
  });

  @override
  State<ReceiveScreen> createState() => _ReceiveScreenState();
}

class _ReceiveScreenState extends State<ReceiveScreen> {
  final _amountController = TextEditingController();
  double? _requestedAmount;

  @override
  void dispose() {
    _amountController.dispose();
    super.dispose();
  }

  String get _qrData {
    return QrData(
      address: widget.address,
      amount: _requestedAmount,
    ).encode();
  }

  void _onAmountChanged(String value) {
    final amount = double.tryParse(value.trim());
    setState(() {
      _requestedAmount = (amount != null && amount > 0) ? amount : null;
    });
  }

  void _copyAddress() {
    Clipboard.setData(ClipboardData(text: widget.address));
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Address disalin ke clipboard')),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Terima YTE')),
      body: ListView(
        padding: const EdgeInsets.all(24),
        children: [
          // QR Code
          Center(child: QrDisplayWidget(data: _qrData)),
          const SizedBox(height: 24),

          // Address display
          AppCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Wallet Address',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: AppTheme.darkTextSecondary,
                      ),
                ),
                const SizedBox(height: 8),
                Row(
                  children: [
                    Expanded(
                      child: SelectableText(
                        widget.address,
                        style: const TextStyle(
                          fontFamily: 'monospace',
                          fontSize: 13,
                        ),
                      ),
                    ),
                    IconButton(
                      icon: const Icon(Icons.copy, size: 18),
                      onPressed: _copyAddress,
                      tooltip: 'Salin address',
                    ),
                  ],
                ),
                if (widget.username != null) ...[
                  const SizedBox(height: 4),
                  Text(
                    widget.username!,
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                          color: AppTheme.darkTextSecondary,
                        ),
                  ),
                ],
              ],
            ),
          ),
          const SizedBox(height: 16),

          // Optional amount input
          AppCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Jumlah yang Diminta (opsional)',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: AppTheme.darkTextSecondary,
                      ),
                ),
                const SizedBox(height: 8),
                TextFormField(
                  controller: _amountController,
                  keyboardType:
                      const TextInputType.numberWithOptions(decimal: true),
                  decoration: const InputDecoration(
                    hintText: 'Masukkan jumlah YTE',
                    prefixIcon: Icon(Icons.token),
                    suffixText: 'YTE',
                  ),
                  onChanged: _onAmountChanged,
                ),
                const SizedBox(height: 8),
                Text(
                  _requestedAmount != null
                      ? 'QR akan berisi: ${formatYTE(_requestedAmount!)}'
                      : 'Kosongkan jika pengirim yang menentukan jumlah',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: AppTheme.darkTextSecondary,
                      ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 24),

          // Instructions
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: AppTheme.primary.withValues(alpha: 0.08),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: AppTheme.primary.withValues(alpha: 0.2)),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    const Icon(Icons.info_outline,
                        size: 18, color: AppTheme.primary),
                    const SizedBox(width: 8),
                    Text(
                      'Cara Menggunakan',
                      style: Theme.of(context).textTheme.titleSmall?.copyWith(
                            color: AppTheme.primary,
                          ),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                _instructionStep('1', 'Tunjukkan QR code ini ke pengirim'),
                _instructionStep('2',
                    'Pengirim scan QR menggunakan tombol scan di halaman Kirim'),
                _instructionStep('3',
                    'Address dan jumlah (jika diisi) akan otomatis terisi di form pengirim'),
                _instructionStep('4',
                    'Pengirim menandatangani transaksi dan mengirim'),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _instructionStep(String number, String text) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            width: 24,
            height: 24,
            decoration: BoxDecoration(
              color: AppTheme.primary.withValues(alpha: 0.15),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Center(
              child: Text(
                number,
                style: const TextStyle(
                  color: AppTheme.primary,
                  fontSize: 12,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Text(text, style: const TextStyle(fontSize: 13)),
          ),
        ],
      ),
    );
  }
}
