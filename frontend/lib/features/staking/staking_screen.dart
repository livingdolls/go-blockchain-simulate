import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../core/api/api_client.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/utils/formatters.dart';
import '../../shared/widgets/app_widgets.dart';

/// Staking screen — stake YTE untuk earn rewards.
///
/// Flow:
/// 1. User lihat staking info (APR, min amount, dll)
/// 2. User input amount + lock period
/// 3. POST /staking/stake -> sukses
/// 4. User bisa lihat status stakes + unstake yang sudah unlocked
class StakingScreen extends StatefulWidget {
  final String userAddress;
  const StakingScreen({super.key, required this.userAddress});

  @override
  State<StakingScreen> createState() => _StakingScreenState();
}

class _StakingScreenState extends State<StakingScreen> {
  final _api = ApiClient();
  final _amountController = TextEditingController();
  final _daysController = TextEditingController(text: '30');

  Map<String, dynamic>? _stakingInfo;
  Map<String, dynamic>? _stakingStatus;
  bool _isLoading = true;
  bool _isStaking = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  @override
  void dispose() {
    _amountController.dispose();
    _daysController.dispose();
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
        _api.getStakingInfo(),
        _api.getStakingStatus(widget.userAddress),
      ]);

      if (!mounted) return;
      setState(() {
        _stakingInfo = results[0].data['data'] as Map<String, dynamic>;
        _stakingStatus = results[1].data['data'] as Map<String, dynamic>;
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

  Future<void> _stake() async {
    final amount = double.tryParse(_amountController.text.trim());
    final days = int.tryParse(_daysController.text.trim());

    if (amount == null || amount <= 0) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Masukkan jumlah yang valid')),
      );
      return;
    }

    if (days == null || days < 1) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Lock period minimal 1 hari')),
      );
      return;
    }

    setState(() => _isStaking = true);

    try {
      await _api.stake(
        address: widget.userAddress,
        amount: amount,
        lockDays: days,
      );

      _amountController.clear();
      _loadData();

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Staking berhasil: ${formatYTE(amount)} selama $days hari'),
            backgroundColor: AppTheme.success,
          ),
        );
      }
    } on ApiException catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: ${e.message}'), backgroundColor: AppTheme.error),
        );
      }
    } finally {
      if (mounted) setState(() => _isStaking = false);
    }
  }

  Future<void> _unstake(int stakeId) async {
    try {
      await _api.unstake(address: widget.userAddress, stakeId: stakeId);
      _loadData();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Unstake berhasil'), backgroundColor: AppTheme.success),
        );
      }
    } on ApiException catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: ${e.message}'), backgroundColor: AppTheme.error),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Staking'),
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
          // Staking info card
          if (_stakingInfo != null) _buildInfoCard(),
          const SizedBox(height: 16),

          // Stake form
          _buildStakeForm(),
          const SizedBox(height: 24),

          // My stakes
          Text('Stakes Saya', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 8),
          if (_stakingStatus != null) ...[
            _buildStatusSummary(),
            const SizedBox(height: 8),
            _buildStakesList(),
          ],
        ],
      ),
    );
  }

  Widget _buildInfoCard() {
    final apr = (_stakingInfo!['staking_apr'] as num?)?.toDouble() ?? 0;
    final minAmount = (_stakingInfo!['min_stake_amount'] as num?)?.toDouble() ?? 0;
    final minLock = (_stakingInfo!['min_lock_duration_seconds'] as int?) ?? 0;
    final totalStaked = (_stakingInfo!['total_staked'] as num?)?.toDouble() ?? 0;

    return AppCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.lock, size: 20, color: AppTheme.primary),
              const SizedBox(width: 8),
              Text('Staking Info',
                  style: Theme.of(context).textTheme.titleMedium),
            ],
          ),
          const SizedBox(height: 16),
          _infoRow('APR', '${apr.toStringAsFixed(1)}%'),
          _infoRow('Min Stake', '${minAmount.toStringAsFixed(2)} YTE'),
          _infoRow('Min Lock', '${minLock ~/ 86400} hari'),
          _infoRow('Total Staked', formatYTE(totalStaked)),
        ],
      ),
    );
  }

  Widget _buildStakeForm() {
    return AppCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('Stake YTE', style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: 16),
          TextFormField(
            controller: _amountController,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            decoration: const InputDecoration(
              labelText: 'Jumlah YTE',
              hintText: '100.0',
              prefixIcon: Icon(Icons.token),
            ),
          ),
          const SizedBox(height: 12),
          TextFormField(
            controller: _daysController,
            keyboardType: TextInputType.number,
            decoration: const InputDecoration(
              labelText: 'Lock Period (hari)',
              hintText: '30',
              prefixIcon: Icon(Icons.lock_clock),
            ),
          ),
          const SizedBox(height: 20),
          SizedBox(
            width: double.infinity,
            child: ElevatedButton.icon(
              onPressed: _isStaking ? null : _stake,
              icon: _isStaking
                  ? const SizedBox(
                      height: 18,
                      width: 18,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.lock),
              label: const Text('Stake'),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatusSummary() {
    final totalStaked = (_stakingStatus!['total_staked'] as num?)?.toDouble() ?? 0;
    final totalRewards = (_stakingStatus!['total_rewards'] as num?)?.toDouble() ?? 0;
    final activeStakes = _stakingStatus!['active_stakes'] as int? ?? 0;

    return Row(
      children: [
        Expanded(
          child: StatCard(
            label: 'Total Staked',
            value: formatYTE(totalStaked),
            icon: Icons.lock,
            color: AppTheme.primary,
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: StatCard(
            label: 'Rewards',
            value: formatYTE(totalRewards),
            icon: Icons.token,
            color: AppTheme.success,
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: StatCard(
            label: 'Active',
            value: '$activeStakes',
            icon: Icons.check_circle,
          ),
        ),
      ],
    );
  }

  Widget _buildStakesList() {
    final records = (_stakingStatus!['records'] as List<dynamic>?)
            ?.map((r) => r as Map<String, dynamic>)
            .toList() ??
        [];

    if (records.isEmpty) {
      return const EmptyState(
        icon: Icons.lock_open,
        title: 'Belum ada stake',
        subtitle: 'Stake YTE untuk mulai earn rewards',
      );
    }

    return Column(
      children: records.map((record) => _StakeCard(
        record: record,
        onUnstake: () {
          final id = record['id'] as int;
          _unstake(id);
        },
      )).toList(),
    );
  }

  Widget _infoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: const TextStyle(color: AppTheme.darkTextSecondary)),
          Text(value, style: const TextStyle(fontWeight: FontWeight.w600)),
        ],
      ),
    );
  }
}

class _StakeCard extends StatelessWidget {
  final Map<String, dynamic> record;
  final VoidCallback onUnstake;

  const _StakeCard({required this.record, required this.onUnstake});

  @override
  Widget build(BuildContext context) {
    final amount = (record['amount'] as num?)?.toDouble() ?? 0;
    final rewardEarned = (record['reward_earned'] as num?)?.toDouble() ?? 0;
    final status = record['status'] as String? ?? '';
    final lockUntil = record['lock_until'] as int? ?? 0;
    final stakeId = record['id'] as int? ?? 0;

    final isLocked = lockUntil > DateTime.now().millisecondsSinceEpoch ~/ 1000;
    final lockDate = DateTime.fromMillisecondsSinceEpoch(lockUntil * 1000);

    final statusColor = switch (status) {
      'ACTIVE' => isLocked ? AppTheme.warning : AppTheme.success,
      'UNSTAKING' => AppTheme.warning,
      'WITHDRAWN' => AppTheme.darkTextSecondary,
      _ => AppTheme.darkTextSecondary,
    };

    return AppCard(
      margin: const EdgeInsets.only(bottom: 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                formatYTE(amount),
                style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(
                  color: statusColor.withValues(alpha: 0.15),
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Text(
                  isLocked ? 'LOCKED' : 'UNLOCKED',
                  style: TextStyle(
                    color: statusColor,
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          _stakeInfoRow('Reward', formatYTE(rewardEarned)),
          _stakeInfoRow('Lock until',
              '${lockDate.day}/${lockDate.month}/${lockDate.year} ${lockDate.hour}:${lockDate.minute.toString().padLeft(2, '0')}'),
          if (isLocked)
            _stakeInfoRow('Remaining',
                '${lockUntil - DateTime.now().millisecondsSinceEpoch ~/ 1000}s'),
          const SizedBox(height: 8),
          if (!isLocked && status == 'ACTIVE')
            SizedBox(
              width: double.infinity,
              child: ElevatedButton.icon(
                onPressed: onUnstake,
                icon: const Icon(Icons.lock_open),
                label: const Text('Unstake'),
                style: ElevatedButton.styleFrom(
                  backgroundColor: AppTheme.success,
                ),
              ),
            ),
        ],
      ),
    );
  }

  Widget _stakeInfoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label,
              style:
                  const TextStyle(color: AppTheme.darkTextSecondary, fontSize: 12)),
          Text(value, style: const TextStyle(fontSize: 12)),
        ],
      ),
    );
  }
}
