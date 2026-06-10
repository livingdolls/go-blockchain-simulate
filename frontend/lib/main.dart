import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'core/config/router.dart';
import 'core/theme/app_theme.dart';
import 'features/notifications/notification_provider.dart';

void main() {
  runApp(
    ChangeNotifierProvider(
      create: (_) => NotificationProvider()..loadFromStorage(),
      child: const BlockchainApp(),
    ),
  );
}

/// Aplikasi utama blockchain frontend.
///
/// Menggunakan:
/// - GoRouter untuk navigasi
/// - Dark theme sebagai default (cocok untuk dashboard crypto)
/// - Material 3 design system
/// - Provider untuk notification state
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
