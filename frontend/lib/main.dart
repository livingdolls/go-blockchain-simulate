import 'package:flutter/material.dart';
import 'core/config/router.dart';
import 'core/theme/app_theme.dart';

void main() {
  runApp(const BlockchainApp());
}

/// Aplikasi utama blockchain frontend.
///
/// Menggunakan:
/// - GoRouter untuk navigasi
/// - Dark theme sebagai default (cocok untuk dashboard crypto)
/// - Material 3 design system
class BlockchainApp extends StatelessWidget {
  const BlockchainApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'YuteBlockchain',
      theme: AppTheme.darkTheme,
      darkTheme: AppTheme.darkTheme,
      themeMode: ThemeMode.dark,
      routerConfig: appRouter,
      debugShowCheckedModeBanner: false,
    );
  }
}
