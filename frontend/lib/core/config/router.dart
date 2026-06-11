import 'package:go_router/go_router.dart';
import '../../features/auth/login_screen.dart';
import '../../features/dashboard/dashboard_screen.dart';
import '../../features/blocks/blocks_screen.dart';
import '../../features/blocks/block_detail_screen.dart';
import '../../features/wallet/wallet_screen.dart';
import '../../features/transactions/transaction_detail_screen.dart';
import '../../features/transactions/transaction_history_screen.dart';
import '../../features/transactions/send_screen.dart';
import '../../features/transactions/buy_screen.dart';
import '../../features/transactions/sell_screen.dart';
import '../../features/notifications/notification_screen.dart';
import '../../features/qr/receive_screen.dart';
import '../../features/market/market_screen.dart';
import '../../features/explorer/rich_list_screen.dart';
import '../../features/explorer/mempool_screen.dart';

/// Router konfigurasi aplikasi.
///
/// Struktur:
/// - `/` → Dashboard (home)
/// - `/login` → Login/Register
/// - `/blocks` → Block explorer
/// - `/blocks/:id` → Block detail
/// - `/wallet/:address` → Wallet detail
/// - `/transactions/:id` → Transaction detail
/// - `/transactions/send` → Send YTE
/// - `/transactions/buy` → Buy YTE
/// - `/transactions/sell` → Sell YTE
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
      path: '/wallet/:address/history',
      builder: (context, state) {
        final address = state.pathParameters['address']!;
        return TransactionHistoryScreen(address: address);
      },
    ),
    GoRoute(
      path: '/receive',
      builder: (context, state) {
        final address = state.uri.queryParameters['address'] ?? '';
        return ReceiveScreen(address: address);
      },
    ),
    GoRoute(
      path: '/transactions/send',
      builder: (context, state) => const SendScreen(),
    ),
    GoRoute(
      path: '/transactions/buy',
      builder: (context, state) => const BuyScreen(),
    ),
    GoRoute(
      path: '/transactions/sell',
      builder: (context, state) => const SellScreen(),
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
    GoRoute(
      path: '/notifications',
      builder: (context, state) => const NotificationScreen(),
    ),
    GoRoute(
      path: '/explorer/richlist',
      builder: (context, state) => const RichListScreen(),
    ),
    GoRoute(
      path: '/explorer/mempool',
      builder: (context, state) => const MempoolScreen(),
    ),
  ],
);
