package com.takahashi.yutecoin.ui.auth

import android.app.Application
import android.net.Uri
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.takahashi.yutecoin.crypto.KeystoreUtil
import com.takahashi.yutecoin.crypto.WalletGenerator
import com.takahashi.yutecoin.data.dto.NetworkResult
import com.takahashi.yutecoin.data.local.SessionManager
import com.takahashi.yutecoin.data.repository.AuthRepository
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.receiveAsFlow
import kotlinx.coroutines.launch

data class LoginUiState(
    val username: String = "",
    val mnemonicInput: String = "",
    val useKeystore: Boolean = false,
    val keystoreFileUri: Uri? = null,
    val keystoreFileName: String? = null,
    val keystorePassword: String = "",
    val isLoading: Boolean = false,
    val error: String? = null,
    val networkResult: String? = null,
    val isLoggedIn: Boolean = false,
    val hasSavedWallet: Boolean = false
)

sealed class LoginEvent {
    object NavigateHome : LoginEvent()
    data class ShowError(val message: String) : LoginEvent()
}

class LoginViewModel(
    application: Application,
    private val authRepository: AuthRepository,
    private val sessionManager: SessionManager
) : AndroidViewModel(application) {

    private val _state = MutableStateFlow(LoginUiState())
    val state: StateFlow<LoginUiState> = _state.asStateFlow()

    private val _events = Channel<LoginEvent>(Channel.BUFFERED)
    val events = _events.receiveAsFlow()

    init {
        checkSavedWallet()
    }

    private fun checkSavedWallet() {
        val mnemonic = sessionManager.getMnemonic()
        val username = sessionManager.getUsername()
        if (mnemonic != null && username != null) {
            _state.value = _state.value.copy(hasSavedWallet = true, username = username)
        }
    }

    fun onUsernameChange(value: String) {
        _state.value = _state.value.copy(username = value, error = null)
    }

    fun onMnemonicChange(value: String) {
        _state.value = _state.value.copy(mnemonicInput = value, error = null)
    }

    fun toggleMethod(useKeystore: Boolean) {
        _state.value = _state.value.copy(useKeystore = useKeystore, error = null)
    }

    fun onKeystoreFileSelected(uri: Uri, fileName: String) {
        _state.value = _state.value.copy(
            keystoreFileUri = uri,
            keystoreFileName = fileName,
            error = null
        )
    }

    fun onKeystorePasswordChange(value: String) {
        _state.value = _state.value.copy(keystorePassword = value, error = null)
    }

    fun login() {
        val username = _state.value.username.trim()
        if (username.isEmpty()) {
            _state.value = _state.value.copy(error = "Username is required")
            return
        }

        if (_state.value.useKeystore) {
            if (_state.value.keystoreFileUri == null) {
                _state.value = _state.value.copy(error = "Select a wallet JSON file first")
                return
            }
            if (_state.value.keystorePassword.isEmpty()) {
                _state.value = _state.value.copy(error = "Enter your keystore password")
                return
            }
        } else {
            val mnemonicWords = _state.value.mnemonicInput.trim().split("\\s+".toRegex())
            if (mnemonicWords.size != 12) {
                _state.value = _state.value.copy(error = "Mnemonic must be exactly 12 words")
                return
            }
            if (!WalletGenerator.isValidMnemonic(mnemonicWords)) {
                _state.value = _state.value.copy(error = "Invalid mnemonic phrase")
                return
            }
        }

        _state.value = _state.value.copy(isLoading = true, error = null)
        val isKeystore = _state.value.useKeystore
        val password = _state.value.keystorePassword
        val uri = _state.value.keystoreFileUri
        var keystoreContent: String? = null

        viewModelScope.launch(kotlinx.coroutines.Dispatchers.Default) {
            val result: NetworkResult<String> = if (isKeystore) {
                // Read file content first
                val content = kotlin.runCatching {
                    val app = getApplication<Application>()
                    app.contentResolver.openInputStream(uri!!)?.bufferedReader()?.use { it.readText() }
                        ?: throw IllegalStateException("Cannot read file")
                }.getOrElse { e ->
                    _state.value = _state.value.copy(isLoading = false, error = "Cannot read file: ${e.message}")
                    return@launch
                }
                keystoreContent = content
                authRepository.loginWithKeystoreContent(content, password, username)
            } else {
                val mnemonicWords = _state.value.mnemonicInput.trim().split("\\s+".toRegex())
                authRepository.loginWithMnemonic(mnemonicWords, username)
            }

            _state.value = _state.value.copy(isLoading = false)

            when (result) {
                is NetworkResult.Success<*> -> {
                    val address = result.data as String
                    if (!isKeystore) {
                        val mnemonicWords = _state.value.mnemonicInput.trim().split("\\s+".toRegex())
                        val wallet = WalletGenerator.walletFromMnemonic(mnemonicWords)
                        sessionManager.saveMnemonic(mnemonicWords, username)
                        sessionManager.savePrivateKey(wallet.privateKeyHex)
                    } else {
                        val content = keystoreContent ?: return@launch
                        val wallet = KeystoreUtil.decryptKeystore(content, password)
                        sessionManager.savePrivateKey(wallet.privateKeyHex)
                    }
                    sessionManager.saveAddress(address)
                    sessionManager.setLoggedIn(true)
                    _state.value = _state.value.copy(isLoggedIn = true)
                    _events.send(LoginEvent.NavigateHome)
                }
                is NetworkResult.Error -> {
                    _state.value = _state.value.copy(error = result.message)
                }
                is NetworkResult.Loading -> {}
            }
        }
    }

    fun loginWithSavedMnemonic() {
        val mnemonic = sessionManager.getMnemonic()
        val username = sessionManager.getUsername()

        if (mnemonic == null || username == null) {
            _state.value = _state.value.copy(error = "No saved wallet found")
            return
        }

        _state.value = _state.value.copy(isLoading = true, error = null)

        viewModelScope.launch(kotlinx.coroutines.Dispatchers.Default) {
            val result = authRepository.loginWithMnemonic(mnemonic, username)

            _state.value = _state.value.copy(isLoading = false)

            when (result) {
                is NetworkResult.Success -> {
                    val wallet = WalletGenerator.walletFromMnemonic(mnemonic)
                    sessionManager.savePrivateKey(wallet.privateKeyHex)
                    sessionManager.saveAddress(result.data)
                    sessionManager.setLoggedIn(true)
                    _state.value = _state.value.copy(isLoggedIn = true)
                    _events.send(LoginEvent.NavigateHome)
                }
                is NetworkResult.Error -> {
                    _state.value = _state.value.copy(
                        error = result.message,
                        username = username,
                        mnemonicInput = mnemonic.joinToString(" ")
                    )
                }
                is NetworkResult.Loading -> {}
            }
        }
    }

    fun clearError() {
        _state.value = _state.value.copy(error = null)
    }
}
