import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart' hide Block;
import '../../core/api/api_client.dart';
import '../../core/models/models.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/utils/formatters.dart';
import '../../shared/widgets/app_widgets.dart';
import '../qr/receive_screen.dart';

class WalletScreen extends StatefulWidget {
  final String address;
  const WalletScreen({super.key, required this.address});

  @override
  State<WalletScreen> createState() => _WalletScreenState();
}

class _WalletScreenState extends State<WalletScreen> {
  final _api = ApiClient();
  UserBalance? _balance;
  TransactionWithTypeResponse? _transactions;
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Future<void> _loadData() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final results = await Future.wait([
        _api.getBalance(widget.address),
        _api.getWallet(widget.address, limit: 20),
      ]);

      setState(() {
        _balance = UserBalance.fromJson(
            results[0].data['data'] as Map<String, dynamic>);
        _transactions = TransactionWithTypeResponse.fromJson(
            results[1].data as Map<String, dynamic>);
        _isLoading = false;
      });
    } on ApiException catch (e) {
      setState(() {
        _error = e.message;
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(shortAddress(widget.address)),
        actions: [
          IconButton(
            icon: const Icon(Icons.qr_code),
            onPressed: () {
              Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (_) => ReceiveScreen(address: widget.address),
                ),
              );
            },
            tooltip: 'Terima YTE (QR)',
          ),
          IconButton(icon: const Icon(Icons.refresh), onPressed: _loadData),
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
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        // Balance cards
        if (_balance != null) ...[
          Row(
            children: [
              Expanded(
                child: StatCard(
                  label: 'YTE Balance',
                  value: formatYTE(_balance!.yteBalance),
                  icon: Icons.token,
                  color: AppTheme.primary,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: StatCard(
                  label: 'USD Balance',
                  value: formatUSD(_balance!.usdBalance),
                  icon: Icons.attach_money,
                  color: AppTheme.accent,
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
        ],

        // Transactions
        Text('Riwayat Transaksi', style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: 8),
        if (_transactions != null && _transactions!.transactions.isNotEmpty)
          ..._transactions!.transactions.map((tx) => _buildTxCard(tx))
        else
          const EmptyState(
            icon: Icons.receipt_long,
            title: 'Belum ada transaksi',
          ),
      ],
    );
  }

  Widget _buildTxCard(Transaction tx) {
    final isSend =
        tx.fromAddress.toLowerCase() == widget.address.toLowerCase();
    final icon = isSend ? Icons.arrow_upward : Icons.arrow_downward;
    final color = isSend ? AppTheme.error : AppTheme.success;

    return AppCard(
      margin: const EdgeInsets.only(bottom: 8),
      onTap: () => context.go('/transactions/${tx.id}'),
      child: Row(
        children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(
              color: color.withValues(alpha: 0.15),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Icon(icon, color: color, size: 20),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(tx.typeLabel, style: const TextStyle(fontWeight: FontWeight.w600)),
                Text(
                  '${shortAddress(isSend ? tx.toAddress : tx.fromAddress)} • ${formatRelativeTime(DateTime.parse(tx.createdAt).millisecondsSinceEpoch ~/ 1000)}',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: AppTheme.darkTextSecondary,
                      ),
                ),
              ],
            ),
          ),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Text(
                '${isSend ? "-" : "+"}${formatYTE(tx.amount)}',
                style: TextStyle(
                  fontWeight: FontWeight.bold,
                  color: color,
                ),
              ),
              StatusBadge(status: tx.status),
            ],
          ),
        ],
      ),
    );
  }
}
