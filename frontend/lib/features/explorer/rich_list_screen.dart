import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../core/api/api_client.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/utils/formatters.dart';
import '../../shared/widgets/app_widgets.dart';

/// Rich List screen — top holders berdasarkan YTE balance.
///
/// Endpoint: GET /explorer/richlist?limit=100
class RichListScreen extends StatefulWidget {
  const RichListScreen({super.key});

  @override
  State<RichListScreen> createState() => _RichListScreenState();
}

class _RichListScreenState extends State<RichListScreen> {
  final _api = ApiClient();
  List<Map<String, dynamic>> _holders = [];
  bool _isLoading = true;
  String? _error;
  int _limit = 100;

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
      final resp = await _api.getRichList(limit: _limit);
      if (!mounted) return;
      final data = resp.data['data'] as Map<String, dynamic>;
      final holders = (data['holders'] as List<dynamic>)
          .map((e) => e as Map<String, dynamic>)
          .toList();
      setState(() {
        _holders = holders;
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
        title: const Text('Rich List — Top Holders'),
        actions: [
          PopupMenuButton<int>(
            icon: const Icon(Icons.filter_list),
            onSelected: (value) {
              setState(() => _limit = value);
              _loadData();
            },
            itemBuilder: (_) => [
              const PopupMenuItem(value: 10, child: Text('Top 10')),
              const PopupMenuItem(value: 50, child: Text('Top 50')),
              const PopupMenuItem(value: 100, child: Text('Top 100')),
              const PopupMenuItem(value: 500, child: Text('Top 500')),
            ],
          ),
        ],
      ),
      body: _isLoading
          ? const AppLoading()
          : _error != null
              ? AppError(message: _error!, onRetry: _loadData)
              : _buildList(),
    );
  }

  Widget _buildList() {
    if (_holders.isEmpty) {
      return const EmptyState(
        icon: Icons.leaderboard,
        title: 'Belum ada data',
        subtitle: 'Rich list akan muncul setelah ada wallet dengan saldo',
      );
    }

    return Column(
      children: [
        // Header info
        Container(
          padding: const EdgeInsets.all(16),
          color: AppTheme.darkSurface,
          child: Row(
            children: [
              const Icon(Icons.leaderboard, size: 20, color: AppTheme.primary),
              const SizedBox(width: 8),
              Text(
                '${_holders.length} holder teratas',
                style: const TextStyle(fontWeight: FontWeight.w600),
              ),
              const Spacer(),
              Text(
                'Top $_limit',
                style: const TextStyle(
                    color: AppTheme.darkTextSecondary, fontSize: 12),
              ),
            ],
          ),
        ),
        // Table header
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          color: AppTheme.darkCard,
          child: const Row(
            children: [
              SizedBox(width: 40, child: Text('#', style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold))),
              Expanded(child: Text('Address', style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold))),
              SizedBox(
                width: 120,
                child: Text('YTE Balance',
                    textAlign: TextAlign.right,
                    style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold)),
              ),
            ],
          ),
        ),
        // List
        Expanded(
          child: ListView.separated(
            itemCount: _holders.length,
            separatorBuilder: (_, __) => const Divider(height: 1),
            itemBuilder: (context, index) {
              final holder = _holders[index];
              final rank = holder['rank'] as int;
              final address = holder['address'] as String;
              final balance = (holder['yte_balance'] as num).toDouble();

              return _RichListTile(
                rank: rank,
                address: address,
                balance: balance,
              );
            },
          ),
        ),
      ],
    );
  }
}

class _RichListTile extends StatelessWidget {
  final int rank;
  final String address;
  final double balance;

  const _RichListTile({
    required this.rank,
    required this.address,
    required this.balance,
  });

  @override
  Widget build(BuildContext context) {
    final isTop3 = rank <= 3;
    final rankColor = switch (rank) {
      1 => const Color(0xFFFFD700), // Gold
      2 => const Color(0xFFC0C0C0), // Silver
      3 => const Color(0xFFCD7F32), // Bronze
      _ => AppTheme.darkTextSecondary,
    };

    return ListTile(
      dense: true,
      leading: Container(
        width: 32,
        height: 32,
        decoration: BoxDecoration(
          color: rankColor.withValues(alpha: 0.15),
          borderRadius: BorderRadius.circular(6),
        ),
        child: Center(
          child: isTop3
              ? Icon(Icons.emoji_events, color: rankColor, size: 18)
              : Text(
                  '$rank',
                  style: TextStyle(
                    color: rankColor,
                    fontWeight: FontWeight.bold,
                    fontSize: 12,
                  ),
                ),
        ),
      ),
      title: Text(
        shortAddress(address),
        style: const TextStyle(fontFamily: 'monospace', fontSize: 13),
      ),
      subtitle: Text(
        address,
        style: const TextStyle(
            fontFamily: 'monospace', fontSize: 10, color: AppTheme.darkTextSecondary),
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
      ),
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            formatYTE(balance),
            style: TextStyle(
              fontWeight: isTop3 ? FontWeight.bold : FontWeight.w600,
              fontSize: 13,
            ),
          ),
          const SizedBox(width: 4),
          InkWell(
            onTap: () {
              Clipboard.setData(ClipboardData(text: address));
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(content: Text('Address disalin')),
              );
            },
            child: const Icon(Icons.copy, size: 14, color: AppTheme.darkTextSecondary),
          ),
        ],
      ),
    );
  }
}
