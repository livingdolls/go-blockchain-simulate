import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart' hide Block;
import '../../core/api/api_client.dart';
import '../../core/models/models.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/utils/formatters.dart';

/// Search delegate untuk Global Search.
///
/// Mencari di 3 kategori:
/// - Block by number: `GET /blocks/detail/:number`
/// - Block by hash: `GET /blocks/search?hash=...`
/// - Transaction by ID: `GET /transaction/:id`
///
/// Query bisa berupa:
/// - Angka murni → coba block number, lalu transaction ID
/// - Hex string (0x...) → coba block hash search
/// - Address → coba block by miner address
class BlockchainSearchDelegate extends SearchDelegate<String> {
  final ApiClient _api = ApiClient();

  BlockchainSearchDelegate() : super(searchFieldLabel: 'Cari block, transaksi, atau address...');

  @override
  ThemeData appBarTheme(BuildContext context) {
    return Theme.of(context).copyWith(
      appBarTheme: const AppBarTheme(
        backgroundColor: AppTheme.darkSurface,
        foregroundColor: AppTheme.darkText,
      ),
      inputDecorationTheme: const InputDecorationTheme(
        hintStyle: TextStyle(color: AppTheme.darkTextSecondary),
      ),
    );
  }

  @override
  List<Widget>? buildActions(BuildContext context) {
    return [
      if (query.isNotEmpty)
        IconButton(
          icon: const Icon(Icons.clear),
          onPressed: () => query = '',
        ),
    ];
  }

  @override
  Widget? buildLeading(BuildContext context) {
    return IconButton(
      icon: const Icon(Icons.arrow_back),
      onPressed: () => close(context, ''),
    );
  }

  @override
  Widget buildResults(BuildContext context) {
    if (query.trim().isEmpty) {
      return const Center(child: Text('Masukkan kata kunci untuk mencari'));
    }

    return FutureBuilder<List<SearchResult>>(
      future: _search(query.trim()),
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.waiting) {
          return const Center(child: CircularProgressIndicator());
        }

        if (snapshot.hasError) {
          return Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.error_outline, size: 48, color: AppTheme.error),
                const SizedBox(height: 16),
                Text('Error: ${snapshot.error}'),
              ],
            ),
          );
        }

        final results = snapshot.data ?? [];
        if (results.isEmpty) {
          return Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.search_off, size: 64, color: AppTheme.darkTextSecondary),
                const SizedBox(height: 16),
                Text(
                  'Tidak ada hasil untuk "$query"',
                  style: const TextStyle(color: AppTheme.darkTextSecondary),
                ),
              ],
            ),
          );
        }

        return ListView.separated(
          itemCount: results.length,
          separatorBuilder: (_, __) => const Divider(height: 1),
          itemBuilder: (context, index) {
            final result = results[index];
            return _SearchResultTile(
              result: result,
              onTap: () {
                close(context, result.route);
                context.go(result.route);
              },
            );
          },
        );
      },
    );
  }

  @override
  Widget buildSuggestions(BuildContext context) {
    if (query.trim().isEmpty) {
      return _buildSearchHints();
    }

    return FutureBuilder<List<SearchResult>>(
      future: _search(query.trim()),
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.waiting) {
          return const Center(child: CircularProgressIndicator());
        }

        final results = snapshot.data ?? [];
        if (results.isEmpty) {
          return const SizedBox.shrink();
        }

        return ListView.builder(
          itemCount: results.length,
          itemBuilder: (context, index) {
            final result = results[index];
            return _SearchResultTile(
              result: result,
              onTap: () {
                close(context, result.route);
                context.go(result.route);
              },
            );
          },
        );
      },
    );
  }

  Widget _buildSearchHints() {
    return Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Cari Berdasarkan',
            style: TextStyle(
              color: AppTheme.darkTextSecondary,
              fontSize: 12,
              fontWeight: FontWeight.w600,
              letterSpacing: 1.2,
            ),
          ),
          const SizedBox(height: 16),
          _hintTile(Icons.view_module, 'Block Number', 'contoh: 42'),
          _hintTile(Icons.tag, 'Block Hash', 'contoh: 0000abc...'),
          _hintTile(Icons.receipt_long, 'Transaction ID', 'contoh: 123'),
          _hintTile(Icons.account_balance_wallet, 'Address', 'contoh: 0x1234...'),
        ],
      ),
    );
  }

  Widget _hintTile(IconData icon, String title, String subtitle) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 8),
      child: Row(
        children: [
          Icon(icon, size: 20, color: AppTheme.primary),
          const SizedBox(width: 12),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(title, style: const TextStyle(fontWeight: FontWeight.w600)),
              Text(subtitle,
                  style: const TextStyle(
                      color: AppTheme.darkTextSecondary, fontSize: 12)),
            ],
          ),
        ],
      ),
    );
  }

  Future<List<SearchResult>> _search(String q) async {
    final results = <SearchResult>[];

    // 1. Coba parse sebagai angka (block number atau transaction ID)
    final number = int.tryParse(q);
    if (number != null && number > 0) {
      // Block by number
      try {
        final resp = await _api.getBlockByNumber(number);
        final block =
            Block.fromJson(resp.data['data'] as Map<String, dynamic>);
        results.add(SearchResult(
          type: SearchResultType.block,
          title: 'Block #${block.blockNumber}',
          subtitle:
              '${block.shortHash} • ${block.transactions?.length ?? 0} txns',
          route: '/blocks/${block.id}',
          icon: Icons.view_module,
        ));
      } on ApiException catch (_) {}

      // Transaction by ID
      try {
        final resp = await _api.getTransaction(number);
        final tx = Transaction.fromJson(
            resp.data['data']['transaction'] as Map<String, dynamic>);
        results.add(SearchResult(
          type: SearchResultType.transaction,
          title: 'Transaction #${tx.id}',
          subtitle:
              '${tx.typeLabel} • ${formatYTE(tx.amount)} • ${shortAddress(tx.fromAddress)}',
          route: '/transactions/${tx.id}',
          icon: Icons.receipt_long,
        ));
      } on ApiException catch (_) {}
    }

    // 2. Jika query dimulai dengan 0x (address/hash)
    if (q.startsWith('0x')) {
      // Block hash search
      try {
        final resp = await _api.searchBlocksByHash(q);
        final blocks = (resp.data['data'] as List<dynamic>)
            .map((b) => Block.fromJson(b as Map<String, dynamic>))
            .toList();
        for (final block in blocks.take(5)) {
          results.add(SearchResult(
            type: SearchResultType.block,
            title: 'Block #${block.blockNumber}',
            subtitle:
                '${block.shortHash} • ${block.transactions?.length ?? 0} txns',
            route: '/blocks/${block.id}',
            icon: Icons.view_module,
          ));
        }
      } on ApiException catch (_) {}

      // Address search (blocks by miner)
      try {
        final resp =
            await _api.searchByMiner(address: q, limit: 5);
        final blocks = (resp.data['data'] as List<dynamic>)
            .map((b) => Block.fromJson(b as Map<String, dynamic>))
            .toList();
        for (final block in blocks) {
          // Hindari duplikat jika sudah ada dari hash search
          if (!results.any((r) => r.route == '/blocks/${block.id}')) {
            results.add(SearchResult(
              type: SearchResultType.address,
              title: 'Block #${block.blockNumber} (miner)',
              subtitle:
                  '${block.shortHash} • ${formatRelativeTime(block.timestamp)}',
              route: '/blocks/${block.id}',
              icon: Icons.account_balance_wallet,
            ));
          }
        }
      } on ApiException catch (_) {}
    }

    return results;
  }
}

enum SearchResultType { block, transaction, address }

class SearchResult {
  final SearchResultType type;
  final String title;
  final String subtitle;
  final String route;
  final IconData icon;

  const SearchResult({
    required this.type,
    required this.title,
    required this.subtitle,
    required this.route,
    required this.icon,
  });
}

class _SearchResultTile extends StatelessWidget {
  final SearchResult result;
  final VoidCallback onTap;

  const _SearchResultTile({required this.result, required this.onTap});

  @override
  Widget build(BuildContext context) {
    final typeLabel = switch (result.type) {
      SearchResultType.block => 'Block',
      SearchResultType.transaction => 'Tx',
      SearchResultType.address => 'Address',
    };

    final typeColor = switch (result.type) {
      SearchResultType.block => AppTheme.primary,
      SearchResultType.transaction => AppTheme.accent,
      SearchResultType.address => AppTheme.warning,
    };

    return ListTile(
      leading: Container(
        width: 40,
        height: 40,
        decoration: BoxDecoration(
          color: typeColor.withValues(alpha: 0.15),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Icon(result.icon, color: typeColor, size: 20),
      ),
      title: Text(result.title, style: const TextStyle(fontWeight: FontWeight.w600)),
      subtitle: Text(result.subtitle, style: const TextStyle(fontSize: 12)),
      trailing: Container(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
        decoration: BoxDecoration(
          color: typeColor.withValues(alpha: 0.15),
          borderRadius: BorderRadius.circular(4),
        ),
        child: Text(
          typeLabel,
          style: TextStyle(color: typeColor, fontSize: 11, fontWeight: FontWeight.w600),
        ),
      ),
      onTap: onTap,
    );
  }
}
