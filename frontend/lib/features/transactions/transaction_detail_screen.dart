import 'package:flutter/material.dart';
import '../../core/api/api_client.dart';
import '../../core/models/models.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/utils/formatters.dart';
import '../../shared/widgets/app_widgets.dart';

class TransactionDetailScreen extends StatefulWidget {
  final int transactionId;
  const TransactionDetailScreen({super.key, required this.transactionId});

  @override
  State<TransactionDetailScreen> createState() =>
      _TransactionDetailScreenState();
}

class _TransactionDetailScreenState extends State<TransactionDetailScreen> {
  final _api = ApiClient();
  Transaction? _tx;
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadTx();
  }

  Future<void> _loadTx() async {
    try {
      final resp = await _api.getTransaction(widget.transactionId);
      if (!mounted) return;
      setState(() {
        _tx = Transaction.fromJson(
            resp.data['data']['transaction'] as Map<String, dynamic>);
        _isLoading = false;
      });
    } on ApiException catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.message;
        _isLoading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = 'Terjadi kesalahan: $e';
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
          title:
              Text(_tx != null ? 'Tx #${_tx!.id}' : 'Transaction Detail')),
      body: _isLoading
          ? const AppLoading()
          : _error != null
              ? AppError(message: _error!, onRetry: _loadTx)
              : _buildContent(),
    );
  }

  Widget _buildContent() {
    final tx = _tx!;
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        AppCard(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text('Transaction #${tx.id}',
                      style: Theme.of(context).textTheme.titleLarge),
                  StatusBadge(status: tx.status),
                ],
              ),
              const Divider(),
              _infoRow('Type', tx.typeLabel),
              _infoRow('Amount', formatYTE(tx.amount)),
              _infoRow('Fee', formatNumber(tx.fee)),
              _infoRow('From', tx.fromAddress),
              _infoRow('To', tx.toAddress),
              _infoRow('Signature', tx.signature),
              _infoRow('Created', formatISODate(tx.createdAt)),
            ],
          ),
        ),
      ],
    );
  }

  Widget _infoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 100,
            child:
                Text(label, style: const TextStyle(color: AppTheme.darkTextSecondary)),
          ),
          Expanded(
            child: SelectableText(
              value,
              style: const TextStyle(fontFamily: 'monospace', fontSize: 13),
            ),
          ),
        ],
      ),
    );
  }
}
