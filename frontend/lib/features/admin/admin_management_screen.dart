import 'package:flutter/material.dart';
import '../../core/api/api_client.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/widgets/app_widgets.dart';

/// Admin management screen — list, create, update role/status admins.
///
/// Endpoints:
/// - GET /admin/admins?limit=10&offset=0
/// - POST /admin/admins
/// - PUT /admin/admins/:id/role
/// - PUT /admin/admins/:id/status
/// - DELETE /admin/admins/:id
class AdminManagementScreen extends StatefulWidget {
  const AdminManagementScreen({super.key});

  @override
  State<AdminManagementScreen> createState() => _AdminManagementScreenState();
}

class _AdminManagementScreenState extends State<AdminManagementScreen> {
  final _api = ApiClient();
  List<Map<String, dynamic>> _admins = [];
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
      final resp = await _api.getAdmins(limit: 100);
      if (!mounted) return;
      final data = resp.data['data'];
      setState(() {
        _admins = (data['admins'] as List<dynamic>?)
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
        title: const Text('Admin Management'),
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
    if (_admins.isEmpty) {
      return const EmptyState(
        icon: Icons.admin_panel_settings,
        title: 'Belum ada admin',
      );
    }

    return ListView.separated(
      padding: const EdgeInsets.all(16),
      itemCount: _admins.length,
      separatorBuilder: (_, __) => const Divider(height: 1),
      itemBuilder: (context, index) {
        final admin = _admins[index];
        return _AdminTile(
          admin: admin,
          onRoleChanged: (newRole) => _updateRole(admin['id'] as int, newRole),
          onStatusChanged: (newStatus) =>
              _updateStatus(admin['id'] as int, newStatus),
          onDelete: () => _deleteAdmin(admin['id'] as int, admin['username'] as String),
        );
      },
    );
  }

  Future<void> _updateRole(int id, String role) async {
    try {
      await _api.updateAdminRole(id: id, role: role);
      _loadData();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Role updated to $role')),
        );
      }
    } on ApiException catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: ${e.message}')),
        );
      }
    }
  }

  Future<void> _updateStatus(int id, String status) async {
    try {
      await _api.updateAdminStatus(id: id, status: status);
      _loadData();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Status updated to $status')),
        );
      }
    } on ApiException catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: ${e.message}')),
        );
      }
    }
  }

  Future<void> _deleteAdmin(int id, String username) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Hapus Admin'),
        content: Text('Hapus admin "$username"?'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Batal')),
          ElevatedButton(
            onPressed: () => Navigator.pop(ctx, true),
            style: ElevatedButton.styleFrom(backgroundColor: AppTheme.error),
            child: const Text('Hapus'),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    try {
      await _api.deleteAdmin(id);
      _loadData();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Admin "$username" dihapus')),
        );
      }
    } on ApiException catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: ${e.message}')),
        );
      }
    }
  }
}

class _AdminTile extends StatelessWidget {
  final Map<String, dynamic> admin;
  final void Function(String) onRoleChanged;
  final void Function(String) onStatusChanged;
  final VoidCallback onDelete;

  const _AdminTile({
    required this.admin,
    required this.onRoleChanged,
    required this.onStatusChanged,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    final username = admin['username'] as String? ?? '';
    final role = admin['role'] as String? ?? '';
    final status = admin['status'] as String? ?? '';
    final isActive = status == 'active';

    return ListTile(
      leading: CircleAvatar(
        backgroundColor: isActive ? AppTheme.success : AppTheme.error,
        child: Text(
          username.isNotEmpty ? username[0].toUpperCase() : '?',
          style: const TextStyle(color: Colors.white),
        ),
      ),
      title: Text(username, style: const TextStyle(fontWeight: FontWeight.w600)),
      subtitle: Text('Role: $role • Status: $status'),
      trailing: PopupMenuButton<String>(
        onSelected: (value) {
          switch (value) {
            case 'role_admin':
              onRoleChanged('admin');
              break;
            case 'role_moderator':
              onRoleChanged('moderator');
              break;
            case 'role_support':
              onRoleChanged('support');
              break;
            case 'status_active':
              onStatusChanged('active');
              break;
            case 'status_inactive':
              onStatusChanged('inactive');
              break;
            case 'status_suspended':
              onStatusChanged('suspended');
              break;
            case 'delete':
              onDelete();
              break;
          }
        },
        itemBuilder: (_) => [
          const PopupMenuItem(
            enabled: false,
            child: Text('Change Role', style: TextStyle(fontWeight: FontWeight.bold)),
          ),
          const PopupMenuItem(value: 'role_admin', child: Text('  admin')),
          const PopupMenuItem(value: 'role_moderator', child: Text('  moderator')),
          const PopupMenuItem(value: 'role_support', child: Text('  support')),
          const PopupMenuDivider(),
          const PopupMenuItem(
            enabled: false,
            child: Text('Change Status', style: TextStyle(fontWeight: FontWeight.bold)),
          ),
          const PopupMenuItem(value: 'status_active', child: Text('  active')),
          const PopupMenuItem(value: 'status_inactive', child: Text('  inactive')),
          const PopupMenuItem(value: 'status_suspended', child: Text('  suspended')),
          const PopupMenuDivider(),
          const PopupMenuItem(
            value: 'delete',
            child: Text('Delete', style: TextStyle(color: AppTheme.error)),
          ),
        ],
      ),
    );
  }
}
