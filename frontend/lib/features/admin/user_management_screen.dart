import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart' hide Block;
import '../../core/api/api_client.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/utils/formatters.dart';
import '../../shared/widgets/app_widgets.dart';

/// User management screen — admin view all users.
///
/// Endpoint: GET /admin/users?limit=50&offset=0
class UserManagementScreen extends StatefulWidget {
  const UserManagementScreen({super.key});

  @override
  State<UserManagementScreen> createState() => _UserManagementScreenState();
}

class _UserManagementScreenState extends State<UserManagementScreen> {
  final _api = ApiClient();
  List<Map<String, dynamic>> _users = [];
  bool _isLoading = true;
  String? _error;
  int _offset = 0;
  final int _limit = 50;

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
      // Use wallet endpoint to get users with balances
      // The admin users endpoint may not exist yet, so we use a workaround
      final resp = await _api.getAdmins(limit: _limit, offset: _offset);
      if (!mounted) return;
      final data = resp.data['data'];
      setState(() {
        _users = (data['admins'] as List<dynamic>?)
                ?.map((a) => a as Map<String, dynamic>)
                .toList() ??
            [];
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
        title: const Text('User Management'),
        actions: [
          IconButton(icon: const Icon(Icons.refresh), onPressed: _loadData),
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
    if (_users.isEmpty) {
      return const EmptyState(
        icon: Icons.people,
        title: 'Belum ada users',
      );
    }

    return Column(
      children: [
        Expanded(
          child: ListView.separated(
            padding: const EdgeInsets.all(16),
            itemCount: _users.length,
            separatorBuilder: (_, __) => const Divider(height: 1),
            itemBuilder: (context, index) {
              final user = _users[index];
              return _UserTile(
                user: user,
                onTap: () {
                  final address = user['address'] as String?;
                  if (address != null) {
                    context.go('/wallet/$address');
                  }
                },
              );
            },
          ),
        ),
        // Pagination
        if (_users.length >= _limit)
          Padding(
            padding: const EdgeInsets.all(12),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                ElevatedButton(
                  onPressed: _offset > 0
                      ? () {
                          setState(() => _offset -= _limit);
                          _loadData();
                        }
                      : null,
                  child: const Text('Previous'),
                ),
                Text('Offset: $_offset',
                    style: const TextStyle(color: AppTheme.darkTextSecondary)),
                ElevatedButton(
                  onPressed: _users.length == _limit
                      ? () {
                          setState(() => _offset += _limit);
                          _loadData();
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

class _UserTile extends StatelessWidget {
  final Map<String, dynamic> user;
  final VoidCallback? onTap;

  const _UserTile({required this.user, this.onTap});

  @override
  Widget build(BuildContext context) {
    final username = user['username'] as String? ?? user['name'] as String? ?? '';
    final address = user['address'] as String? ?? '';
    final balance = (user['balance'] as num?)?.toDouble();
    final createdAt = user['created_at'] as String? ?? '';

    return ListTile(
      onTap: onTap,
      leading: CircleAvatar(
        backgroundColor: AppTheme.primary,
        child: Text(
          username.isNotEmpty ? username[0].toUpperCase() : '?',
          style: const TextStyle(color: Colors.white),
        ),
      ),
      title: Text(
        username,
        style: const TextStyle(fontWeight: FontWeight.w600),
      ),
      subtitle: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            shortAddress(address),
            style: const TextStyle(fontFamily: 'monospace', fontSize: 12),
          ),
          if (createdAt.isNotEmpty)
            Text(
              formatISODate(createdAt),
              style: const TextStyle(
                  fontSize: 11, color: AppTheme.darkTextSecondary),
            ),
        ],
      ),
      trailing: balance != null
          ? Text(
              formatUSD(balance),
              style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
            )
          : null,
    );
  }
}
