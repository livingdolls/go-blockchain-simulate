package com.takahashi.yutecoin.data.dto

import com.google.gson.annotations.SerializedName

data class NotificationItem(
    @SerializedName("id") val id: String,
    @SerializedName("type") val type: String,
    @SerializedName("priority") val priority: String,
    @SerializedName("recipient_address") val recipientAddress: String,
    @SerializedName("title") val title: String,
    @SerializedName("message") val message: String?,
    @SerializedName("is_read") val isRead: Boolean,
    @SerializedName("created_at") val createdAt: Long
)

data class NotificationListResponse(
    @SerializedName("notifications") val notifications: List<NotificationItem>?,
    @SerializedName("unread_count") val unreadCount: Int,
    @SerializedName("limit") val limit: Int,
    @SerializedName("offset") val offset: Int
)
