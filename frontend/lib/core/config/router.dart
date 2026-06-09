import 'package:go_router/go_router.dart';
import '../../features/auth/login_screen.dart';
import '../../features/dashboard/dashboard_screen.dart';
import '../../features/blocks/blocks_screen.dart';
import '../../features/blocks/block_detail_screen.dart';
import '../../features/wallet/wallet_screen.dart';
import '../../features/transactions/transaction_detail_screen.dart';
import '../../features/market/market_screen.dart';

/// Router konfigurasi aplikasi.
///
/// Struktur:
/// - `/` → Dashboard (home)
/// - `/login` → Login/Register
/// - `/blocks` → Block explorer
/// - `/blocks/:id` → Block detail
/// - `/wallet/:address` → Wallet detail
/// - `/transactions/:id` → Transaction detail
/// - `/market` → Market data
final appRouter = GoRouter(
  initialLocation: '/',
  routes: [
    GoRoute(
      path: '/',
      builder: (context, state) => const DashboardScreen(),
    ),
    GoRoute(
      path: '/login',
      builder: (context, state) => const AuthScreen(),
    ),
    GoRoute(
      path: '/blocks',
      builder: (context, state) => const BlocksScreen(),
    ),
    GoRoute(
      path: '/blocks/:id',
      builder: (context, state) {
        final id = int.parse(state.pathParameters['id']!);
        return BlockDetailScreen(blockId: id);
      },
    ),
    GoRoute(
      path: '/wallet/:address',
      builder: (context, state) {
        final address = state.pathParameters['address']!;
        return WalletScreen(address: address);
      },
    ),
    GoRoute(
      path: '/transactions/:id',
      builder: (context, state) {
        final id = int.parse(state.pathParameters['id']!);
        return TransactionDetailScreen(transactionId: id);
      },
    ),
    GoRoute(
      path: '/market',
      builder: (context, state) => const MarketScreen(),
    ),
  ],
);
