package com.takahashi.yutecoin.ui.dashboard

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.takahashi.yutecoin.data.dto.BlockItem
import com.takahashi.yutecoin.data.dto.BlockStatsResponse
import com.takahashi.yutecoin.data.dto.NetworkResult
import com.takahashi.yutecoin.data.repository.BlockRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class BlockExplorerUiState(
    val blocks: List<BlockItem> = emptyList(),
    val stats: BlockStatsResponse? = null,
    val isLoading: Boolean = false,
    val error: String? = null
)

class BlockExplorerViewModel(
    private val blockRepository: BlockRepository
) : ViewModel() {

    private val _state = MutableStateFlow(BlockExplorerUiState())
    val state: StateFlow<BlockExplorerUiState> = _state.asStateFlow()

    init { load() }

    fun load() {
        _state.value = _state.value.copy(isLoading = true, error = null)
        viewModelScope.launch {
            val statsResult = blockRepository.getStats()
            if (statsResult is NetworkResult.Success) {
                _state.value = _state.value.copy(stats = statsResult.data)
            }

            when (val result = blockRepository.getBlocks(20, 0)) {
                is NetworkResult.Success -> {
                    _state.value = _state.value.copy(isLoading = false, blocks = result.data)
                }
                is NetworkResult.Error -> {
                    _state.value = _state.value.copy(isLoading = false, error = result.message)
                }
                is NetworkResult.Loading -> {}
            }
        }
    }
}
