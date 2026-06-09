import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../../core/api/api_client.dart';
import '../../core/theme/app_theme.dart';

/// Login/Register screen dengan challenge-response authentication.
///
/// Flow:
/// 1. User masukkan address → POST /challenge/:address → dapat nonce
/// 2. User sign nonce dengan wallet → POST /challenge/verify → dapat JWT cookie
class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _formKey = GlobalKey<FormState>();
  final _addressController = TextEditingController();
  final _usernameController = TextEditingController();
  final _nonceController = TextEditingController();
  final _signatureController = TextEditingController();
  final _api = ApiClient();

  bool _isLoading = false;
  bool _isChallengeMode = false;
  String? _error;

  @override
  void dispose() {
    _addressController.dispose();
    _usernameController.dispose();
    _nonceController.dispose();
    _signatureController.dispose();
    super.dispose();
  }

  Future<void> _getChallenge() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final resp = await _api.challenge(_addressController.text);
      final nonce = resp.data['data']['challenge'] as String;
      _nonceController.text = nonce;

      setState(() => _isChallengeMode = true);
    } on ApiException catch (e) {
      setState(() => _error = e.message);
    } finally {
      setState(() => _isLoading = false);
    }
  }

  Future<void> _verify() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      await _api.verify(
        address: _addressController.text,
        nonce: _nonceController.text,
        signature: _signatureController.text,
        username: _usernameController.text,
      );

      if (mounted) {
        context.go('/');
      }
    } on ApiException catch (e) {
      setState(() => _error = e.message);
    } finally {
      setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 400),
          child: Card(
            margin: const EdgeInsets.all(24),
            child: Padding(
              padding: const EdgeInsets.all(32),
              child: Form(
                key: _formKey,
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    Text(
                      'YuteBlockchain',
                      style: Theme.of(context)
                          .textTheme
                          .headlineMedium
                          ?.copyWith(fontWeight: FontWeight.bold),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 8),
                    Text(
                      _isChallengeMode
                          ? 'Verifikasi signature Anda'
                          : 'Masukkan address untuk memulai',
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            color: AppTheme.darkTextSecondary,
                          ),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 32),

                    // Address field
                    TextFormField(
                      controller: _addressController,
                      decoration: const InputDecoration(
                        labelText: 'Wallet Address',
                        hintText: '0x...',
                        prefixIcon: Icon(Icons.account_balance_wallet),
                      ),
                      validator: (v) {
                        if (v == null || v.isEmpty) return 'Address wajib diisi';
                        if (!v.startsWith('0x') || v.length != 42) {
                          return 'Format address tidak valid';
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 16),

                    if (_isChallengeMode) ...[
                      // Username
                      TextFormField(
                        controller: _usernameController,
                        decoration: const InputDecoration(
                          labelText: 'Username',
                          prefixIcon: Icon(Icons.person),
                        ),
                        validator: (v) {
                          if (v == null || v.length < 3) {
                            return 'Username minimal 3 karakter';
                          }
                          return null;
                        },
                      ),
                      const SizedBox(height: 16),

                      // Nonce (read-only)
                      TextFormField(
                        controller: _nonceController,
                        readOnly: true,
                        decoration: const InputDecoration(
                          labelText: 'Nonce (challenge)',
                          prefixIcon: Icon(Icons.key),
                        ),
                      ),
                      const SizedBox(height: 16),

                      // Signature
                      TextFormField(
                        controller: _signatureController,
                        decoration: const InputDecoration(
                          labelText: 'Signature',
                          hintText: '0x...',
                          prefixIcon: Icon(Icons.edit),
                        ),
                        maxLines: 2,
                        validator: (v) {
                          if (v == null || v.isEmpty) {
                            return 'Signature wajib diisi';
                          }
                          return null;
                        },
                      ),
                      const SizedBox(height: 24),

                      ElevatedButton(
                        onPressed: _isLoading ? null : _verify,
                        child: _isLoading
                            ? const SizedBox(
                                height: 20,
                                width: 20,
                                child: CircularProgressIndicator(strokeWidth: 2),
                              )
                            : const Text('Verifikasi & Login'),
                      ),
                    ] else ...[
                      const SizedBox(height: 24),
                      ElevatedButton(
                        onPressed: _isLoading ? null : _getChallenge,
                        child: _isLoading
                            ? const SizedBox(
                                height: 20,
                                width: 20,
                                child: CircularProgressIndicator(strokeWidth: 2),
                              )
                            : const Text('Dapatkan Challenge'),
                      ),
                    ],

                    if (_error != null) ...[
                      const SizedBox(height: 16),
                      Container(
                        padding: const EdgeInsets.all(12),
                        decoration: BoxDecoration(
                          color: AppTheme.error.withValues(alpha: 0.1),
                          borderRadius: BorderRadius.circular(8),
                          border: Border.all(
                              color: AppTheme.error.withValues(alpha: 0.3)),
                        ),
                        child: Text(
                          _error!,
                          style: const TextStyle(color: AppTheme.error),
                          textAlign: TextAlign.center,
                        ),
                      ),
                    ],
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
