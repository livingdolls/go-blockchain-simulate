package com.takahashi.yutecoin.data.dto

import com.google.gson.annotations.SerializedName

data class BlockItem(
    @SerializedName("id") val id: Long,
    @SerializedName("block_number") val blockNumber: Int,
    @SerializedName("previous_hash") val previousHash: String,
    @SerializedName("current_hash") val currentHash: String,
    @SerializedName("nonce") val nonce: Long,
    @SerializedName("difficulty") val difficulty: Int,
    @SerializedName("timestamp") val timestamp: Long,
    @SerializedName("merkle_root") val merkleRoot: String?,
    @SerializedName("miner_address") val minerAddress: String?,
    @SerializedName("block_reward") val blockReward: Double,
    @SerializedName("total_fees") val totalFees: Double,
    @SerializedName("created_at") val createdAt: String?
)

data class BlockListResponse(
    @SerializedName("blocks") val blocks: List<BlockItem>?
)

data class BlockStatsResponse(
    @SerializedName("total_blocks") val totalBlocks: Long,
    @SerializedName("average_difficulty") val averageDifficulty: Double,
    @SerializedName("total_transactions") val totalTransactions: Long,
    @SerializedName("total_fees") val totalFees: Double,
    @SerializedName("average_block_reward") val averageBlockReward: Double,
    @SerializedName("avg_tx_per_block") val avgTxPerBlock: Double,
    @SerializedName("latest_block_number") val latestBlockNumber: Long
)
