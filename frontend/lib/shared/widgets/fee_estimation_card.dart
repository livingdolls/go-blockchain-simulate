import 'package:flutter/material.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/utils/formatters.dart';

/// Shared widget untuk menampilkan estimasi fee.
/// Dipakai di send, buy, sell screens untuk konsistensi.
class FeeEstimationCard extends StatelessWidget {
  final Map<String, dynamic> feeData;
  final String priority;

  const FeeEstimationCard({
    super.key,
    required this.feeData,
    required this.priority,
  });

  @override
  Widget build(BuildContext context) {
    final baseFee = (feeData['base_fee'] as num?)?.toDouble() ?? 0;
    final estimatedFee = (feeData['estimated_fee'] as num?)?.toDouble() ?? 0;
    final congestionLevel = feeData['congestion_level'] as String? ?? 'low';
    final pendingCount = feeData['pending_count'] as int? ?? 0;
    final congestionMult = (feeData['congestion_multiplier'] as num?)?.toDouble() ?? 1;
    final priorityMult = (feeData['priority_multiplier'] as num?)?.toDouble() ?? 1;

    final congestionColor = switch (congestionLevel) {
      'low' => AppTheme.success,
      'medium' => AppTheme.warning,
      'high' => AppTheme.error,
      'very_high' => AppTheme.error,
      _ => AppTheme.darkTextSecondary,
    };

    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: AppTheme.darkSurface,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: AppTheme.darkBorder),
      ),
      child: Column(
        children: [
          _feeRow('Base Fee', formatYTE(baseFee)),
          _feeRow('Congestion', '$congestionMult x ($pendingCount pending)'),
          _feeRow('Priority', '$priorityMult x ($priority)'),
          const Divider(height: 16),
          _feeRow('Estimasi Fee', formatYTE(estimatedFee), isBold: true),
          const SizedBox(height: 8),
          Row(
            children: [
              Expanded(
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(4),
                  child: LinearProgressIndicator(
                    value: ((feeData['congestion_percent'] as num?)?.toDouble() ?? 0) / 100,
                    backgroundColor: AppTheme.darkCard,
                    valueColor: AlwaysStoppedAnimation<Color>(congestionColor),
                    minHeight: 6,
                  ),
                ),
              ),
              const SizedBox(width: 8),
              Text(
                congestionLevel.toUpperCase(),
                style: TextStyle(
                  color: congestionColor,
                  fontSize: 11,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _feeRow(String label, String value, {bool isBold = false}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label,
              style: TextStyle(
                  color: AppTheme.darkTextSecondary,
                  fontSize: 12,
                  fontWeight: isBold ? FontWeight.bold : FontWeight.normal)),
          Text(value,
              style: TextStyle(
                  fontSize: 12,
                  fontWeight: isBold ? FontWeight.bold : FontWeight.w600)),
        ],
      ),
    );
  }
}
