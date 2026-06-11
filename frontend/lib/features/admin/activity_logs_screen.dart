import 'package:flutter/material.dart';
import '../../core/api/api_client.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/utils/formatters.dart';
import '../../shared/widgets/app_widgets.dart';

/// Activity logs screen — admin activity history.
///
/// Endpoint: GET /admin/activity-logs?limit=50&offset=0
class ActivityLogsScreen extends StatefulWidget {
  const ActivityLogsScreen({super.key});

  @override
  State<ActivityLogsScreen> createState() => _ActivityLogsScreenState();
}

class _ActivityLogsScreenState extends State<ActivityLogsScreen> {
  final _api = ApiClient();
  List<Map<String, dynamic>> _logs = [];
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
      final resp = await _api.getActivityLogs(limit: 100);
      if (!mounted) return;
      final data = resp.data['data'];
      setState(() {
        _logs = (data['logs'] as List<dynamic>?)
                ?.map((l) => l as Map<String, dynamic>)
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
        title: const Text('Activity Logs'),
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
    if (_logs.isEmpty) {
      return const EmptyState(
        icon: Icons.history,
        title: 'Belum ada activity logs',
      );
    }

    return ListView.separated(
      padding: const EdgeInsets.all(16),
      itemCount: _logs.length,
      separatorBuilder: (_, __) => const Divider(height: 1),
      itemBuilder: (context, index) {
        final log = _logs[index];
        return _LogTile(log: log);
      },
    );
  }
}

class _LogTile extends StatelessWidget {
  final Map<String, dynamic> log;
  const _LogTile({required this.log});

  @override
  Widget build(BuildContext context) {
    final action = log['action'] as String? ?? '';
    final username = log['username'] as String? ?? '';
    final status = log['status'] as String? ?? '';
    final timestamp = log['created_at'] as String? ?? '';
    final details = log['changes_summary'] as String? ?? '';

    final (icon, color) = switch (action.toLowerCase()) {
      'login' => (Icons.login, AppTheme.success),
      'logout' => (Icons.logout, AppTheme.darkTextSecondary),
      'create_admin' => (Icons.person_add, AppTheme.primary),
      'update_admin_role' => (Icons.edit, AppTheme.warning),
      'update_admin_status' => (Icons.toggle_on, AppTheme.warning),
      'delete_admin' => (Icons.delete, AppTheme.error),
      _ => (Icons.info, AppTheme.darkTextSecondary),
    };

    return ListTile(
      leading: Container(
        width: 36,
        height: 36,
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.15),
          borderRadius: BorderRadius.circular(6),
        ),
        child: Icon(icon, color: color, size: 18),
      ),
      title: Text(
        '$username — $action',
        style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
      ),
      subtitle: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (details.isNotEmpty)
            Text(details, style: const TextStyle(fontSize: 12)),
          Text(
            formatISODate(timestamp),
            style: const TextStyle(fontSize: 11, color: AppTheme.darkTextSecondary),
          ),
        ],
      ),
      trailing: Container(
        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
        decoration: BoxDecoration(
          color: (status == 'success' ? AppTheme.success : AppTheme.error)
              .withValues(alpha: 0.15),
          borderRadius: BorderRadius.circular(4),
        ),
        child: Text(
          status,
          style: TextStyle(
            color: status == 'success' ? AppTheme.success : AppTheme.error,
            fontSize: 11,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
    );
  }
}
