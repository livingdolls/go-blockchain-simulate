package com.takahashi.yutecoin.data.dto

import com.google.gson.annotations.SerializedName

data class WalletResponse(
    @SerializedName("address") val address: String,
    @SerializedName("transactions") val transactions: TransactionListResponse
)

data class TransactionListResponse(
    @SerializedName("transactions") val transactions: List<TransactionItem>,
    @SerializedName("total") val total: Long,
    @SerializedName("page") val page: Int,
    @SerializedName("limit") val limit: Int,
    @SerializedName("total_pages") val totalPages: Int
)

data class TransactionItem(
    @SerializedName("id") val id: Long,
    @SerializedName("from_address") val fromAddress: String,
    @SerializedName("to_address") val toAddress: String,
    @SerializedName("amount") val amount: Double,
    @SerializedName("fee") val fee: Double,
    @SerializedName("type") val type: String,
    @SerializedName("status") val status: String,
    @SerializedName("created_at") val createdAt: String
)
