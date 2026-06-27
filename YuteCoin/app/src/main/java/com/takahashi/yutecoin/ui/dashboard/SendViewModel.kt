package com.takahashi.yutecoin.ui.dashboard

import android.util.Log
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.takahashi.yutecoin.crypto.WalletGenerator
import com.takahashi.yutecoin.data.dto.NetworkResult
import com.takahashi.yutecoin.data.dto.SendRequest
import com.takahashi.yutecoin.data.local.SessionManager
import com.takahashi.yutecoin.data.repository.SendRepository
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.util.Locale

data class SendUiState(
    val toAddress: String = "",
    val amount: String = "",
    val isSubmitting: Boolean = false,
    val error: String? = null,
    val successMessage: String? = null
)

class SendViewModel(
    private val sendRepository: SendRepository,
    private val sessionManager: SessionManager
) : ViewModel() {

    private val _state = MutableStateFlow(SendUiState())
    val state: StateFlow<SendUiState> = _state.asStateFlow()

    fun onToAddressChange(value: String) {
        _state.value = _state.value.copy(toAddress = value, error = null, successMessage = null)
    }

    fun onAmountChange(value: String) {
        _state.value = _state.value.copy(amount = value, error = null, successMessage = null)
    }

    fun send() {
        val toAddr = _state.value.toAddress.trim()
        val amountStr = _state.value.amount.trim()

        if (!toAddr.startsWith("0x") || toAddr.length != 42) {
            _state.value = _state.value.copy(error = "Invalid destination address (0x + 40 hex)")
            return
        }

        val amount = amountStr.toDoubleOrNull()
        if (amount == null || amount <= 0) {
            _state.value = _state.value.copy(error = "Enter a valid amount")
            return
        }

        val fromAddress = sessionManager.getAddress()?.lowercase()
        var privateKeyHex = sessionManager.getPrivateKeyHex()

        if (fromAddress == null) {
            _state.value = _state.value.copy(error = "Wallet not found. Please login again.")
            return
        }

        if (privateKeyHex == null) {
            val mnemonic = sessionManager.getMnemonic()
            if (mnemonic != null) {
                try {
                    val wallet = WalletGenerator.walletFromMnemonic(mnemonic)
                    privateKeyHex = wallet.privateKeyHex
                    sessionManager.savePrivateKey(privateKeyHex!!)
                } catch (e: Exception) {
                    Log.e("SendVM", "Failed to derive key from mnemonic", e)
                }
            }
        }

        if (privateKeyHex == null) {
            _state.value = _state.value.copy(error = "Wallet key not found. Please login again.")
            return
        }

        _state.value = _state.value.copy(isSubmitting = true, error = null)

        viewModelScope.launch {
            try {
                // Step 1: Generate nonce
                val nonceResult = sendRepository.generateNonce(fromAddress)
                if (nonceResult is NetworkResult.Error) {
                    _state.value = _state.value.copy(isSubmitting = false, error = nonceResult.message)
                    return@launch
                }
                val nonce = (nonceResult as NetworkResult.Success).data
                Log.d("SendVM", "Nonce: $nonce")

                // Step 2: Sign the message
                // Backend expects: "Send %.2f to %s nonce:%s"
                val message = "Send %.2f".format(Locale.US, amount) + " to $toAddr nonce:$nonce"
                Log.d("SendVM", "Message to sign: $message")

                val signature = withContext(Dispatchers.Default) {
                    WalletGenerator.signMessageToHex(message, privateKeyHex)
                }
                Log.d("SendVM", "Signature: ${signature.take(20)}...")

                // Step 3: Submit
                val request = SendRequest(
                    fromAddress = fromAddress,
                    toAddress = toAddr,
                    amount = amount,
                    nonce = nonce,
                    signature = signature
                )

                val sendResult = sendRepository.send(request)
                when (sendResult) {
                    is NetworkResult.Success -> {
                        _state.value = _state.value.copy(
                            isSubmitting = false,
                            successMessage = "Transaction submitted!",
                            toAddress = "",
                            amount = ""
                        )
                    }
                    is NetworkResult.Error -> {
                        _state.value = _state.value.copy(isSubmitting = false, error = sendResult.message)
                    }
                    is NetworkResult.Loading -> {}
                }
            } catch (e: Exception) {
                Log.e("SendVM", "Send error", e)
                _state.value = _state.value.copy(isSubmitting = false, error = e.message ?: "Unexpected error")
            }
        }
    }

    fun clearSuccess() {
        _state.value = _state.value.copy(successMessage = null)
    }
}
