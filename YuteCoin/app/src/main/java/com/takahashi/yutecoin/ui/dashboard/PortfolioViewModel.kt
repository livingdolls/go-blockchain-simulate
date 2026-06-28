package com.takahashi.yutecoin.ui.dashboard

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.takahashi.yutecoin.data.dto.NetworkResult
import com.takahashi.yutecoin.data.dto.PortfolioResponse
import com.takahashi.yutecoin.data.local.SessionManager
import com.takahashi.yutecoin.data.repository.PortfolioRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class PortfolioUiState(
    val data: PortfolioResponse? = null,
    val isLoading: Boolean = false,
    val error: String? = null
)

class PortfolioViewModel(
    private val portfolioRepository: PortfolioRepository,
    private val sessionManager: SessionManager
) : ViewModel() {

    private val _state = MutableStateFlow(PortfolioUiState())
    val state: StateFlow<PortfolioUiState> = _state.asStateFlow()

    init {
        load()
    }

    fun load() {
        val address = sessionManager.getAddress()?.lowercase() ?: return
        _state.value = _state.value.copy(isLoading = true, error = null)

        viewModelScope.launch {
            when (val result = portfolioRepository.getPortfolio(address)) {
                is NetworkResult.Success -> {
                    _state.value = _state.value.copy(isLoading = false, data = result.data)
                }
                is NetworkResult.Error -> {
                    _state.value = _state.value.copy(isLoading = false, error = result.message)
                }
                is NetworkResult.Loading -> {}
            }
        }
    }
}
