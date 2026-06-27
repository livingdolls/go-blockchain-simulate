package com.takahashi.yutecoin.ui.dashboard

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.takahashi.yutecoin.data.dto.NetworkResult
import com.takahashi.yutecoin.data.dto.TransactionItem
import com.takahashi.yutecoin.data.local.SessionManager
import com.takahashi.yutecoin.data.repository.WalletRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class TxHistoryUiState(
    val transactions: List<TransactionItem> = emptyList(),
    val isLoading: Boolean = false,
    val error: String? = null,
    val page: Int = 1,
    val totalPages: Int = 1
)

class TransactionHistoryViewModel(
    private val walletRepository: WalletRepository,
    private val sessionManager: SessionManager
) : ViewModel() {

    private val _state = MutableStateFlow(TxHistoryUiState())
    val state: StateFlow<TxHistoryUiState> = _state.asStateFlow()

    init {
        loadTransactions()
    }

    fun loadTransactions(page: Int = 1) {
        val address = sessionManager.getAddress()?.lowercase() ?: return
        _state.value = _state.value.copy(isLoading = true, error = null)

        viewModelScope.launch {
            when (val result = walletRepository.getTransactions(address, page, 20)) {
                is NetworkResult.Success -> {
                    _state.value = _state.value.copy(
                        isLoading = false,
                        transactions = result.data.transactions.transactions,
                        page = page,
                        totalPages = result.data.transactions.totalPages
                    )
                }
                is NetworkResult.Error -> {
                    _state.value = _state.value.copy(isLoading = false, error = result.message)
                }
                is NetworkResult.Loading -> {}
            }
        }
    }
}
