import 'package:flutter/material.dart';
import 'package:qr_flutter/qr_flutter.dart';
import '../../core/theme/app_theme.dart';

/// Widget reusable untuk menampilkan QR code.
/// Dioptimasi untuk dark theme dengan foreground putih.
class QrDisplayWidget extends StatelessWidget {
  final String data;
  final double size;

  const QrDisplayWidget({
    super.key,
    required this.data,
    this.size = 250,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
      ),
      child: QrImageView(
        data: data,
        version: QrVersions.auto,
        size: size,
        backgroundColor: Colors.white,
        eyeStyle: const QrEyeStyle(
          eyeShape: QrEyeShape.square,
          color: AppTheme.darkBg,
        ),
        dataModuleStyle: const QrDataModuleStyle(
          dataModuleShape: QrDataModuleShape.square,
          color: AppTheme.darkBg,
        ),
      ),
    );
  }
}
