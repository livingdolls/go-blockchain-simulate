package com.takahashi.yutecoin.ui.auth

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
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
    val keystorePassword: String = "",
    val isLoading: Boolean = false,
    val error: String? = null,
    val networkResult: String? = null,
    val isLoggedIn: Boolean = false
)

sealed class LoginEvent {
    object NavigateHome : LoginEvent()
    data class ShowError(val message: String) : LoginEvent()
}

class LoginViewModel(
    private val authRepository: AuthRepository,
    private val sessionManager: SessionManager
) : AndroidViewModel(Application()) {

    private val _state = MutableStateFlow(LoginUiState())
    val state: StateFlow<LoginUiState> = _state.asStateFlow()

    private val _events = Channel<LoginEvent>(Channel.BUFFERED)
    val events = _events.receiveAsFlow()

    fun onUsernameChange(value: String) {
        _state.value = _state.value.copy(username = value, error = null)
    }

    fun onMnemonicChange(value: String) {
        _state.value = _state.value.copy(mnemonicInput = value, error = null)
    }

    fun toggleMethod(useKeystore: Boolean) {
        _state.value = _state.value.copy(useKeystore = useKeystore, error = null)
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

        _state.value = _state.value.copy(isLoading = true, error = null)

        viewModelScope.launch {
            val result = if (_state.value.useKeystore) {
                NetworkResult.Error(400, "Keystore login not yet implemented")
            } else {
                val mnemonicWords = _state.value.mnemonicInput.trim().split("\\s+".toRegex())
                if (mnemonicWords.size != 12) {
                    NetworkResult.Error(400, "Mnemonic must be 12 words")
                } else {
                    authRepository.loginWithMnemonic(mnemonicWords, username)
                }
            }

            _state.value = _state.value.copy(isLoading = false)

            when (result) {
                is NetworkResult.Success -> {
                    val mnemonicWords = _state.value.mnemonicInput.trim().split("\\s+".toRegex())
                    sessionManager.saveMnemonic(mnemonicWords, username)
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

        _state.value = _state.value.copy(
            username = username,
            mnemonicInput = mnemonic.joinToString(" "),
            isLoading = true,
            error = null
        )

        viewModelScope.launch {
            val result = authRepository.loginWithMnemonic(mnemonic, username)
            _state.value = _state.value.copy(isLoading = false)

            when (result) {
                is NetworkResult.Success -> {
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

    fun clearError() {
        _state.value = _state.value.copy(error = null)
    }
}
