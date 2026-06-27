package com.takahashi.yutecoin.ui.dashboard

import android.util.Log
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.takahashi.yutecoin.crypto.WalletGenerator
import com.takahashi.yutecoin.data.dto.BuySellRequest
import com.takahashi.yutecoin.data.dto.NetworkResult
import com.takahashi.yutecoin.data.local.SessionManager
import com.takahashi.yutecoin.data.repository.BalanceRepository
import com.takahashi.yutecoin.data.repository.SendRepository
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.util.Locale
import kotlin.math.max

data class BuySellUiState(
    val amount: String = "",
    val yteBalance: Double = 0.0,
    val usdBalance: Double = 0.0,
    val isSubmitting: Boolean = false,
    val isLoadingBalance: Boolean = false,
    val error: String? = null,
    val successMessage: String? = null
)

class BuySellViewModel(
    private val sendRepository: SendRepository,
    private val balanceRepository: BalanceRepository,
    private val sessionManager: SessionManager
) : ViewModel() {

    private val _state = MutableStateFlow(BuySellUiState())
    val state: StateFlow<BuySellUiState> = _state.asStateFlow()

    init {
        loadBalances()
    }

    fun loadBalances() {
        val address = sessionManager.getAddress()?.lowercase() ?: return
        _state.value = _state.value.copy(isLoadingBalance = true)
        viewModelScope.launch {
            when (val result = balanceRepository.getBalance(address)) {
                is NetworkResult.Success -> {
                    _state.value = _state.value.copy(
                        isLoadingBalance = false,
                        yteBalance = result.data.yteBalance,
                        usdBalance = result.data.usdBalance
                    )
                }
                is NetworkResult.Error -> _state.value = _state.value.copy(isLoadingBalance = false)
                is NetworkResult.Loading -> {}
            }
        }
    }

    fun onAmountChange(value: String) {
        _state.value = _state.value.copy(amount = value, error = null, successMessage = null)
    }

    fun buy() {
        executeTransaction("BUY")
    }

    fun sell() {
        executeTransaction("SELL")
    }

    private fun executeTransaction(type: String) {
        val amountStr = _state.value.amount.trim()
        val amount = amountStr.toDoubleOrNull()

        if (amount == null || amount <= 0) {
            _state.value = _state.value.copy(error = "Enter a valid amount")
            return
        }

        val address = sessionManager.getAddress()?.lowercase()
        var privateKeyHex = sessionManager.getPrivateKeyHex()

        if (address == null) {
            _state.value = _state.value.copy(error = "Wallet not found")
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
                    Log.e("BuySellVM", "Key derivation failed", e)
                }
            }
        }

        if (privateKeyHex == null) {
            _state.value = _state.value.copy(error = "Wallet key not found. Please login again.")
            return
        }

        val fee = max(amount * 0.001, 0.001)
        val totalRequired = amount + fee

        if (type == "BUY") {
            val usd = _state.value.usdBalance
            if (usd < totalRequired) {
                _state.value = _state.value.copy(
                    error = "Insufficient USD (have %.2f, need %.4f + %.4f fee)".format(Locale.US, usd, amount, fee)
                )
                return
            }
        } else {
            val yte = _state.value.yteBalance
            if (yte < totalRequired) {
                _state.value = _state.value.copy(
                    error = "Insufficient YTE (have %.4f, need %.4f + %.4f fee)".format(Locale.US, yte, amount, fee)
                )
                return
            }
        }

        _state.value = _state.value.copy(isSubmitting = true, error = null)

        viewModelScope.launch {
            try {
                val nonceResult = sendRepository.generateNonce(address)
                if (nonceResult is NetworkResult.Error) {
                    _state.value = _state.value.copy(isSubmitting = false, error = nonceResult.message)
                    return@launch
                }
                val nonce = (nonceResult as NetworkResult.Success).data

                val message = " $type %.2f nonce:$nonce".format(Locale.US, amount)
                Log.d("BuySellVM", "Sign msg: $message")

                val signature = withContext(Dispatchers.Default) {
                    WalletGenerator.signMessageToHex(message, privateKeyHex)
                }

                val request = BuySellRequest(address = address, amount = amount, nonce = nonce, signature = signature)
                val result = if (type == "BUY") sendRepository.buy(request) else sendRepository.sell(request)

                when (result) {
                    is NetworkResult.Success -> {
                        _state.value = _state.value.copy(
                            isSubmitting = false,
                            successMessage = "$type submitted!",
                            amount = ""
                        )
                    }
                    is NetworkResult.Error -> {
                        _state.value = _state.value.copy(isSubmitting = false, error = result.message)
                    }
                    is NetworkResult.Loading -> {}
                }
            } catch (e: Exception) {
                Log.e("BuySellVM", "Error", e)
                _state.value = _state.value.copy(isSubmitting = false, error = e.message ?: "Unexpected error")
            }
        }
    }

    fun clearSuccess() {
        _state.value = _state.value.copy(successMessage = null)
    }
}
