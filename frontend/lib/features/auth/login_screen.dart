import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../../core/api/api_client.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/widgets/app_widgets.dart';

/// Auth screen terpadu dengan tab Register dan Login.
///
/// **Register flow:**
/// 1. User isi username, address, public_key
/// 2. POST /register → dapat JWT cookie → redirect ke dashboard
///
/// **Login flow (challenge-response):**
/// 1. User masukkan address → POST /challenge/:address → dapat nonce
/// 2. User sign nonce dengan wallet → POST /challenge/verify → dapat JWT cookie
class AuthScreen extends StatefulWidget {
  const AuthScreen({super.key});

  @override
  State<AuthScreen> createState() => _AuthScreenState();
}

class _AuthScreenState extends State<AuthScreen>
    with SingleTickerProviderStateMixin {
  late final TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 440),
          child: Card(
            margin: const EdgeInsets.all(24),
            child: Padding(
              padding: const EdgeInsets.all(32),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  // Header
                  Text(
                    'YuteBlockchain',
                    style: Theme.of(context)
                        .textTheme
                        .headlineMedium
                        ?.copyWith(fontWeight: FontWeight.bold),
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: 24),

                  // Tab bar
                  Container(
                    decoration: BoxDecoration(
                      color: AppTheme.darkSurface,
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: TabBar(
                      controller: _tabController,
                      indicator: BoxDecoration(
                        color: AppTheme.primary,
                        borderRadius: BorderRadius.circular(8),
                      ),
                      labelColor: Colors.white,
                      unselectedLabelColor: AppTheme.darkTextSecondary,
                      tabs: const [
                        Tab(text: 'Register'),
                        Tab(text: 'Login'),
                      ],
                    ),
                  ),
                  const SizedBox(height: 24),

                  // Tab content
                  Flexible(
                    child: TabBarView(
                      controller: _tabController,
                      children: const [
                        _RegisterTab(),
                        _LoginTab(),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

// ==================== REGISTER TAB ====================

class _RegisterTab extends StatefulWidget {
  const _RegisterTab();

  @override
  State<_RegisterTab> createState() => _RegisterTabState();
}

class _RegisterTabState extends State<_RegisterTab> {
  final _formKey = GlobalKey<FormState>();
  final _usernameController = TextEditingController();
  final _addressController = TextEditingController();
  final _publicKeyController = TextEditingController();
  final _api = ApiClient();

  bool _isLoading = false;
  String? _error;

  @override
  void dispose() {
    _usernameController.dispose();
    _addressController.dispose();
    _publicKeyController.dispose();
    super.dispose();
  }

  Future<void> _register() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      await _api.register(
        username: _usernameController.text.trim(),
        address: _addressController.text.trim(),
        publicKey: _publicKeyController.text.trim(),
      );

      if (mounted) {
        context.go('/');
      }
    } on ApiException catch (e) {
      setState(() => _error = e.message);
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Form(
      key: _formKey,
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(
            'Buat akun baru',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: AppTheme.darkTextSecondary,
                ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 20),

          // Username
          TextFormField(
            controller: _usernameController,
            decoration: const InputDecoration(
              labelText: 'Username',
              hintText: 'alice',
              prefixIcon: Icon(Icons.person),
            ),
            validator: (v) {
              if (v == null || v.trim().isEmpty) return 'Username wajib diisi';
              if (v.trim().length < 3) return 'Minimal 3 karakter';
              if (v.trim().length > 50) return 'Maksimal 50 karakter';
              return null;
            },
          ),
          const SizedBox(height: 16),

          // Address
          TextFormField(
            controller: _addressController,
            decoration: const InputDecoration(
              labelText: 'Wallet Address',
              hintText: '0x1234...abcd',
              prefixIcon: Icon(Icons.account_balance_wallet),
            ),
            validator: (v) {
              if (v == null || v.trim().isEmpty) return 'Address wajib diisi';
              final addr = v.trim();
              if (!addr.startsWith('0x') || addr.length != 42) {
                return 'Format address tidak valid (0x + 40 hex)';
              }
              if (!RegExp(r'^0x[0-9a-fA-F]{40}$').hasMatch(addr)) {
                return 'Address harus hex string';
              }
              return null;
            },
          ),
          const SizedBox(height: 16),

          // Public Key
          TextFormField(
            controller: _publicKeyController,
            decoration: const InputDecoration(
              labelText: 'Public Key',
              hintText: 'abcdef1234... (hex tanpa 0x)',
              prefixIcon: Icon(Icons.vpn_key),
            ),
            maxLines: 2,
            validator: (v) {
              if (v == null || v.trim().isEmpty) return 'Public key wajib diisi';
              if (!RegExp(r'^[0-9a-fA-F]+$').hasMatch(v.trim())) {
                return 'Public key harus hex string';
              }
              return null;
            },
          ),
          const SizedBox(height: 24),

          // Register button
          ElevatedButton(
            onPressed: _isLoading ? null : _register,
            child: _isLoading
                ? const SizedBox(
                    height: 20,
                    width: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Text('Register'),
          ),

          // Error
          if (_error != null) ...[
            const SizedBox(height: 16),
            ErrorBanner(message: _error!),
          ],
        ],
      ),
    );
  }
}

// ==================== LOGIN TAB ====================

class _LoginTab extends StatefulWidget {
  const _LoginTab();

  @override
  State<_LoginTab> createState() => _LoginTabState();
}

class _LoginTabState extends State<_LoginTab> {
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
      final resp = await _api.challenge(_addressController.text.trim());
      final nonce = resp.data['data']['challenge'] as String;
      _nonceController.text = nonce;

      setState(() => _isChallengeMode = true);
    } on ApiException catch (e) {
      setState(() => _error = e.message);
    } finally {
      if (mounted) setState(() => _isLoading = false);
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
        address: _addressController.text.trim(),
        nonce: _nonceController.text,
        signature: _signatureController.text.trim(),
        username: _usernameController.text.trim(),
      );

      if (mounted) {
        context.go('/');
      }
    } on ApiException catch (e) {
      setState(() => _error = e.message);
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  void _reset() {
    setState(() {
      _isChallengeMode = false;
      _error = null;
      _nonceController.clear();
      _signatureController.clear();
    });
  }

  @override
  Widget build(BuildContext context) {
    return Form(
      key: _formKey,
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(
            _isChallengeMode
                ? 'Verifikasi signature Anda'
                : 'Masukkan address untuk login',
            style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                  color: AppTheme.darkTextSecondary,
                ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 20),

          // Address field
          TextFormField(
            controller: _addressController,
            readOnly: _isChallengeMode,
            decoration: const InputDecoration(
              labelText: 'Wallet Address',
              hintText: '0x...',
              prefixIcon: Icon(Icons.account_balance_wallet),
            ),
            validator: (v) {
              if (v == null || v.trim().isEmpty) return 'Address wajib diisi';
              if (!v.trim().startsWith('0x') || v.trim().length != 42) {
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
                if (v == null || v.trim().length < 3) {
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
                if (v == null || v.trim().isEmpty) {
                  return 'Signature wajib diisi';
                }
                return null;
              },
            ),
            const SizedBox(height: 24),

            // Verify button
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
            const SizedBox(height: 8),

            // Back button
            TextButton(
              onPressed: _isLoading ? null : _reset,
              child: const Text('Ganti address'),
            ),
          ] else ...[
            const SizedBox(height: 24),

            // Get challenge button
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

          // Error
          if (_error != null) ...[
            const SizedBox(height: 16),
            ErrorBanner(message: _error!),
          ],
        ],
      ),
    );
  }
}

