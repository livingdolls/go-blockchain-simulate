package com.takahashi.yutecoin.ui.dashboard

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.takahashi.yutecoin.data.dto.NetworkResult
import com.takahashi.yutecoin.data.dto.NotificationItem
import com.takahashi.yutecoin.data.local.SessionManager
import com.takahashi.yutecoin.data.repository.NotificationRepository
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

data class NotificationUiState(
    val notifications: List<NotificationItem> = emptyList(),
    val unreadCount: Int = 0,
    val isLoading: Boolean = false,
    val error: String? = null
)

class NotificationViewModel(
    private val notificationRepository: NotificationRepository,
    private val sessionManager: SessionManager
) : ViewModel() {

    private val _state = MutableStateFlow(NotificationUiState())
    val state: StateFlow<NotificationUiState> = _state.asStateFlow()

    init { load() }

    fun load() {
        val address = sessionManager.getAddress()?.lowercase() ?: return
        _state.value = _state.value.copy(isLoading = true)
        viewModelScope.launch {
            when (val result = notificationRepository.getNotifications(address, 50)) {
                is NetworkResult.Success -> {
                    _state.value = _state.value.copy(
                        isLoading = false,
                        notifications = result.data.notifications ?: emptyList(),
                        unreadCount = result.data.unreadCount
                    )
                }
                is NetworkResult.Error -> {
                    _state.value = _state.value.copy(isLoading = false, error = result.message)
                }
                is NetworkResult.Loading -> {}
            }
        }
    }

    fun markAsRead(id: String) {
        viewModelScope.launch {
            notificationRepository.markAsRead(id)
            _state.value = _state.value.copy(
                notifications = _state.value.notifications.map { if (it.id == id) it.copy(isRead = true) else it },
                unreadCount = (_state.value.unreadCount - 1).coerceAtLeast(0)
            )
        }
    }

    fun markAllAsRead() {
        val address = sessionManager.getAddress()?.lowercase() ?: return
        viewModelScope.launch {
            notificationRepository.markAllAsRead(address)
            _state.value = _state.value.copy(
                notifications = _state.value.notifications.map { it.copy(isRead = true) },
                unreadCount = 0
            )
        }
    }

    fun delete(id: String) {
        viewModelScope.launch {
            notificationRepository.delete(id)
            val wasUnread = _state.value.notifications.find { it.id == id }?.isRead == false
            _state.value = _state.value.copy(
                notifications = _state.value.notifications.filter { it.id != id },
                unreadCount = if (wasUnread) (_state.value.unreadCount - 1).coerceAtLeast(0) else _state.value.unreadCount
            )
        }
    }
}
