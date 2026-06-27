package com.takahashi.yutecoin.ui.dashboard

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.takahashi.yutecoin.data.api.RetrofitClient
import com.takahashi.yutecoin.data.dto.NetworkResult
import com.takahashi.yutecoin.data.local.SessionManager
import com.takahashi.yutecoin.data.repository.BalanceRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class HomeUiState(
    val address: String = "",
    val name: String = "",
    val yteBalance: Double = 0.0,
    val usdBalance: Double = 0.0,
    val isLoading: Boolean = false,
    val error: String? = null
)

class HomeViewModel(
    application: Application,
    private val balanceRepository: BalanceRepository,
    private val sessionManager: SessionManager
) : AndroidViewModel(application) {

    private val _state = MutableStateFlow(HomeUiState())
    val state: StateFlow<HomeUiState> = _state.asStateFlow()

    init {
        loadBalance()
    }

    fun loadBalance() {
        val address = sessionManager.getAddress()?.lowercase()
        if (address == null) {
            _state.value = _state.value.copy(error = "Address not found")
            return
        }

        _state.value = _state.value.copy(isLoading = true, error = null)

        viewModelScope.launch {
            when (val result = balanceRepository.getBalance(address)) {
                is NetworkResult.Success -> {
                    _state.value = _state.value.copy(
                        isLoading = false,
                        address = result.data.address,
                        name = result.data.name,
                        yteBalance = result.data.yteBalance,
                        usdBalance = result.data.usdBalance,
                        error = null
                    )
                }
                is NetworkResult.Error -> {
                    _state.value = _state.value.copy(
                        isLoading = false,
                        error = result.message
                    )
                }
                is NetworkResult.Loading -> {}
            }
        }
    }

    fun logout() {
        sessionManager.clearAll()
        RetrofitClient.clearSession()
    }

    fun clearError() {
        _state.value = _state.value.copy(error = null)
    }
}
