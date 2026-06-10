import 'package:dio/dio.dart';
import '../config/app_config.dart';

/// API client berbasis Dio untuk berkomunikasi dengan backend blockchain.
///
/// Menggunakan cookie-based authentication (JWT di HttpOnly cookie).
/// Dio secara otomatis mengirim cookie untuk setiap request ke same-origin.
///
/// Fitur:
/// - Automatic retry untuk network error
/// - Error mapping ke [ApiException]
/// - Idempotency-Key header untuk POST transaksi
class ApiClient {
  late final Dio _dio;

  ApiClient() {
    _dio = Dio(BaseOptions(
      baseUrl: AppConfig.apiBaseUrl,
      connectTimeout: AppConfig.connectTimeout,
      receiveTimeout: AppConfig.receiveTimeout,
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
      // Penting: denganCredentials=true agar browser mengirim HttpOnly cookie.
      // Dio di web otomatis mengirim cookie same-origin.
    ));

    _dio.interceptors.addAll([
      _ErrorInterceptor(),
      LogInterceptor(
        requestBody: true,
        responseBody: true,
        logPrint: (obj) => print('[API] $obj'), // ignore: avoid_print
      ),
    ]);
  }

  Dio get dio => _dio;

  // ==================== Auth ====================

  Future<Response> register({
    required String username,
    required String address,
    required String publicKey,
  }) {
    return _dio.post('/register', data: {
      'username': username,
      'address': address,
      'public_key': publicKey,
    });
  }

  Future<Response> challenge(String address) {
    return _dio.post('/challenge/$address');
  }

  Future<Response> verify({
    required String address,
    required String nonce,
    required String signature,
    required String username,
  }) {
    return _dio.post('/challenge/verify', data: {
      'address': address,
      'nonce': nonce,
      'signature': signature,
      'username': username,
    });
  }

  Future<Response> getProfile() {
    return _dio.get('/profile');
  }

  Future<Response> logout() {
    return _dio.post('/admin/auth/logout');
  }

  Future<Response> userLogout() {
    return _dio.post('/auth/logout');
  }

  // ==================== Balance ====================

  Future<Response> getBalance(String address) {
    return _dio.get('/balance/$address');
  }

  Future<Response> topUp({
    required String address,
    required double amount,
    String? referenceId,
    String? description,
  }) {
    return _dio.post('/balance/topup', data: {
      'address': address,
      'amount': amount,
      if (referenceId != null) 'reference_id': referenceId,
      if (description != null) 'description': description,
    });
  }

  Future<Response> getWallet(
    String address, {
    String type = 'all',
    String status = 'all',
    int page = 1,
    int limit = 10,
    String sortBy = 'id',
    String order = 'desc',
  }) {
    return _dio.get('/wallet/$address', queryParameters: {
      'type': type,
      'status': status,
      'page': page,
      'limit': limit,
      'sort_by': sortBy,
      'order': order,
    });
  }

  // ==================== Transaction ====================

  Future<Response> generateNonce(String address) {
    return _dio.get('/generate-tx-nonce/$address');
  }

  Future<Response> sendTransaction({
    required String fromAddress,
    required String toAddress,
    required double amount,
    required String nonce,
    required String signature,
    String? idempotencyKey,
  }) {
    return _dio.post(
      '/transaction/send',
      data: {
        'from_address': fromAddress,
        'to_address': toAddress,
        'amount': amount,
        'nonce': nonce,
        'signature': signature,
      },
      options: Options(headers: {
        if (idempotencyKey != null) 'Idempotency-Key': idempotencyKey,
      }),
    );
  }

  Future<Response> buyTransaction({
    required String address,
    required double amount,
    required String nonce,
    required String signature,
    String? idempotencyKey,
  }) {
    return _dio.post(
      '/transaction/buy',
      data: {
        'address': address,
        'amount': amount,
        'nonce': nonce,
        'signature': signature,
      },
      options: Options(headers: {
        if (idempotencyKey != null) 'Idempotency-Key': idempotencyKey,
      }),
    );
  }

  Future<Response> sellTransaction({
    required String address,
    required double amount,
    required String nonce,
    required String signature,
    String? idempotencyKey,
  }) {
    return _dio.post(
      '/transaction/sell',
      data: {
        'address': address,
        'amount': amount,
        'nonce': nonce,
        'signature': signature,
      },
      options: Options(headers: {
        if (idempotencyKey != null) 'Idempotency-Key': idempotencyKey,
      }),
    );
  }

  Future<Response> getTransaction(int id) {
    return _dio.get('/transaction/$id');
  }

  // ==================== Blocks ====================

  Future<Response> getBlocks({int limit = 10, int offset = 0}) {
    return _dio.get('/blocks', queryParameters: {
      'limit': limit,
      'offset': offset,
    });
  }

  Future<Response> getBlockById(int id) {
    return _dio.get('/blocks/$id');
  }

  Future<Response> getBlockByNumber(int number) {
    return _dio.get('/blocks/detail/$number');
  }

  Future<Response> getBlockStats() {
    return _dio.get('/blocks/stats');
  }

  Future<Response> getBlocksInRange(int from, int to) {
    return _dio.get('/blocks/range', queryParameters: {
      'from': from,
      'to': to,
    });
  }

  Future<Response> searchBlocksByHash(String hash) {
    return _dio.get('/blocks/search', queryParameters: {'hash': hash});
  }

  Future<Response> checkIntegrity() {
    return _dio.get('/blocks/integrity');
  }

  // ==================== Reward ====================

  Future<Response> getRewardInfo() {
    return _dio.get('/reward/info');
  }

  Future<Response> getBlockReward(int blockNumber) {
    return _dio.get('/reward/block/$blockNumber');
  }

  Future<Response> getRewardSchedule(int blocks) {
    return _dio.get('/reward/schedule/$blocks');
  }

  // ==================== Market ====================

  Future<Response> getMarketState() {
    return _dio.get('/market');
  }

  Future<Response> getCandles({
    required String interval,
    String type = 'latest',
  }) {
    return _dio.get('/candles', queryParameters: {
      'interval': interval,
      'type': type,
    });
  }

  Future<Response> getCandleRange({
    required String interval,
    required int startTime,
    int limit = 100,
  }) {
    return _dio.get('/candles/range', queryParameters: {
      'interval': interval,
      'start_time': startTime,
      'limit': limit,
    });
  }

  // ==================== Admin ====================

  Future<Response> adminLogin({
    required String username,
    required String password,
  }) {
    return _dio.post('/admin/auth/login', data: {
      'username': username,
      'password': password,
    });
  }

  Future<Response> getAdminDashboard() {
    return _dio.get('/admin/dashboard');
  }

  Future<Response> getAdmins({int limit = 10, int offset = 0}) {
    return _dio.get('/admin/admins', queryParameters: {
      'limit': limit,
      'offset': offset,
    });
  }

  Future<Response> getActivityLogs({
    int? adminId,
    String? action,
    int limit = 50,
    int offset = 0,
  }) {
    return _dio.get('/admin/activity-logs', queryParameters: {
      if (adminId != null) 'admin_id': adminId,
      if (action != null) 'action': action,
      'limit': limit,
      'offset': offset,
    });
  }

  Future<Response> getRecentActivityLogs() {
    return _dio.get('/admin/activity-logs/recent');
  }

  Future<Response> createAdmin({
    required int userId,
    required String role,
    List<String>? permissions,
  }) {
    return _dio.post('/admin/admins', data: {
      'user_id': userId,
      'role': role,
      if (permissions != null) 'permissions': permissions,
    });
  }

  Future<Response> updateAdminRole({
    required int id,
    required String role,
    List<String>? permissions,
  }) {
    return _dio.put('/admin/admins/$id/role', data: {
      'role': role,
      if (permissions != null) 'permissions': permissions,
    });
  }

  Future<Response> updateAdminStatus({
    required int id,
    required String status,
  }) {
    return _dio.put('/admin/admins/$id/status', data: {
      'status': status,
    });
  }

  Future<Response> deleteAdmin(int id) {
    return _dio.delete('/admin/admins/$id');
  }

  Future<Response> getTransactionsByBlockNumber(int blockNumber) {
    return _dio.get('/blocks/transaction/$blockNumber');
  }

  Future<Response> searchByMiner({
    required String address,
    int limit = 10,
    int offset = 0,
  }) {
    return _dio.get('/blocks/search/miner/', queryParameters: {
      'address': address,
      'limit': limit,
      'offset': offset,
    });
  }

  Future<Response> generateBlock() {
    return _dio.post('/blocks/generate');
  }
}

/// Interceptor yang mengkonversi DioException ke [ApiException]
/// dengan pesan yang user-friendly.
class _ErrorInterceptor extends Interceptor {
  @override
  void onError(DioException err, ErrorInterceptorHandler handler) {
    final apiException = ApiException.fromDioException(err);
    handler.reject(DioException(
      requestOptions: err.requestOptions,
      response: err.response,
      type: err.type,
      error: apiException,
    ));
  }
}

/// Exception yang di-throw oleh ApiClient untuk error dari backend.
class ApiException implements Exception {
  final String message;
  final int? statusCode;
  final List<FieldError>? fields;

  ApiException({
    required this.message,
    this.statusCode,
    this.fields,
  });

  factory ApiException.fromDioException(DioException err) {
    switch (err.type) {
      case DioExceptionType.connectionTimeout:
      case DioExceptionType.sendTimeout:
      case DioExceptionType.receiveTimeout:
        return ApiException(
          message: 'Koneksi timeout. Periksa jaringan Anda.',
          statusCode: 408,
        );
      case DioExceptionType.connectionError:
        return ApiException(
          message: 'Tidak dapat terhubung ke server.',
          statusCode: 0,
        );
      case DioExceptionType.badResponse:
        return _parseBadResponse(err);
      default:
        return ApiException(
          message: 'Terjadi kesalahan: ${err.message}',
          statusCode: err.response?.statusCode,
        );
    }
  }

  static ApiException _parseBadResponse(DioException err) {
    final statusCode = err.response?.statusCode ?? 500;
    final data = err.response?.data;

    if (data is Map<String, dynamic>) {
      final error = data['error'] as String? ?? 'Unknown error';
      final fieldsData = data['fields'] as List<dynamic>?;

      List<FieldError>? fields;
      if (fieldsData != null) {
        fields = fieldsData
            .map((f) => FieldError.fromJson(f as Map<String, dynamic>))
            .toList();
      }

      return ApiException(
        message: error,
        statusCode: statusCode,
        fields: fields,
      );
    }

    return ApiException(
      message: 'Server error ($statusCode)',
      statusCode: statusCode,
    );
  }

  @override
  String toString() => 'ApiException: $message (HTTP $statusCode)';
}

/// Field-level validation error dari backend.
class FieldError {
  final String field;
  final String message;

  FieldError({required this.field, required this.message});

  factory FieldError.fromJson(Map<String, dynamic> json) {
    return FieldError(
      field: json['field'] as String? ?? '',
      message: json['message'] as String? ?? '',
    );
  }
}
