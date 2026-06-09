import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart' hide Block;
import '../../core/api/api_client.dart';
import '../../core/models/models.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/utils/formatters.dart';
import '../../shared/widgets/app_widgets.dart';

class BlocksScreen extends StatefulWidget {
  const BlocksScreen({super.key});

  @override
  State<BlocksScreen> createState() => _BlocksScreenState();
}

class _BlocksScreenState extends State<BlocksScreen> {
  final _api = ApiClient();
  List<Block>? _blocks;
  bool _isLoading = true;
  String? _error;
  int _offset = 0;
  final int _limit = 20;

  @override
  void initState() {
    super.initState();
    _loadBlocks();
  }

  Future<void> _loadBlocks() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final resp = await _api.getBlocks(limit: _limit, offset: _offset);
      setState(() {
        _blocks = (resp.data['data'] as List<dynamic>)
            .map((b) => Block.fromJson(b as Map<String, dynamic>))
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

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Block Explorer')),
      body: _isLoading
          ? const AppLoading()
          : _error != null
              ? AppError(message: _error!, onRetry: _loadBlocks)
              : _buildList(),
    );
  }

  Widget _buildList() {
    if (_blocks == null || _blocks!.isEmpty) {
      return const EmptyState(
        icon: Icons.block,
        title: 'Belum ada block',
      );
    }

    return Column(
      children: [
        Expanded(
          child: ListView.separated(
            padding: const EdgeInsets.all(16),
            itemCount: _blocks!.length,
            separatorBuilder: (_, __) => const Divider(height: 1),
            itemBuilder: (context, index) {
              final b = _blocks![index];
              return ListTile(
                leading: Container(
                  width: 48,
                  height: 48,
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
                      ),
                    ),
                  ),
                ),
                title: Text(
                  b.shortHash,
                  style: const TextStyle(fontFamily: 'monospace'),
                ),
                subtitle: Text(
                  '${b.transactions?.length ?? 0} txns • Diff: ${b.difficulty} • ${formatRelativeTime(b.timestamp)}',
                ),
                trailing: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: [
                    Text(
                      formatYTE(b.blockReward),
                      style: const TextStyle(fontWeight: FontWeight.w600),
                    ),
                    Text(
                      'Nonce: ${b.nonce}',
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  ],
                ),
                onTap: () => context.go('/blocks/${b.id}'),
              );
            },
          ),
        ),
        // Pagination
        Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              ElevatedButton(
                onPressed: _offset > 0
                    ? () {
                        setState(() => _offset -= _limit);
                        _loadBlocks();
                      }
                    : null,
                child: const Text('Previous'),
              ),
              Text('Offset: $_offset'),
              ElevatedButton(
                onPressed: _blocks!.length == _limit
                    ? () {
                        setState(() => _offset += _limit);
                        _loadBlocks();
                      }
                    : null,
                child: const Text('Next'),
              ),
            ],
          ),
        ),
      ],
    );
  }
}
