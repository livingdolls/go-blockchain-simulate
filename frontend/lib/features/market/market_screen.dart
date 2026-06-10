import 'dart:async';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import '../../core/api/api_client.dart';
import '../../core/config/app_config.dart';
import '../../core/models/models.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/utils/formatters.dart';
import '../../shared/widgets/app_widgets.dart';

/// Market screen dengan real-time price updates via WebSocket.
class MarketScreen extends StatefulWidget {
  const MarketScreen({super.key});

  @override
  State<MarketScreen> createState() => _MarketScreenState();
}

class _MarketScreenState extends State<MarketScreen> {
  final _api = ApiClient();
  MarketState? _market;
  List<Candle> _candles = [];
  WebSocketChannel? _channel;
  bool _isLoading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadData();
    _connectWebSocket();
  }

  @override
  void dispose() {
    _channel?.sink.close();
    super.dispose();
  }

  Future<void> _loadData() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final results = await Future.wait([
        _api.getMarketState(),
        _api.getCandles(interval: '1h', type: 'all'),
      ]);

      setState(() {
        _market = MarketState.fromJson(
            results[0].data['data'] as Map<String, dynamic>);
        _candles = (results[1].data['data'] as List<dynamic>)
            .map((c) => Candle.fromJson(c as Map<String, dynamic>))
            .toList();
        _isLoading = false;
      });
    } on ApiException catch (e) {
      setState(() {
        _error = e.message;
        _isLoading = false;
      });
    }
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
          'events': ['market.update'],
        },
      }));

      _channel!.stream.listen(
        (data) {
          try {
            final msg = jsonDecode(data as String) as Map<String, dynamic>;
            final type = msg['type'] as String?;

            if (type == 'market.update' && mounted) {
              final priceData = msg['data'] as Map<String, dynamic>;
              setState(() {
                _market = MarketState.fromJson(priceData);
              });
            }
          } catch (_) {}
        },
        onError: (_) {},
        onDone: () {
          // Reconnect setelah delay
          Future.delayed(const Duration(seconds: 5), () {
            if (mounted) _connectWebSocket();
          });
        },
      );
    } catch (_) {}
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Market'),
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
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        // Price card
        if (_market != null) _buildPriceCard(),
        const SizedBox(height: 16),

        // Market stats
        if (_market != null) _buildMarketStats(),
        const SizedBox(height: 16),

        // Recent candles
        if (_candles.isNotEmpty) _buildCandleList(),
      ],
    );
  }

  Widget _buildPriceCard() {
    final m = _market!;
    final isUp = m.change24h >= 0;

    return AppCard(
      child: Column(
        children: [
          Text(
            'YTE/USD',
            style: Theme.of(context).textTheme.bodySmall?.copyWith(
                  color: AppTheme.darkTextSecondary,
                ),
          ),
          const SizedBox(height: 8),
          Text(
            formatUSD(m.price),
            style: Theme.of(context).textTheme.displaySmall?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
          ),
          const SizedBox(height: 8),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
            decoration: BoxDecoration(
              color: (isUp ? AppTheme.success : AppTheme.error).withValues(alpha: 0.15),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Text(
              formatPercent(m.change24h),
              style: TextStyle(
                color: isUp ? AppTheme.success : AppTheme.error,
                fontWeight: FontWeight.bold,
                fontSize: 16,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildMarketStats() {
    final m = _market!;
    return GridView.count(
      crossAxisCount: 2,
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      mainAxisSpacing: 12,
      crossAxisSpacing: 12,
      childAspectRatio: 1.8,
      children: [
        StatCard(
          label: 'High 24h',
          value: formatUSD(m.high24h),
          icon: Icons.arrow_upward,
          color: AppTheme.success,
        ),
        StatCard(
          label: 'Low 24h',
          value: formatUSD(m.low24h),
          icon: Icons.arrow_downward,
          color: AppTheme.error,
        ),
        StatCard(
          label: 'Volume 24h',
          value: formatUSD(m.volume24h),
          icon: Icons.bar_chart,
        ),
        StatCard(
          label: 'Change',
          value: formatPercent(m.change24h),
          icon: Icons.trending_up,
          color: m.change24h >= 0 ? AppTheme.success : AppTheme.error,
        ),
      ],
    );
  }

  Widget _buildCandleList() {
    return AppCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Candle Data (1h)', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 8),
          ..._candles.take(10).map((c) => _buildCandleRow(c)),
        ],
      ),
    );
  }

  Widget _buildCandleRow(Candle c) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        children: [
          SizedBox(
            width: 80,
            child: Text(
              '${c.dateTime.hour}:00',
              style: const TextStyle(fontFamily: 'monospace', fontSize: 12),
            ),
          ),
          Expanded(
            child: Row(
              children: [
                _candleValue('O', c.open),
                _candleValue('H', c.high),
                _candleValue('L', c.low),
                _candleValue('C', c.close),
              ],
            ),
          ),
          SizedBox(
            width: 80,
            child: Text(
              formatNumber(c.volume, decimals: 0),
              textAlign: TextAlign.right,
              style: TextStyle(
                color: c.isBullish ? AppTheme.success : AppTheme.error,
                fontSize: 12,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _candleValue(String label, double value) {
    return Expanded(
      child: Text(
        '$label: ${value.toStringAsFixed(2)}',
        style: const TextStyle(fontFamily: 'monospace', fontSize: 11),
      ),
    );
  }
}
