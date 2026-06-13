package com.takahashi.yutecoin.ui.auth

import android.app.Application
import android.util.Log
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
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
import kotlin.math.log

data class RegisterUiState(
    val currentStep: Int = 0,        // 0=Password, 1=Mnemonic, 2=Username
    val password: String = "",
    val confirmPassword: String = "",
    val mnemonic: List<String> = emptyList(),
    val privateKeyHex: String = "",
    val publicKeyHex: String = "",
    val address: String = "",
    val username: String = "",
    val hasBackedUp: Boolean = false,
    val isGenerating: Boolean = false,
    val isLoading: Boolean = false,
    val error: String? = null,
    val passwordError: String? = null,
    val usernameError: String? = null,
    val isSuccess: Boolean = false
)

sealed class RegisterEvent {
    object NavigateHome : RegisterEvent()
    data class ShowError(val message: String) : RegisterEvent()
}

class RegisterViewModel(
    private val authRepository: AuthRepository,
    private val sessionManager: SessionManager
) : AndroidViewModel(Application()) {

    private val _state = MutableStateFlow(RegisterUiState())
    val state: StateFlow<RegisterUiState> = _state.asStateFlow()

    private val _events = Channel<RegisterEvent>(Channel.BUFFERED)
    val events = _events.receiveAsFlow()

    // --- Step 0: Password ---

    fun onPasswordChange(value: String) {
        _state.value = _state.value.copy(password = value, passwordError = null)
    }

    fun onConfirmPasswordChange(value: String) {
        _state.value = _state.value.copy(confirmPassword = value, passwordError = null)
    }

    fun goToMnemonicStep() {
        val pwd = _state.value.password
        val confirm = _state.value.confirmPassword

        if (pwd.length < 8) {
            _state.value = _state.value.copy(passwordError = "Password must be at least 8 characters")
            return
        }
        if (pwd != confirm) {
            _state.value = _state.value.copy(passwordError = "Passwords do not match")
            return
        }

        _state.value = _state.value.copy(currentStep = 1, passwordError = null)
    }

    // --- Step 1: Mnemonic ---

    fun generateWallet() {
        _state.value = _state.value.copy(isGenerating = true)
        viewModelScope.launch(kotlinx.coroutines.Dispatchers.Default) {
            try {
                val wallet = WalletGenerator.generateWallet()
                Log.d("yute", "Debug Message $wallet")
                _state.value = _state.value.copy(
                    mnemonic = wallet.mnemonic,
                    privateKeyHex = wallet.privateKeyHex,
                    publicKeyHex = wallet.publicKeyHex,
                    address = wallet.address,
                    isGenerating = false,
                    hasBackedUp = false
                )
            } catch (e: Exception) {
                Log.d("yute", "debug error $e")
                _state.value = _state.value.copy(
                    isGenerating = false,
                    error = "Failed to generate wallet: ${e.message}"
                )
            }
        }
    }

    fun onBackupConfirmed() {
        _state.value = _state.value.copy(hasBackedUp = true)
    }

    fun goToUsernameStep() {
        if (!_state.value.hasBackedUp) return
        _state.value = _state.value.copy(currentStep = 2)
    }

    // --- Step 2: Username ---

    fun onUsernameChange(value: String) {
        _state.value = _state.value.copy(username = value, usernameError = null)
    }

    fun submitRegistration() {
        val username = _state.value.username.trim()
        if (username.length < 3) {
            _state.value = _state.value.copy(usernameError = "Username must be at least 3 characters")
            return
        }

        _state.value = _state.value.copy(isLoading = true, error = null)

        viewModelScope.launch {
            val result = authRepository.register(
                username = username,
                address = _state.value.address,
                publicKey = _state.value.publicKeyHex
            )

            _state.value = _state.value.copy(isLoading = false)

            when (result) {
                is NetworkResult.Success -> {
                    sessionManager.saveWallet(
                        mnemonic = _state.value.mnemonic,
                        privateKeyHex = _state.value.privateKeyHex,
                        address = _state.value.address
                    )
                    sessionManager.saveUsername(username)
                    sessionManager.setLoggedIn(true)
                    _state.value = _state.value.copy(isSuccess = true)
                    _events.send(RegisterEvent.NavigateHome)
                }
                is NetworkResult.Error -> {
                    _state.value = _state.value.copy(error = result.message)
                }
                is NetworkResult.Loading -> {}
            }
        }
    }

    // --- Navigation ---

    fun goBack() {
        if (_state.value.currentStep > 0) {
            _state.value = _state.value.copy(currentStep = _state.value.currentStep - 1)
        }
    }

    fun clearError() {
        _state.value = _state.value.copy(error = null, passwordError = null, usernameError = null)
    }
}
