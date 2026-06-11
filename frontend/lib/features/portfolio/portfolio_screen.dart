import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart' hide Block;
import '../../core/api/api_client.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/utils/formatters.dart';
import '../../shared/widgets/app_widgets.dart';

/// Portfolio Analytics screen — P&L tracking, asset allocation.
///
/// Endpoint: GET /portfolio/:address
class PortfolioScreen extends StatefulWidget {
  final String address;
  const PortfolioScreen({super.key, required this.address});

  @override
  State<PortfolioScreen> createState() => _PortfolioScreenState();
}

class _PortfolioScreenState extends State<PortfolioScreen> {
  final _api = ApiClient();
  Map<String, dynamic>? _portfolio;
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
      final resp = await _api.getPortfolio(widget.address);
      if (!mounted) return;
      setState(() {
        _portfolio = resp.data['data'] as Map<String, dynamic>;
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
        title: const Text('Portfolio'),
        actions: [
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
    return RefreshIndicator(
      onRefresh: _loadData,
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Total value card
          _buildTotalValueCard(),
          const SizedBox(height: 16),

          // P&L card
          _buildPnLCard(),
          const SizedBox(height: 16),

          // Asset allocation
          _buildAllocationCard(),
          const SizedBox(height: 16),

          // Trading stats
          _buildTradingStats(),
        ],
      ),
    );
  }

  Widget _buildTotalValueCard() {
    final totalValue = (_portfolio!['total_value_usd'] as num?)?.toDouble() ?? 0;
    final yteBalance = (_portfolio!['yte_balance'] as num?)?.toDouble() ?? 0;
    final usdBalance = (_portfolio!['usd_balance'] as num?)?.toDouble() ?? 0;
    final ytePrice = (_portfolio!['yte_price'] as num?)?.toDouble() ?? 0;

    return AppCard(
      child: Column(
        children: [
          Text(
            'Total Portfolio Value',
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: AppTheme.darkTextSecondary,
                ),
          ),
          const SizedBox(height: 8),
          Text(
            formatUSD(totalValue),
            style: Theme.of(context).textTheme.headlineLarge?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
          ),
          const SizedBox(height: 4),
          Text(
            'YTE Price: ${formatUSD(ytePrice)}',
            style: const TextStyle(
                color: AppTheme.darkTextSecondary, fontSize: 12),
          ),
          const SizedBox(height: 16),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceEvenly,
            children: [
              _balanceItem('YTE', formatNumber(yteBalance, decimals: 4)),
              _balanceItem('USD', formatUSD(usdBalance)),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildPnLCard() {
    final realized = (_portfolio!['realized_pnl'] as num?)?.toDouble() ?? 0;
    final unrealized = (_portfolio!['unrealized_pnl'] as num?)?.toDouble() ?? 0;
    final pnlPercent = (_portfolio!['pnl_percent'] as num?)?.toDouble() ?? 0;

    final isPositive = pnlPercent >= 0;
    final color = isPositive ? AppTheme.success : AppTheme.error;

    return AppCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.trending_up, size: 20, color: AppTheme.primary),
              const SizedBox(width: 8),
              Text('Profit & Loss',
                  style: Theme.of(context).textTheme.titleMedium),
            ],
          ),
          const SizedBox(height: 16),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('Total P&L',
                      style: TextStyle(
                          color: AppTheme.darkTextSecondary, fontSize: 12)),
                  Text(
                    '${isPositive ? "+" : ""}${formatPercent(pnlPercent)}',
                    style: TextStyle(
                      color: color,
                      fontWeight: FontWeight.bold,
                      fontSize: 20,
                    ),
                  ),
                ],
              ),
              Column(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  _pnlRow('Realized', realized),
                  _pnlRow('Unrealized', unrealized),
                ],
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildAllocationCard() {
    final allocation = _portfolio!['allocation'] as Map<String, dynamic>? ?? {};
    final ytePercent = (allocation['yte_percent'] as num?)?.toDouble() ?? 0;
    final usdPercent = (allocation['usd_percent'] as num?)?.toDouble() ?? 0;

    return AppCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.pie_chart, size: 20, color: AppTheme.primary),
              const SizedBox(width: 8),
              Text('Asset Allocation',
                  style: Theme.of(context).textTheme.titleMedium),
            ],
          ),
          const SizedBox(height: 16),
          // Simple bar visualization
          ClipRRect(
            borderRadius: BorderRadius.circular(4),
            child: Row(
              children: [
                Expanded(
                  flex: ytePercent.round().clamp(1, 99),
                  child: Container(
                    height: 24,
                    color: AppTheme.primary,
                    alignment: Alignment.center,
                    child: Text(
                      'YTE ${ytePercent.toStringAsFixed(1)}%',
                      style: const TextStyle(
                          color: Colors.white,
                          fontSize: 11,
                          fontWeight: FontWeight.w600),
                    ),
                  ),
                ),
                Expanded(
                  flex: usdPercent.round().clamp(1, 99),
                  child: Container(
                    height: 24,
                    color: AppTheme.success,
                    alignment: Alignment.center,
                    child: Text(
                      'USD ${usdPercent.toStringAsFixed(1)}%',
                      style: const TextStyle(
                          color: Colors.white,
                          fontSize: 11,
                          fontWeight: FontWeight.w600),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTradingStats() {
    final deposited =
        (_portfolio!['total_deposited'] as num?)?.toDouble() ?? 0;
    final withdrawn =
        (_portfolio!['total_withdrawn'] as num?)?.toDouble() ?? 0;
    final traded = (_portfolio!['total_traded'] as num?)?.toDouble() ?? 0;

    return GridView.count(
      crossAxisCount: 3,
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      mainAxisSpacing: 8,
      crossAxisSpacing: 8,
      childAspectRatio: 1.4,
      children: [
        StatCard(
          label: 'Deposited',
          value: formatUSD(deposited),
          icon: Icons.arrow_downward,
          color: AppTheme.success,
        ),
        StatCard(
          label: 'Withdrawn',
          value: formatUSD(withdrawn),
          icon: Icons.arrow_upward,
          color: AppTheme.error,
        ),
        StatCard(
          label: 'Traded',
          value: formatUSD(traded),
          icon: Icons.swap_horiz,
          color: AppTheme.primary,
        ),
      ],
    );
  }

  Widget _balanceItem(String label, String value) {
    return Column(
      children: [
        Text(label, style: const TextStyle(color: AppTheme.darkTextSecondary, fontSize: 12)),
        Text(value, style: const TextStyle(fontWeight: FontWeight.bold)),
      ],
    );
  }

  Widget _pnlRow(String label, double value) {
    final color = value >= 0 ? AppTheme.success : AppTheme.error;
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Text(
        '$label: ${formatUSD(value)}',
        style: TextStyle(color: color, fontSize: 12),
      ),
    );
  }
}
