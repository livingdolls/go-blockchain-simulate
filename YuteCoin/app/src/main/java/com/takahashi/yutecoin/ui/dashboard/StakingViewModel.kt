package com.takahashi.yutecoin.ui.dashboard

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.takahashi.yutecoin.data.dto.NetworkResult
import com.takahashi.yutecoin.data.dto.StakeRequest
import com.takahashi.yutecoin.data.dto.StakingInfoResponse
import com.takahashi.yutecoin.data.dto.StakingRecordItem
import com.takahashi.yutecoin.data.dto.UnstakeRequest
import com.takahashi.yutecoin.data.local.SessionManager
import com.takahashi.yutecoin.data.repository.StakingRepository
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import java.util.Locale

data class StakingUiState(
    val amount: String = "",
    val lockDays: String = "30",
    val info: StakingInfoResponse? = null,
    val records: List<StakingRecordItem> = emptyList(),
    val totalStaked: Double = 0.0,
    val totalRewards: Double = 0.0,
    val isSubmitting: Boolean = false,
    val isLoading: Boolean = false,
    val error: String? = null,
    val successMessage: String? = null
)

class StakingViewModel(
    private val stakingRepository: StakingRepository,
    private val sessionManager: SessionManager
) : ViewModel() {

    private val _state = MutableStateFlow(StakingUiState())
    val state: StateFlow<StakingUiState> = _state.asStateFlow()

    init {
        loadData()
    }

    fun loadData() {
        val address = sessionManager.getAddress()?.lowercase() ?: return
        _state.value = _state.value.copy(isLoading = true)

        viewModelScope.launch {
            val infoResult = stakingRepository.getInfo()
            if (infoResult is NetworkResult.Success) {
                _state.value = _state.value.copy(info = infoResult.data)
            }

            val statusResult = stakingRepository.getStatus(address)
            when (statusResult) {
                is NetworkResult.Success -> {
                    _state.value = _state.value.copy(
                        isLoading = false,
                        records = statusResult.data.records ?: emptyList(),
                        totalStaked = statusResult.data.totalStaked,
                        totalRewards = statusResult.data.totalRewards
                    )
                }
                is NetworkResult.Error -> {
                    _state.value = _state.value.copy(isLoading = false, error = statusResult.message)
                }
                is NetworkResult.Loading -> {}
            }
        }
    }

    fun onAmountChange(value: String) {
        _state.value = _state.value.copy(amount = value, error = null, successMessage = null)
    }

    fun onLockDaysChange(value: String) {
        _state.value = _state.value.copy(lockDays = value, error = null, successMessage = null)
    }

    fun stake() {
        val address = sessionManager.getAddress()?.lowercase() ?: return
        val amount = _state.value.amount.toDoubleOrNull()
        val days = _state.value.lockDays.toIntOrNull()

        if (amount == null || amount <= 0) {
            _state.value = _state.value.copy(error = "Enter a valid amount")
            return
        }
        if (days == null || days < 1) {
            _state.value = _state.value.copy(error = "Enter valid lock days (>0)")
            return
        }

        _state.value = _state.value.copy(isSubmitting = true, error = null)

        viewModelScope.launch {
            when (val result = stakingRepository.stake(StakeRequest(address, amount, days))) {
                is NetworkResult.Success -> {
                    _state.value = _state.value.copy(
                        isSubmitting = false,
                        successMessage = "Staked ${"%.4f".format(Locale.US, amount)} YTE for $days days!",
                        amount = ""
                    )
                    delay(2000)
                    loadData()
                }
                is NetworkResult.Error -> {
                    _state.value = _state.value.copy(isSubmitting = false, error = result.message)
                }
                is NetworkResult.Loading -> {}
            }
        }
    }

    fun unstake(stakeId: Long) {
        val address = sessionManager.getAddress()?.lowercase() ?: return
        _state.value = _state.value.copy(isSubmitting = true, error = null)

        viewModelScope.launch {
            when (val result = stakingRepository.unstake(UnstakeRequest(address, stakeId))) {
                is NetworkResult.Success -> {
                    _state.value = _state.value.copy(isSubmitting = false, successMessage = "Unstaked!")
                    delay(2000)
                    loadData()
                }
                is NetworkResult.Error -> {
                    _state.value = _state.value.copy(isSubmitting = false, error = result.message)
                }
                is NetworkResult.Loading -> {}
            }
        }
    }

    fun clearSuccess() {
        _state.value = _state.value.copy(successMessage = null)
    }
}
