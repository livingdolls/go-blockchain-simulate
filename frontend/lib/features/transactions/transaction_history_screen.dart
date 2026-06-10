import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart' hide Block;
import '../../core/api/api_client.dart';
import '../../core/models/models.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/utils/formatters.dart';
import '../../shared/widgets/app_widgets.dart';

/// Transaction History screen dengan filter lengkap.
///
/// Filter yang tersedia:
/// - Type: all, send, received, buy, sell
/// - Status: all, pending, confirmed
/// - Sort: id, amount
/// - Order: asc, desc
///
/// Endpoint: GET /wallet/:address?type=...&status=...&page=...&limit=...
class TransactionHistoryScreen extends StatefulWidget {
  final String address;
  const TransactionHistoryScreen({super.key, required this.address});

  @override
  State<TransactionHistoryScreen> createState() =>
      _TransactionHistoryScreenState();
}

class _TransactionHistoryScreenState extends State<TransactionHistoryScreen> {
  final _api = ApiClient();

  TransactionWithTypeResponse? _data;
  bool _isLoading = true;
  String? _error;

  // Filter state
  String _type = 'all';
  String _status = 'all';
  String _sortBy = 'id';
  String _order = 'desc';
  int _page = 1;
  final int _limit = 20;

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
      final resp = await _api.getWallet(
        widget.address,
        type: _type,
        status: _status,
        page: _page,
        limit: _limit,
        sortBy: _sortBy,
        order: _order,
      );

      if (!mounted) return;
      setState(() {
        _data = TransactionWithTypeResponse.fromJson(
            resp.data as Map<String, dynamic>);
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

  void _applyFilter() {
    setState(() => _page = 1);
    _loadData();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Riwayat Transaksi'),
        actions: [
          IconButton(
            icon: const Icon(Icons.filter_list),
            onPressed: _showFilterSheet,
          ),
        ],
      ),
      body: Column(
        children: [
          // Active filters chips
          if (_hasActiveFilters()) _buildActiveFilters(),

          // Transaction list
          Expanded(
            child: _isLoading
                ? const AppLoading()
                : _error != null
                    ? AppError(message: _error!, onRetry: _loadData)
                    : _buildList(),
          ),
        ],
      ),
    );
  }

  bool _hasActiveFilters() {
    return _type != 'all' || _status != 'all' || _sortBy != 'id' || _order != 'desc';
  }

  Widget _buildActiveFilters() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      color: AppTheme.darkSurface,
      child: Row(
        children: [
          if (_type != 'all')
            _filterChip(_typeLabel(_type), () {
              setState(() => _type = 'all');
              _applyFilter();
            }),
          if (_status != 'all')
            _filterChip(_status, () {
              setState(() => _status = 'all');
              _applyFilter();
            }),
          if (_sortBy != 'id')
            _filterChip('Sort: $_sortBy', () {
              setState(() => _sortBy = 'id');
              _applyFilter();
            }),
          if (_order != 'desc')
            _filterChip(_order.toUpperCase(), () {
              setState(() => _order = 'desc');
              _applyFilter();
            }),
          const Spacer(),
          TextButton(
            onPressed: () {
              setState(() {
                _type = 'all';
                _status = 'all';
                _sortBy = 'id';
                _order = 'desc';
                _page = 1;
              });
              _loadData();
            },
            child: const Text('Reset'),
          ),
        ],
      ),
    );
  }

  Widget _filterChip(String label, VoidCallback onRemove) {
    return Padding(
      padding: const EdgeInsets.only(right: 8),
      child: Chip(
        label: Text(label, style: const TextStyle(fontSize: 12)),
        deleteIcon: const Icon(Icons.close, size: 16),
        onDeleted: onRemove,
        backgroundColor: AppTheme.primary.withValues(alpha: 0.15),
        labelStyle: const TextStyle(color: AppTheme.primary),
        deleteIconColor: AppTheme.primary,
        materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
        visualDensity: VisualDensity.compact,
      ),
    );
  }

  Widget _buildList() {
    if (_data == null || _data!.transactions.isEmpty) {
      return const EmptyState(
        icon: Icons.receipt_long,
        title: 'Belum ada transaksi',
        subtitle: 'Transaksi akan muncul di sini setelah Anda mengirim atau menerima YTE',
      );
    }

    return Column(
      children: [
        Expanded(
          child: ListView.separated(
            padding: const EdgeInsets.all(16),
            itemCount: _data!.transactions.length,
            separatorBuilder: (_, __) => const Divider(height: 1),
            itemBuilder: (context, index) {
              final tx = _data!.transactions[index];
              return _TransactionTile(
                tx: tx,
                currentAddress: widget.address,
                onTap: () => context.go('/transactions/${tx.id}'),
              );
            },
          ),
        ),
        // Pagination
        if (_data!.totalPages > 1) _buildPagination(),
      ],
    );
  }

  Widget _buildPagination() {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: const BoxDecoration(
        border: Border(top: BorderSide(color: AppTheme.darkBorder)),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            'Halaman $_page dari ${_data!.totalPages}',
            style: const TextStyle(color: AppTheme.darkTextSecondary, fontSize: 13),
          ),
          Row(
            children: [
              IconButton(
                icon: const Icon(Icons.chevron_left),
                onPressed: _page > 1
                    ? () {
                        setState(() => _page--);
                        _loadData();
                      }
                    : null,
              ),
              IconButton(
                icon: const Icon(Icons.chevron_right),
                onPressed: _page < _data!.totalPages
                    ? () {
                        setState(() => _page++);
                        _loadData();
                      }
                    : null,
              ),
            ],
          ),
        ],
      ),
    );
  }

  void _showFilterSheet() {
    showModalBottomSheet(
      context: context,
      backgroundColor: AppTheme.darkSurface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (ctx) => _FilterSheet(
        type: _type,
        status: _status,
        sortBy: _sortBy,
        order: _order,
        onApply: (type, status, sortBy, order) {
          setState(() {
            _type = type;
            _status = status;
            _sortBy = sortBy;
            _order = order;
            _page = 1;
          });
          _loadData();
        },
      ),
    );
  }

  String _typeLabel(String type) {
    return switch (type) {
      'send' => 'Kirim',
      'received' => 'Terima',
      'buy' => 'Beli',
      'sell' => 'Jual',
      _ => type,
    };
  }
}

class _TransactionTile extends StatelessWidget {
  final Transaction tx;
  final String currentAddress;
  final VoidCallback onTap;

  const _TransactionTile({
    required this.tx,
    required this.currentAddress,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final isSend = tx.fromAddress.toLowerCase() == currentAddress.toLowerCase();
    final icon = isSend ? Icons.arrow_upward : Icons.arrow_downward;
    final color = isSend ? AppTheme.error : AppTheme.success;

    return ListTile(
      onTap: onTap,
      leading: Container(
        width: 40,
        height: 40,
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.15),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Icon(icon, color: color, size: 20),
      ),
      title: Text(
        tx.typeLabel,
        style: const TextStyle(fontWeight: FontWeight.w600),
      ),
      subtitle: Text(
        '${shortAddress(isSend ? tx.toAddress : tx.fromAddress)} • Fee: ${formatNumber(tx.fee)}',
        style: const TextStyle(fontSize: 12, fontFamily: 'monospace'),
      ),
      trailing: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          Text(
            '${isSend ? "-" : "+"}${formatYTE(tx.amount)}',
            style: TextStyle(
              fontWeight: FontWeight.bold,
              color: color,
              fontSize: 13,
            ),
          ),
          const SizedBox(height: 2),
          StatusBadge(status: tx.status),
        ],
      ),
    );
  }
}

/// Bottom sheet untuk filter transaksi.
class _FilterSheet extends StatefulWidget {
  final String type;
  final String status;
  final String sortBy;
  final String order;
  final void Function(String type, String status, String sortBy, String order)
      onApply;

  const _FilterSheet({
    required this.type,
    required this.status,
    required this.sortBy,
    required this.order,
    required this.onApply,
  });

  @override
  State<_FilterSheet> createState() => _FilterSheetState();
}

class _FilterSheetState extends State<_FilterSheet> {
  late String _type;
  late String _status;
  late String _sortBy;
  late String _order;

  @override
  void initState() {
    super.initState();
    _type = widget.type;
    _status = widget.status;
    _sortBy = widget.sortBy;
    _order = widget.order;
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text('Filter Transaksi',
              style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 20),

          // Type filter
          Text('Tipe', style: Theme.of(context).textTheme.bodySmall),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            children: [
              for (final t in ['all', 'send', 'received', 'buy', 'sell'])
                ChoiceChip(
                  label: Text(_typeLabel(t)),
                  selected: _type == t,
                  onSelected: (_) => setState(() => _type = t),
                ),
            ],
          ),
          const SizedBox(height: 16),

          // Status filter
          Text('Status', style: Theme.of(context).textTheme.bodySmall),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            children: [
              for (final s in ['all', 'pending', 'confirmed'])
                ChoiceChip(
                  label: Text(s == 'all' ? 'Semua' : s),
                  selected: _status == s,
                  onSelected: (_) => setState(() => _status = s),
                ),
            ],
          ),
          const SizedBox(height: 16),

          // Sort
          Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text('Urutkan',
                        style: Theme.of(context).textTheme.bodySmall),
                    const SizedBox(height: 8),
                    Wrap(
                      spacing: 8,
                      children: [
                        ChoiceChip(
                          label: const Text('ID'),
                          selected: _sortBy == 'id',
                          onSelected: (_) => setState(() => _sortBy = 'id'),
                        ),
                        ChoiceChip(
                          label: const Text('Jumlah'),
                          selected: _sortBy == 'amount',
                          onSelected: (_) =>
                              setState(() => _sortBy = 'amount'),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 16),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text('Arah', style: Theme.of(context).textTheme.bodySmall),
                  const SizedBox(height: 8),
                  Wrap(
                    spacing: 8,
                    children: [
                      ChoiceChip(
                        label: const Text('DESC'),
                        selected: _order == 'desc',
                        onSelected: (_) => setState(() => _order = 'desc'),
                      ),
                      ChoiceChip(
                        label: const Text('ASC'),
                        selected: _order == 'asc',
                        onSelected: (_) => setState(() => _order = 'asc'),
                      ),
                    ],
                  ),
                ],
              ),
            ],
          ),
          const SizedBox(height: 24),

          // Apply button
          ElevatedButton(
            onPressed: () {
              widget.onApply(_type, _status, _sortBy, _order);
              Navigator.pop(context);
            },
            child: const Text('Terapkan Filter'),
          ),
        ],
      ),
    );
  }

  String _typeLabel(String type) {
    return switch (type) {
      'all' => 'Semua',
      'send' => 'Kirim',
      'received' => 'Terima',
      'buy' => 'Beli',
      'sell' => 'Jual',
      _ => type,
    };
  }
}
