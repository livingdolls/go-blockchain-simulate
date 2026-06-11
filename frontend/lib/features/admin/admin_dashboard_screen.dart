import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../../core/api/api_client.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/widgets/app_widgets.dart';

/// Admin dashboard dengan statistik dan navigasi ke management screens.
///
/// Endpoint: GET /admin/dashboard
class AdminDashboardScreen extends StatefulWidget {
  const AdminDashboardScreen({super.key});

  @override
  State<AdminDashboardScreen> createState() => _AdminDashboardScreenState();
}

class _AdminDashboardScreenState extends State<AdminDashboardScreen> {
  final _api = ApiClient();
  Map<String, dynamic>? _stats;
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
      final resp = await _api.getAdminDashboard();
      if (!mounted) return;
      setState(() {
        _stats = resp.data['data'] as Map<String, dynamic>?;
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
        title: const Text('Admin Dashboard'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadData,
          ),
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: () async {
              await _api.logout();
              if (mounted) context.go('/admin/login');
            },
          ),
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
        // Stats grid
        if (_stats != null) _buildStatsGrid(),
        const SizedBox(height: 24),

        // Quick actions
        Text('Management', style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: 12),
        _buildActionGrid(),
      ],
    );
  }

  Widget _buildStatsGrid() {
    final totalUsers = _stats!['total_users'] as int? ?? 0;
    final totalTransactions = _stats!['total_transactions'] as int? ?? 0;
    final totalBlocks = _stats!['total_blocks'] as int? ?? 0;
    final activeUsers = _stats!['active_users'] as int? ?? 0;

    return GridView.count(
      crossAxisCount: 2,
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      mainAxisSpacing: 12,
      crossAxisSpacing: 12,
      childAspectRatio: 1.6,
      children: [
        StatCard(
          label: 'Total Users',
          value: '$totalUsers',
          icon: Icons.people,
        ),
        StatCard(
          label: 'Active Users',
          value: '$activeUsers',
          icon: Icons.person,
          color: AppTheme.success,
        ),
        StatCard(
          label: 'Total Blocks',
          value: '$totalBlocks',
          icon: Icons.view_module,
        ),
        StatCard(
          label: 'Transactions',
          value: '$totalTransactions',
          icon: Icons.receipt_long,
        ),
      ],
    );
  }

  Widget _buildActionGrid() {
    return GridView.count(
      crossAxisCount: 2,
      shrinkWrap: true,
      physics: const NeverScrollableScrollPhysics(),
      mainAxisSpacing: 12,
      crossAxisSpacing: 12,
      childAspectRatio: 2,
      children: [
        _actionCard(
          icon: Icons.admin_panel_settings,
          label: 'Admin Management',
          onTap: () => context.go('/admin/admins'),
        ),
        _actionCard(
          icon: Icons.people,
          label: 'User Management',
          onTap: () => context.go('/admin/users'),
        ),
        _actionCard(
          icon: Icons.history,
          label: 'Activity Logs',
          onTap: () => context.go('/admin/activity-logs'),
        ),
        _actionCard(
          icon: Icons.monetization_on,
          label: 'Force Mine',
          onTap: _forceMine,
        ),
      ],
    );
  }

  Widget _actionCard({
    required IconData icon,
    required String label,
    required VoidCallback onTap,
  }) {
    return Material(
      color: AppTheme.darkCard,
      borderRadius: BorderRadius.circular(12),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(icon, color: AppTheme.primary, size: 28),
              const SizedBox(height: 8),
              Text(label,
                  style: const TextStyle(fontSize: 13),
                  textAlign: TextAlign.center),
            ],
          ),
        ),
      ),
    );
  }

  Future<void> _forceMine() async {
    try {
      await _api.generateBlock();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Block generation triggered')),
        );
        _loadData();
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
