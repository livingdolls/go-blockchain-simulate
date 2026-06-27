package com.takahashi.yutecoin.ui.dashboard

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.takahashi.yutecoin.data.api.RetrofitClient
import com.takahashi.yutecoin.data.dto.CandleResponse
import com.takahashi.yutecoin.data.dto.NetworkResult
import com.takahashi.yutecoin.data.local.SessionManager
import com.takahashi.yutecoin.data.repository.BalanceRepository
import com.takahashi.yutecoin.data.repository.MarketRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class HomeUiState(
    val address: String = "",
    val name: String = "",
    val yteBalance: Double = 0.0,
    val usdBalance: Double = 0.0,
    val price: Double = 0.0,
    val liquidity: Double = 0.0,
    val lastBlock: Long = 0,
    val candles: List<CandleResponse> = emptyList(),
    val isLoadingBalance: Boolean = false,
    val isLoadingMarket: Boolean = false,
    val error: String? = null
)

class HomeViewModel(
    application: Application,
    private val balanceRepository: BalanceRepository,
    private val marketRepository: MarketRepository,
    private val sessionManager: SessionManager
) : AndroidViewModel(application) {

    private val _state = MutableStateFlow(HomeUiState())
    val state: StateFlow<HomeUiState> = _state.asStateFlow()

    init {
        loadBalance()
        loadMarket()
    }

    fun loadBalance() {
        val address = sessionManager.getAddress()?.lowercase()
        if (address == null) {
            _state.value = _state.value.copy(error = "Address not found")
            return
        }

        _state.value = _state.value.copy(isLoadingBalance = true, error = null)

        viewModelScope.launch {
            when (val result = balanceRepository.getBalance(address)) {
                is NetworkResult.Success -> {
                    _state.value = _state.value.copy(
                        isLoadingBalance = false,
                        address = result.data.address,
                        name = result.data.name,
                        yteBalance = result.data.yteBalance,
                        usdBalance = result.data.usdBalance,
                        error = null
                    )
                }
                is NetworkResult.Error -> {
                    _state.value = _state.value.copy(
                        isLoadingBalance = false,
                        error = result.message
                    )
                }
                is NetworkResult.Loading -> {}
            }
        }
    }

    fun loadMarket() {
        _state.value = _state.value.copy(isLoadingMarket = true)

        viewModelScope.launch {
            var hasError = false
            var errorMsg = ""

            val marketResult = marketRepository.getMarketState()
            when (marketResult) {
                is NetworkResult.Success -> {
                    _state.value = _state.value.copy(
                        price = marketResult.data.price,
                        liquidity = marketResult.data.liquidity,
                        lastBlock = marketResult.data.lastBlock
                    )
                }
                is NetworkResult.Error -> {
                    hasError = true
                    errorMsg = marketResult.message
                }
                is NetworkResult.Loading -> {}
            }

            val candlesResult = marketRepository.getCandles("1m", 50)
            when (candlesResult) {
                is NetworkResult.Success -> {
                    _state.value = _state.value.copy(candles = candlesResult.data)
                }
                is NetworkResult.Error -> {
                    if (!hasError) {
                        hasError = true
                        errorMsg = candlesResult.message
                    }
                }
                is NetworkResult.Loading -> {}
            }

            _state.value = _state.value.copy(
                isLoadingMarket = false,
                error = if (hasError && _state.value.error == null) errorMsg else _state.value.error
            )
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
