/// Konfigurasi API dan environment aplikasi.
///
/// Untuk mengubah base URL, edit [_baseUrl] atau gunakan
/// environment variable saat build:
/// ```dart
/// flutter build web --dart-define=API_URL=https://api.example.com
/// ```
class AppConfig {
  static const String appName = 'Blockchain';
  static const String apiVersion = 'v1';

  // Base URL backend. Override via --dart-define=API_URL=...
  static const String _baseUrl = String.fromEnvironment(
    'API_URL',
    defaultValue: 'http://localhost:3010',
  );

  static String get baseUrl => _baseUrl;
  static String get apiBaseUrl => '$_baseUrl';

  // WebSocket URL (otomatis derive dari base URL)
  static String get wsUrl {
    final uri = Uri.parse(_baseUrl);
    final scheme = uri.scheme == 'https' ? 'wss' : 'ws';
    return '$scheme://${uri.host}:${uri.port}';
  }

  // SSE URL
  static String get sseBaseUrl => _baseUrl;

  // Timeouts
  static const Duration connectTimeout = Duration(seconds: 10);
  static const Duration receiveTimeout = Duration(seconds: 30);

  // Rate limit info (untuk UI)
  static const int txRateLimitPerMinute = 10;
  static const int authRateLimitPerMinute = 10;
}
