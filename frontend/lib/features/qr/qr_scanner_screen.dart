import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:mobile_scanner/mobile_scanner.dart';
import '../../core/theme/app_theme.dart';
import 'qr_data.dart';

/// Full-screen QR code scanner menggunakan mobile_scanner.
///
/// Mengembalikan [QrData] via `Navigator.pop(context, result)` jika
/// scan berhasil, atau `null` jika user cancel.
///
/// Platform support:
/// - Android: OK (ML Kit)
/// - iOS: OK (AVFoundation)
/// - Web: Tidak didukung (tombol scan disembunyikan di caller)
class QrScannerScreen extends StatefulWidget {
  const QrScannerScreen({super.key});

  @override
  State<QrScannerScreen> createState() => _QrScannerScreenState();
}

class _QrScannerScreenState extends State<QrScannerScreen> {
  MobileScannerController? _controller;
  bool _isProcessing = false;

  @override
  void initState() {
    super.initState();
    if (!kIsWeb) {
      _controller = MobileScannerController(
        detectionSpeed: DetectionSpeed.normal,
        facing: CameraFacing.back,
      );
    }
  }

  @override
  void dispose() {
    _controller?.dispose();
    super.dispose();
  }

  void _onDetect(BarcodeCapture capture) {
    if (_isProcessing) return;

    for (final barcode in capture.barcodes) {
      final raw = barcode.rawValue;
      if (raw == null || raw.isEmpty) continue;

      final qrData = QrData.parse(raw);
      if (qrData != null) {
        setState(() => _isProcessing = true);

        // Pause scanner sebelum pop
        _controller?.stop();

        // Delay kecil agar user melihat feedback visual
        Future.delayed(const Duration(milliseconds: 300), () {
          if (mounted) {
            Navigator.pop(context, qrData);
          }
        });
        return;
      }
    }

    // Semua barcode di frame ini tidak valid → tunggu frame berikutnya.
    // Jangan show error per-frame (terlalu noisy). Error hanya ditunjukkan
    // jika user menekan tombol "Input Manual".
  }

  void _showManualInputDialog() {
    final controller = TextEditingController();

    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Input Manual'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text(
              'Tempel string QR code (format: ytepay:0x...)',
              style: TextStyle(fontSize: 13),
            ),
            const SizedBox(height: 16),
            TextField(
              controller: controller,
              decoration: const InputDecoration(
                hintText: 'ytepay:0x...',
                border: OutlineInputBorder(),
              ),
              maxLines: 2,
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Batal'),
          ),
          ElevatedButton(
            onPressed: () {
              final raw = controller.text.trim();
              final qrData = QrData.parse(raw);
              if (qrData != null) {
                Navigator.pop(ctx);
                Navigator.pop(context, qrData);
              } else {
                ScaffoldMessenger.of(ctx).showSnackBar(
                  const SnackBar(
                    content: Text('Format tidak valid. Gunakan: ytepay:0x...'),
                  ),
                );
              }
            },
            child: const Text('Validasi'),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    // Web tidak support mobile_scanner camera
    if (kIsWeb) {
      return Scaffold(
        appBar: AppBar(title: const Text('Scan QR')),
        body: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.qr_code_scanner, size: 64, color: AppTheme.darkTextSecondary),
              const SizedBox(height: 16),
              const Text('QR scanner tidak tersedia di web'),
              const SizedBox(height: 16),
              ElevatedButton.icon(
                onPressed: _showManualInputDialog,
                icon: const Icon(Icons.edit),
                label: const Text('Input Manual'),
              ),
            ],
          ),
        ),
      );
    }

    return Scaffold(
      backgroundColor: Colors.black,
      appBar: AppBar(
        backgroundColor: Colors.black,
        foregroundColor: Colors.white,
        title: const Text('Scan QR Code'),
        actions: [
          IconButton(
            icon: const Icon(Icons.keyboard),
            onPressed: _showManualInputDialog,
            tooltip: 'Input manual',
          ),
        ],
      ),
      body: Stack(
        children: [
          // Camera preview
          if (_controller != null)
            MobileScanner(
              controller: _controller!,
              onDetect: _onDetect,
            ),

          // Overlay dengan frame scan
          _buildScanOverlay(),

          // Bottom instructions
          Positioned(
            bottom: 40,
            left: 0,
            right: 0,
            child: Column(
              children: [
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                  decoration: BoxDecoration(
                    color: Colors.black54,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: const Text(
                    'Arahkan kamera ke QR code YTEPay',
                    style: TextStyle(color: Colors.white, fontSize: 14),
                    textAlign: TextAlign.center,
                  ),
                ),
                const SizedBox(height: 16),
                TextButton.icon(
                  onPressed: _showManualInputDialog,
                  icon: const Icon(Icons.edit, color: Colors.white70),
                  label: const Text(
                    'Input manual',
                    style: TextStyle(color: Colors.white70),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildScanOverlay() {
    return Center(
      child: Container(
        width: 280,
        height: 280,
        decoration: BoxDecoration(
          border: Border.all(
            color: _isProcessing ? AppTheme.success : AppTheme.primary,
            width: 3,
          ),
          borderRadius: BorderRadius.circular(16),
        ),
        child: _isProcessing
            ? const Center(
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(Icons.check_circle, color: AppTheme.success, size: 48),
                    SizedBox(height: 8),
                    Text(
                      'QR Terdeteksi',
                      style: TextStyle(
                        color: AppTheme.success,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              )
            : null,
      ),
    );
  }
}
