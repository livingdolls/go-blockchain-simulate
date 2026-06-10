class Block {
  final int id;
  final int blockNumber;
  final String previousHash;
  final String currentHash;
  final int nonce;
  final int difficulty;
  final int timestamp;
  final String merkleRoot;
  final String minerAddress;
  final double blockReward;
  final double totalFees;
  final String createdAt;
  final List<Transaction>? transactions;

  Block({
    required this.id,
    required this.blockNumber,
    required this.previousHash,
    required this.currentHash,
    required this.nonce,
    required this.difficulty,
    required this.timestamp,
    required this.merkleRoot,
    required this.minerAddress,
    required this.blockReward,
    required this.totalFees,
    required this.createdAt,
    this.transactions,
  });

  factory Block.fromJson(Map<String, dynamic> json) {
    return Block(
      id: json['id'] as int,
      blockNumber: json['block_number'] as int,
      previousHash: json['previous_hash'] as String? ?? '',
      currentHash: json['current_hash'] as String? ?? '',
      nonce: json['nonce'] as int? ?? 0,
      difficulty: json['difficulty'] as int? ?? 0,
      timestamp: json['timestamp'] as int? ?? 0,
      merkleRoot: json['merkle_root'] as String? ?? '',
      minerAddress: json['miner_address'] as String? ?? '',
      blockReward: (json['block_reward'] as num?)?.toDouble() ?? 0,
      totalFees: (json['total_fees'] as num?)?.toDouble() ?? 0,
      createdAt: json['created_at'] as String? ?? '',
      transactions: (json['transactions'] as List<dynamic>?)
          ?.map((t) => Transaction.fromJson(t as Map<String, dynamic>))
          .toList(),
    );
  }

  DateTime get dateTime =>
      DateTime.fromMillisecondsSinceEpoch(timestamp * 1000);

  String get shortHash {
    if (currentHash.length <= 14) return currentHash;
    return '${currentHash.substring(0, 8)}...${currentHash.substring(currentHash.length - 6)}';
  }
}

class Transaction {
  final int id;
  final String fromAddress;
  final String toAddress;
  final double amount;
  final double fee;
  final String type;
  final String signature;
  final String status;
  final String createdAt;

  Transaction({
    required this.id,
    required this.fromAddress,
    required this.toAddress,
    required this.amount,
    required this.fee,
    required this.type,
    required this.signature,
    required this.status,
    required this.createdAt,
  });

  factory Transaction.fromJson(Map<String, dynamic> json) {
    return Transaction(
      id: json['id'] as int,
      fromAddress: json['from_address'] as String? ?? '',
      toAddress: json['to_address'] as String? ?? '',
      amount: (json['amount'] as num?)?.toDouble() ?? 0,
      fee: (json['fee'] as num?)?.toDouble() ?? 0,
      type: json['type'] as String? ?? 'TRANSFER',
      signature: json['signature'] as String? ?? '',
      status: json['status'] as String? ?? 'pending',
      createdAt: json['created_at'] as String? ?? '',
    );
  }

  bool get isPending => status == 'pending';
  bool get isConfirmed => status == 'confirmed';
  bool get isTransfer => type == 'TRANSFER';
  bool get isBuy => type == 'BUY';
  bool get isSell => type == 'SELL';

  String get typeLabel {
    switch (type) {
      case 'TRANSFER':
        return 'Transfer';
      case 'BUY':
        return 'Beli YTE';
      case 'SELL':
        return 'Jual YTE';
      default:
        return type;
    }
  }
}

class BlockStats {
  final int totalBlocks;
  final double averageDifficulty;
  final int totalTransactions;
  final double totalFees;
  final double averageBlockReward;
  final double avgTxPerBlock;
  final int latestBlockNumber;

  BlockStats({
    required this.totalBlocks,
    required this.averageDifficulty,
    required this.totalTransactions,
    required this.totalFees,
    required this.averageBlockReward,
    required this.avgTxPerBlock,
    required this.latestBlockNumber,
  });

  factory BlockStats.fromJson(Map<String, dynamic> json) {
    return BlockStats(
      totalBlocks: json['total_blocks'] as int? ?? 0,
      averageDifficulty:
          (json['average_difficulty'] as num?)?.toDouble() ?? 0,
      totalTransactions: json['total_transactions'] as int? ?? 0,
      totalFees: (json['total_fees'] as num?)?.toDouble() ?? 0,
      averageBlockReward:
          (json['average_block_reward'] as num?)?.toDouble() ?? 0,
      avgTxPerBlock: (json['avg_tx_per_block'] as num?)?.toDouble() ?? 0,
      latestBlockNumber: json['latest_block_number'] as int? ?? 0,
    );
  }
}

class RewardInfo {
  final int currentBlockNumber;
  final double currentReward;
  final double nextReward;
  final int nextHalvingBlock;
  final int blocksUntilHalving;
  final double currentSupply;
  final double maxSupply;
  final double supplyPercentage;

  RewardInfo({
    required this.currentBlockNumber,
    required this.currentReward,
    required this.nextReward,
    required this.nextHalvingBlock,
    required this.blocksUntilHalving,
    required this.currentSupply,
    required this.maxSupply,
    required this.supplyPercentage,
  });

  factory RewardInfo.fromJson(Map<String, dynamic> json) {
    return RewardInfo(
      currentBlockNumber: json['current_block_number'] as int? ?? 0,
      currentReward: (json['current_reward'] as num?)?.toDouble() ?? 0,
      nextReward: (json['next_reward'] as num?)?.toDouble() ?? 0,
      nextHalvingBlock: json['next_halving_block'] as int? ?? 0,
      blocksUntilHalving: json['blocks_until_halving'] as int? ?? 0,
      currentSupply: (json['current_supply'] as num?)?.toDouble() ?? 0,
      maxSupply: (json['max_supply'] as num?)?.toDouble() ?? 0,
      supplyPercentage:
          (json['supply_percentage'] as num?)?.toDouble() ?? 0,
    );
  }
}

class MarketState {
  final double price;
  final double volume24h;
  final double high24h;
  final double low24h;
  final double change24h;

  MarketState({
    required this.price,
    required this.volume24h,
    required this.high24h,
    required this.low24h,
    required this.change24h,
  });

  factory MarketState.fromJson(Map<String, dynamic> json) {
    return MarketState(
      price: (json['price'] as num?)?.toDouble() ?? 0,
      volume24h: (json['volume_24h'] as num?)?.toDouble() ?? 0,
      high24h: (json['high_24h'] as num?)?.toDouble() ?? 0,
      low24h: (json['low_24h'] as num?)?.toDouble() ?? 0,
      change24h: (json['change_24h'] as num?)?.toDouble() ?? 0,
    );
  }
}

class Candle {
  final String interval;
  final double open;
  final double high;
  final double low;
  final double close;
  final double volume;
  final int timestamp;

  Candle({
    required this.interval,
    required this.open,
    required this.high,
    required this.low,
    required this.close,
    required this.volume,
    required this.timestamp,
  });

  factory Candle.fromJson(Map<String, dynamic> json) {
    return Candle(
      interval: json['interval'] as String? ?? '',
      open: (json['open'] as num?)?.toDouble() ?? 0,
      high: (json['high'] as num?)?.toDouble() ?? 0,
      low: (json['low'] as num?)?.toDouble() ?? 0,
      close: (json['close'] as num?)?.toDouble() ?? 0,
      volume: (json['volume'] as num?)?.toDouble() ?? 0,
      timestamp: json['timestamp'] as int? ?? 0,
    );
  }

  DateTime get dateTime =>
      DateTime.fromMillisecondsSinceEpoch(timestamp * 1000);
  bool get isBullish => close >= open;
  bool get isBearish => close < open;
}

class UserBalance {
  final String address;
  final String name;
  final double yteBalance;
  final double usdBalance;

  UserBalance({
    required this.address,
    required this.name,
    required this.yteBalance,
    required this.usdBalance,
  });

  factory UserBalance.fromJson(Map<String, dynamic> json) {
    return UserBalance(
      address: json['address'] as String? ?? '',
      name: json['name'] as String? ?? '',
      yteBalance: (json['yte_balance'] as num?)?.toDouble() ?? 0,
      usdBalance: (json['usd_balance'] as num?)?.toDouble() ?? 0,
    );
  }
}

class TopUpResult {
  final String address;
  final double amount;
  final double balanceBefore;
  final double balanceAfter;
  final String? referenceId;
  final String? description;

  TopUpResult({
    required this.address,
    required this.amount,
    required this.balanceBefore,
    required this.balanceAfter,
    this.referenceId,
    this.description,
  });

  factory TopUpResult.fromJson(Map<String, dynamic> json) {
    return TopUpResult(
      address: json['address'] as String? ?? '',
      amount: (json['amount'] as num?)?.toDouble() ?? 0,
      balanceBefore: (json['balance_before'] as num?)?.toDouble() ?? 0,
      balanceAfter: (json['balance_after'] as num?)?.toDouble() ?? 0,
      referenceId: json['reference_id'] as String?,
      description: json['description'] as String?,
    );
  }
}

/// Wrapper untuk response API.
class ApiResponse<T> {
  final bool success;
  final T? data;
  final String? error;

  ApiResponse({required this.success, this.data, this.error});

  factory ApiResponse.fromJson(
    Map<String, dynamic> json,
    T Function(Map<String, dynamic>) fromJson,
  ) {
    return ApiResponse(
      success: json['success'] as bool? ?? false,
      data: json['data'] != null
          ? fromJson(json['data'] as Map<String, dynamic>)
          : null,
      error: json['error'] as String?,
    );
  }

  factory ApiResponse.fromJsonList(
    Map<String, dynamic> json,
    T Function(List<dynamic>) fromJson,
  ) {
    return ApiResponse(
      success: json['success'] as bool? ?? false,
      data: json['data'] != null
          ? fromJson(json['data'] as List<dynamic>)
          : null,
      error: json['error'] as String?,
    );
  }
}

/// Response untuk wallet transactions.
class TransactionWithTypeResponse {
  final List<Transaction> transactions;
  final int total;
  final int page;
  final int limit;
  final int totalPages;

  TransactionWithTypeResponse({
    required this.transactions,
    required this.total,
    required this.page,
    required this.limit,
    required this.totalPages,
  });

  factory TransactionWithTypeResponse.fromJson(Map<String, dynamic> json) {
    return TransactionWithTypeResponse(
      transactions: (json['transactions'] as List<dynamic>?)
              ?.map((t) => Transaction.fromJson(t as Map<String, dynamic>))
              .toList() ??
          [],
      total: json['total'] as int? ?? 0,
      page: json['page'] as int? ?? 1,
      limit: json['limit'] as int? ?? 10,
      totalPages: json['total_pages'] as int? ?? 0,
    );
  }
}
