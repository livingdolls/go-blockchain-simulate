import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart' hide Block;
import '../../core/api/api_client.dart';
import '../../core/models/models.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/utils/formatters.dart';
import '../../shared/widgets/app_widgets.dart';

/// Mempool Viewer — menampilkan pending transactions (transaksi yang
/// sudah dikirim tapi belum dikonfirmasi di block).
///
/// Berguna untuk:
/// - Monitoring congestion (berapa banyak tx menunggu)
/// - Fee estimation (tx dengan fee tinggi diproses duluan)
/// - Debugging (tx stuck karena nonce salah, dll)
///
/// Endpoint: GET /transactions/pending?limit=50
class MempoolScreen extends StatefulWidget {
  const MempoolScreen({super.key});

  @override
  State<MempoolScreen> createState() => _MempoolScreenState();
}

class _MempoolScreenState extends State<MempoolScreen> {
  final _api = ApiClient();
  List<Transaction> _transactions = [];
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Future<void> _loadData() async {
    if (!mounted) return;
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final resp = await _api.getPendingTransactions(limit: 100);
      if (!mounted) return;
      final data = resp.data['data'] as Map<String, dynamic>;
      final txs = (data['transactions'] as List<dynamic>?)
              ?.map((t) => Transaction.fromJson(t as Map<String, dynamic>))
              .toList() ??
          [];
      setState(() {
        _transactions = txs;
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
        title: const Text('Mempool — Pending Transactions'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadData,
          ),
        ],
      ),
      body: _isLoading
          ? const AppLoading()
          : _error != null
              ? AppError(message: _error!, onRetry: _loadData)
              : _buildContent(),
    );
  }

  Widget _buildContent() {
    if (_transactions.isEmpty) {
      return const EmptyState(
        icon: Icons.check_circle,
        title: 'Mempool kosong',
        subtitle: 'Tidak ada transaksi yang menunggu konfirmasi',
      );
    }

    return Column(
      children: [
        // Summary bar
        Container(
          padding: const EdgeInsets.all(16),
          color: AppTheme.darkSurface,
          child: Row(
            children: [
              const Icon(Icons.hourglass_empty, size: 20, color: AppTheme.warning),
              const SizedBox(width: 8),
              Text(
                '${_transactions.length} transaksi menunggu konfirmasi',
                style: const TextStyle(fontWeight: FontWeight.w600),
              ),
              const Spacer(),
              _congestionBadge(),
            ],
          ),
        ),
        // Transaction list
        Expanded(
          child: RefreshIndicator(
            onRefresh: _loadData,
            child: ListView.separated(
              padding: const EdgeInsets.all(16),
              itemCount: _transactions.length,
              separatorBuilder: (_, __) => const Divider(height: 1),
              itemBuilder: (context, index) {
                final tx = _transactions[index];
                return _MempoolTile(tx: tx);
              },
            ),
          ),
        ),
      ],
    );
  }

  Widget _congestionBadge() {
    final count = _transactions.length;
    final (color, label) = switch (count) {
      <= 10 => (AppTheme.success, 'Low'),
      <= 50 => (AppTheme.warning, 'Medium'),
      <= 100 => (AppTheme.error, 'High'),
      _ => (AppTheme.error, 'Very High'),
    };

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.15),
        borderRadius: BorderRadius.circular(4),
      ),
      child: Text(
        'Congestion: $label',
        style: TextStyle(color: color, fontSize: 11, fontWeight: FontWeight.bold),
      ),
    );
  }
}

class _MempoolTile extends StatelessWidget {
  final Transaction tx;
  const _MempoolTile({required this.tx});

  @override
  Widget build(BuildContext context) {
    final typeColor = switch (tx.type) {
      'TRANSFER' => AppTheme.primary,
      'BUY' => AppTheme.success,
      'SELL' => AppTheme.warning,
      _ => AppTheme.darkTextSecondary,
    };

    return ListTile(
      dense: true,
      leading: Container(
        width: 36,
        height: 36,
        decoration: BoxDecoration(
          color: typeColor.withValues(alpha: 0.15),
          borderRadius: BorderRadius.circular(6),
        ),
        child: Center(
          child: Text(
            tx.type[0],
            style: TextStyle(color: typeColor, fontWeight: FontWeight.bold),
          ),
        ),
      ),
      title: Text(
        '${tx.typeLabel} — ${formatYTE(tx.amount)}',
        style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
      ),
      subtitle: Text(
        '${shortAddress(tx.fromAddress)} → ${shortAddress(tx.toAddress)}',
        style: const TextStyle(fontFamily: 'monospace', fontSize: 11),
      ),
      trailing: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          Text(
            'Fee: ${formatNumber(tx.fee)}',
            style: const TextStyle(fontSize: 11, color: AppTheme.darkTextSecondary),
          ),
          Text(
            'ID: ${tx.id}',
            style: const TextStyle(fontSize: 10, color: AppTheme.darkTextSecondary),
          ),
        ],
      ),
      onTap: () => context.go('/transactions/${tx.id}'),
    );
  }
}
