import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart' hide Block;
import '../../core/api/api_client.dart';
import '../../core/models/models.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/utils/formatters.dart';
import '../../shared/widgets/app_widgets.dart';

class BlockDetailScreen extends StatefulWidget {
  final int blockId;
  const BlockDetailScreen({super.key, required this.blockId});

  @override
  State<BlockDetailScreen> createState() => _BlockDetailScreenState();
}

class _BlockDetailScreenState extends State<BlockDetailScreen> {
  final _api = ApiClient();
  Block? _block;
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadBlock();
  }

  Future<void> _loadBlock() async {
    try {
      final resp = await _api.getBlockById(widget.blockId);
      if (!mounted) return;
      setState(() {
        _block = Block.fromJson(resp.data['data'] as Map<String, dynamic>);
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
        title: Text(_block != null ? 'Block #${_block!.blockNumber}' : 'Block Detail'),
      ),
      body: _isLoading
          ? const AppLoading()
          : _error != null
              ? AppError(message: _error!, onRetry: _loadBlock)
              : _buildContent(),
    );
  }

  Widget _buildContent() {
    final b = _block!;
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        // Block info card
        AppCard(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text('Block #${b.blockNumber}',
                  style: Theme.of(context).textTheme.titleLarge),
              const Divider(),
              _infoRow('Hash', b.currentHash),
              _infoRow('Previous Hash', b.previousHash),
              _infoRow('Nonce', '${b.nonce}'),
              _infoRow('Difficulty', '${b.difficulty}'),
              _infoRow('Timestamp', formatDate(b.timestamp)),
              _infoRow('Merkle Root', b.merkleRoot),
              _infoRow('Miner', b.minerAddress),
              _infoRow('Block Reward', formatYTE(b.blockReward)),
              _infoRow('Total Fees', formatNumber(b.totalFees)),
            ],
          ),
        ),
        const SizedBox(height: 16),

        // Transactions
        if (b.transactions != null && b.transactions!.isNotEmpty) ...[
          Text('Transaksi (${b.transactions!.length})',
              style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 8),
          ...b.transactions!.map((tx) => _buildTxCard(tx)),
        ],
      ],
    );
  }

  Widget _buildTxCard(Transaction tx) {
    return AppCard(
      margin: const EdgeInsets.only(bottom: 8),
      onTap: () => context.go('/transactions/${tx.id}'),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text('Tx #${tx.id}', style: const TextStyle(fontWeight: FontWeight.bold)),
              StatusBadge(status: tx.status),
            ],
          ),
          const Divider(),
          Text('${tx.typeLabel}: ${formatYTE(tx.amount)}',
              style: const TextStyle(fontWeight: FontWeight.w600)),
          const SizedBox(height: 4),
          Text('From: ${shortAddress(tx.fromAddress)}',
              style: const TextStyle(fontFamily: 'monospace', fontSize: 12)),
          Text('To: ${shortAddress(tx.toAddress)}',
              style: const TextStyle(fontFamily: 'monospace', fontSize: 12)),
          Text('Fee: ${formatNumber(tx.fee)}',
              style: Theme.of(context).textTheme.bodySmall),
        ],
      ),
    );
  }

  Widget _infoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 120,
            child: Text(label, style: const TextStyle(color: AppTheme.darkTextSecondary)),
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
