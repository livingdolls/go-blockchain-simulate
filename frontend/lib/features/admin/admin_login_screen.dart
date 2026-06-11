import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../../core/api/api_client.dart';
import '../../core/theme/app_theme.dart';
import '../../shared/widgets/app_widgets.dart';

/// Admin login screen dengan username/password.
///
/// Flow:
/// 1. Admin masukkan username + password
/// 2. POST /admin/auth/login → dapat admin_token cookie
/// 3. Redirect ke /admin/dashboard
class AdminLoginScreen extends StatefulWidget {
  const AdminLoginScreen({super.key});

  @override
  State<AdminLoginScreen> createState() => _AdminLoginScreenState();
}

class _AdminLoginScreenState extends State<AdminLoginScreen> {
  final _formKey = GlobalKey<FormState>();
  final _usernameController = TextEditingController();
  final _passwordController = TextEditingController();
  final _api = ApiClient();

  bool _isLoading = false;
  String? _error;

  @override
  void dispose() {
    _usernameController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  Future<void> _login() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      await _api.adminLogin(
        username: _usernameController.text.trim(),
        password: _passwordController.text,
      );

      if (mounted) context.go('/admin/dashboard');
    } on ApiException catch (e) {
      setState(() => _error = e.message);
    } catch (e) {
      setState(() => _error = 'Terjadi kesalahan: $e');
    } finally {
      if (mounted) setState(() => _isLoading = false);
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
                      'Admin Login',
                      style: Theme.of(context)
                          .textTheme
                          .headlineMedium
                          ?.copyWith(fontWeight: FontWeight.bold),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 8),
                    Text(
                      'Masukkan kredensial admin untuk melanjutkan',
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                            color: AppTheme.darkTextSecondary,
                          ),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 32),

                    // Username
                    TextFormField(
                      controller: _usernameController,
                      decoration: const InputDecoration(
                        labelText: 'Username',
                        prefixIcon: Icon(Icons.person),
                      ),
                      validator: (v) {
                        if (v == null || v.trim().isEmpty) {
                          return 'Username wajib diisi';
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 16),

                    // Password
                    TextFormField(
                      controller: _passwordController,
                      obscureText: true,
                      decoration: const InputDecoration(
                        labelText: 'Password',
                        prefixIcon: Icon(Icons.lock),
                      ),
                      validator: (v) {
                        if (v == null || v.isEmpty) {
                          return 'Password wajib diisi';
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: 24),

                    // Login button
                    ElevatedButton.icon(
                      onPressed: _isLoading ? null : _login,
                      icon: _isLoading
                          ? const SizedBox(
                              height: 18,
                              width: 18,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            )
                          : const Icon(Icons.login),
                      label: const Text('Login'),
                    ),

                    // Error
                    if (_error != null) ...[
                      const SizedBox(height: 16),
                      ErrorBanner(message: _error!),
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
