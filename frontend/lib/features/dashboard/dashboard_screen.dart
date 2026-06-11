import 'dart:async';
import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart' hide Block;
import 'package:provider/provider.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import '../../core/api/api_client.dart';
import '../../core/config/app_config.dart';
import '../../core/models/models.dart';
import '../../core/models/notification.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/utils/formatters.dart';
import '../../shared/widgets/app_widgets.dart';
import '../../shared/widgets/notification_bell.dart';
import '../notifications/notification_provider.dart';
import '../search/search_delegate.dart';

/// Dashboard utama yang menampilkan ringkasan blockchain:
/// - Block stats (total blocks, difficulty, fees)
/// - Reward info (current supply, halving)
/// - Market state (price, volume)
/// - Block terbaru
/// - Real-time updates via WebSocket
class DashboardScreen extends StatefulWidget {
  const DashboardScreen({super.key});

  @override
  State<DashboardScreen> createState() => _DashboardScreenState();
}

class _DashboardScreenState extends State<DashboardScreen> {
  final _api = ApiClient();

  BlockStats? _stats;
  RewardInfo? _reward;
  MarketState? _market;
  List<Block>? _recentBlocks;
  Map<String, dynamic>? _congestion;
  bool _isLoading = true;
  bool _isLive = false;
  String? _error;

  WebSocketChannel? _channel;
  Timer? _reconnectTimer;

  @override
  void initState() {
    super.initState();
    _loadData();
    _connectWebSocket();
  }

  @override
  void dispose() {
    _reconnectTimer?.cancel();
    _channel?.sink.close();
    super.dispose();
  }

  Future<void> _loadData() async {
    if (!mounted) return;
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final results = await Future.wait([
        _api.getBlockStats(),
        _api.getRewardInfo(),
        _api.getMarketState(),
        _api.getBlocks(limit: 5),
        _api.estimateFee(amount: 100, priority: 'low'),
      ]);

      if (!mounted) return;
      setState(() {
        _stats = BlockStats.fromJson(
            results[0].data['data'] as Map<String, dynamic>);
        _reward = RewardInfo.fromJson(
            results[1].data['data'] as Map<String, dynamic>);
        _market = MarketState.fromJson(
            results[2].data['data'] as Map<String, dynamic>);
        _recentBlocks = (results[3].data['data'] as List<dynamic>)
            .map((b) => Block.fromJson(b as Map<String, dynamic>))
            .toList();
        _congestion = results[4].data['data'] as Map<String, dynamic>;
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

  Future<void> _logout() async {
    try {
      await _api.userLogout();
    } catch (_) {
      // Logout best-effort: bahkan jika server error, tetap clear local state
    }
    _reconnectTimer?.cancel();
    _channel?.sink.close();
    if (mounted) context.go('/login');
  }

  void _connectWebSocket() {
    try {
      _channel = WebSocketChannel.connect(
        Uri.parse('${AppConfig.wsUrl}/ws/market'),
      );

      // Kirim subscribe message setelah connect.
      // Tanpa ini, backend tidak akan mengirim event apapun karena
      // setiap client harus subscribe ke event types yang diinginkan.
      _channel!.sink.add(jsonEncode({
        'type': 'subscribe',
        'data': {
          'events': [
            'market.update',
            'block.mined',
            'notification.TRANSACTION_CONFIRMED',
            'notification.TRANSACTION_SUBMITTED',
            'notification.BLOCK_CONFIRMED',
            'notification.BALANCE_UPDATED',
            'notification.REWARD_EARNED',
          ],
        },
      }));

      _channel!.stream.listen(
        (data) {
          try {
            final msg = jsonDecode(data as String) as Map<String, dynamic>;
            final type = msg['type'] as String?;
            final payload = msg['data'];

            if (payload is! Map<String, dynamic>) return;

            if (type == 'market.update' && mounted) {
              setState(() {
                _market = MarketState.fromJson(payload);
                _isLive = true;
              });
            } else if (type == 'block.mined' && mounted) {
              final newBlock = Block.fromJson(payload);
              setState(() {
                _recentBlocks = [
                  newBlock,
                  if (_recentBlocks != null) ..._recentBlocks!.take(4),
                ];
                _isLive = true;
              });
            } else if (type != null && type.startsWith('notification.') && mounted) {
              // Notification events dari NotificationWSConsumer
              final notification = AppNotification.fromWsMessage(msg);
              final provider = context.read<NotificationProvider>();
              provider.addNotification(notification);

              // Tampilkan toast/snackbar untuk notifikasi baru
              if (mounted) {
                showNotificationToast(context, notification);
              }
            }
          } catch (_) {}
        },
        onError: (e) {
          if (mounted) {
            setState(() => _isLive = false);
            debugPrint('Dashboard WS error: $e');
          }
        },
        onDone: () {
          if (mounted) {
            setState(() => _isLive = false);
            // Reconnect setelah 5 detik
            _reconnectTimer = Timer(const Duration(seconds: 5), () {
              if (mounted) _connectWebSocket();
            });
          }
        },
      );
    } catch (_) {
      if (mounted) setState(() => _isLive = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('YuteBlockchain Dashboard'),
        actions: [
          // Search button
          IconButton(
            icon: const Icon(Icons.search),
            onPressed: () {
              showSearch(context: context, delegate: BlockchainSearchDelegate());
            },
            tooltip: 'Cari',
          ),
          // Notification bell
          const NotificationBell(),
          // Live indicator
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 14),
            child: LiveIndicator(isLive: _isLive),
          ),
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadData,
          ),
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: _logout,
            tooltip: 'Logout',
          ),
        ],
      ),
      body: _isLoading
          ? const AppLoading(message: 'Memuat data blockchain...')
          : _error != null
              ? AppError(message: _error!, onRetry: _loadData)
              : _buildContent(),
    );
  }

  Widget _buildContent() {
    return RefreshIndicator(
      onRefresh: _loadData,
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Market price card
          if (_market != null) _buildMarketCard(),
          const SizedBox(height: 16),

          // Quick action buttons
          _buildActionButtons(),
          const SizedBox(height: 16),

          // Stats grid
          if (_stats != null) _buildStatsGrid(),
          const SizedBox(height: 16),

          // Network congestion indicator
          if (_congestion != null) _buildCongestionCard(),
          const SizedBox(height: 16),

          // Reward info
          if (_reward != null) _buildRewardCard(),
          const SizedBox(height: 16),

          // Recent blocks
          if (_recentBlocks != null && _recentBlocks!.isNotEmpty)
            _buildRecentBlocks(),
        ],
      ),
    );
  }

  Widget _buildActionButtons() {
    return Row(
      children: [
        Expanded(
          child: _ActionButton(
            icon: Icons.send,
            label: 'Kirim',
            color: AppTheme.primary,
            onTap: () => context.go('/transactions/send'),
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: _ActionButton(
            icon: Icons.shopping_cart,
            label: 'Beli',
            color: AppTheme.success,
            onTap: () => context.go('/transactions/buy'),
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: _ActionButton(
            icon: Icons.sell,
            label: 'Jual',
            color: AppTheme.warning,
            onTap: () => context.go('/transactions/sell'),
          ),
        ),
        const SizedBox(width: 8),
          Expanded(
            child: _ActionButton(
              icon: Icons.lock,
              label: 'Stake',
              color: AppTheme.accent,
              onTap: () => context.go('/staking'),
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: _ActionButton(
              icon: Icons.bar_chart,
              label: 'Portfolio',
              color: AppTheme.darkTextSecondary,
              onTap: () => context.go('/portfolio/all'),
            ),
          ),
      ],
    );
  }

  Widget _buildMarketCard() {
    final m = _market!;
    final isUp = m.change24h >= 0;

    return AppCard(
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Harga YTE',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: AppTheme.darkTextSecondary,
                      ),
                ),
                const SizedBox(height: 4),
                Text(
                  formatUSD(m.price),
                  style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                      ),
                ),
              ],
            ),
          ),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color:
                      (isUp ? AppTheme.success : AppTheme.error).withValues(alpha: 0.15),
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Text(
                  formatPercent(m.change24h),
                  style: TextStyle(
                    color: isUp ? AppTheme.success : AppTheme.error,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
              const SizedBox(height: 4),
              Text(
                'Vol: ${formatUSD(m.volume24h)}',
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: AppTheme.darkTextSecondary,
                    ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildCongestionCard() {
    final congestionLevel = _congestion!['congestion_level'] as String? ?? 'low';
    final pendingCount = _congestion!['pending_count'] as int? ?? 0;
    final congestionPct = (_congestion!['congestion_percent'] as num?)?.toDouble() ?? 0;
    final estimatedFee = (_congestion!['estimated_fee'] as num?)?.toDouble() ?? 0;
    final congestionMult = (_congestion!['congestion_multiplier'] as num?)?.toDouble() ?? 1;

    final congestionColor = switch (congestionLevel) {
      'low' => AppTheme.success,
      'medium' => AppTheme.warning,
      'high' => AppTheme.error,
      'very_high' => AppTheme.error,
      _ => AppTheme.darkTextSecondary,
    };

    return AppCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.network_check, size: 20, color: AppTheme.primary),
              const SizedBox(width: 8),
              Text('Network Congestion',
                  style: Theme.of(context).textTheme.titleMedium),
            ],
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              Expanded(
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(4),
                  child: LinearProgressIndicator(
                    value: congestionPct / 100,
                    backgroundColor: AppTheme.darkCard,
                    valueColor:
                        AlwaysStoppedAnimation<Color>(congestionColor),
                    minHeight: 8,
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: congestionColor.withValues(alpha: 0.15),
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Text(
                  congestionLevel.toUpperCase(),
                  style: TextStyle(
                    color: congestionColor,
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text('$pendingCount pending txns',
                  style: const TextStyle(
                      color: AppTheme.darkTextSecondary, fontSize: 12)),
              Text('Fee mult: ${congestionMult}x',
                  style: const TextStyle(
                      color: AppTheme.darkTextSecondary, fontSize: 12)),
              Text('Est: ${formatYTE(estimatedFee)}',
                  style: const TextStyle(fontSize: 12)),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildStatsGrid() {
    final s = _stats!;
    return GridView.count(
      crossAxisCount: 2,
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      mainAxisSpacing: 12,
      crossAxisSpacing: 12,
      childAspectRatio: 1.6,
      children: [
        StatCard(
          label: 'Total Blocks',
          value: formatNumber(s.totalBlocks, decimals: 0),
          icon: Icons.view_module,
        ),
        StatCard(
          label: 'Transactions',
          value: formatNumber(s.totalTransactions, decimals: 0),
          icon: Icons.receipt_long,
          subtitle: '${formatNumber(s.avgTxPerBlock, decimals: 1)} per block',
        ),
        StatCard(
          label: 'Avg Difficulty',
          value: s.averageDifficulty.toStringAsFixed(1),
          icon: Icons.speed,
        ),
        StatCard(
          label: 'Total Fees',
          value: formatYTE(s.totalFees),
          icon: Icons.monetization_on,
          color: AppTheme.accent,
        ),
      ],
    );
  }

  Widget _buildRewardCard() {
    final r = _reward!;
    return AppCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.token, color: AppTheme.warning, size: 20),
              const SizedBox(width: 8),
              Text(
                'Reward Info',
                style: Theme.of(context).textTheme.titleMedium,
              ),
            ],
          ),
          const SizedBox(height: 16),
          _infoRow('Current Reward', formatYTE(r.currentReward)),
          _infoRow('Next Reward', formatYTE(r.nextReward)),
          _infoRow('Supply', '${formatNumber(r.currentSupply)} / ${formatNumber(r.maxSupply)}'),
          _infoRow('Halving Block', '${r.nextHalvingBlock} (${r.blocksUntilHalving} blocks lagi)'),
          const SizedBox(height: 8),
          LinearProgressIndicator(
            value: r.supplyPercentage / 100,
            backgroundColor: AppTheme.darkCard,
            valueColor: const AlwaysStoppedAnimation(AppTheme.warning),
          ),
          const SizedBox(height: 4),
          Text(
            '${r.supplyPercentage.toStringAsFixed(1)}% supply tercapai',
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: AppTheme.darkTextSecondary,
                ),
          ),
        ],
      ),
    );
  }

  Widget _buildRecentBlocks() {
    return AppCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                'Block Terbaru',
                style: Theme.of(context).textTheme.titleMedium,
              ),
              TextButton(
                onPressed: () => context.go('/blocks'),
                child: const Text('Lihat Semua'),
              ),
            ],
          ),
          const SizedBox(height: 8),
          ...(_recentBlocks!.map((b) => _buildBlockRow(b))),
        ],
      ),
    );
  }

  Widget _buildBlockRow(Block b) {
    return InkWell(
      onTap: () => context.go('/blocks/${b.id}'),
      borderRadius: BorderRadius.circular(8),
      child: Padding(
        padding: const EdgeInsets.symmetric(vertical: 8),
        child: Row(
          children: [
            Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(
                color: AppTheme.primary.withValues(alpha: 0.15),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Center(
                child: Text(
                  '#${b.blockNumber}',
                  style: const TextStyle(
                    color: AppTheme.primary,
                    fontWeight: FontWeight.bold,
                    fontSize: 12,
                  ),
                ),
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    b.shortHash,
                    style: const TextStyle(fontFamily: 'monospace', fontSize: 13),
                  ),
                  Text(
                    '${b.transactions?.length ?? 0} txns • ${formatRelativeTime(b.timestamp)}',
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
                  formatYTE(b.blockReward),
                  style: const TextStyle(fontWeight: FontWeight.w600),
                ),
                Text(
                  'Fee: ${formatNumber(b.totalFees)}',
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: AppTheme.darkTextSecondary,
                      ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _infoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: TextStyle(color: AppTheme.darkTextSecondary)),
          Text(value, style: const TextStyle(fontWeight: FontWeight.w600)),
        ],
      ),
    );
  }
}

/// Action button untuk quick actions di dashboard.
class _ActionButton extends StatelessWidget {
  final IconData icon;
  final String label;
  final Color color;
  final VoidCallback onTap;

  const _ActionButton({
    required this.icon,
    required this.label,
    required this.color,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Material(
      color: color.withValues(alpha: 0.12),
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 16),
          child: Column(
            children: [
              Icon(icon, color: color, size: 28),
              const SizedBox(height: 8),
              Text(
                label,
                style: TextStyle(
                  color: color,
                  fontWeight: FontWeight.w600,
                  fontSize: 13,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
