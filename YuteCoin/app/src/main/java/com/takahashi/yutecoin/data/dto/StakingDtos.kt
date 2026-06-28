package com.takahashi.yutecoin.data.dto

import com.google.gson.annotations.SerializedName

data class StakeRequest(
    @SerializedName("address") val address: String,
    @SerializedName("amount") val amount: Double,
    @SerializedName("lock_days") val lockDays: Int
)

data class StakeResponse(
    @SerializedName("stake_id") val stakeId: Long,
    @SerializedName("amount") val amount: Double,
    @SerializedName("lock_until") val lockUntil: Long,
    @SerializedName("status") val status: String,
    @SerializedName("days_locked") val daysLocked: Int
)

data class UnstakeRequest(
    @SerializedName("address") val address: String,
    @SerializedName("stake_id") val stakeId: Long
)

data class UnstakeResponse(
    @SerializedName("message") val message: String,
    @SerializedName("stake_id") val stakeId: Long
)

data class StakingStatusResponse(
    @SerializedName("total_staked") val totalStaked: Double,
    @SerializedName("total_rewards") val totalRewards: Double,
    @SerializedName("active_stakes") val activeStakes: Int,
    @SerializedName("records") val records: List<StakingRecordItem>
)

data class StakingRecordItem(
    @SerializedName("id") val id: Long,
    @SerializedName("user_address") val userAddress: String,
    @SerializedName("amount") val amount: Double,
    @SerializedName("lock_until") val lockUntil: Long,
    @SerializedName("reward_earned") val rewardEarned: Double,
    @SerializedName("status") val status: String,
    @SerializedName("created_at") val createdAt: String
)

data class StakingInfoResponse(
    @SerializedName("total_staked") val totalStaked: Double,
    @SerializedName("staking_apr") val stakingApr: Double,
    @SerializedName("min_stake_amount") val minStakeAmount: Double,
    @SerializedName("min_lock_duration_seconds") val minLockDurationSeconds: Long,
    @SerializedName("reward_per_block") val rewardPerBlock: Double,
    @SerializedName("next_reward_block") val nextRewardBlock: Long
)
