package com.takahashi.yutecoin.ui.dashboard

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.takahashi.yutecoin.data.dto.TransactionItem
import org.koin.androidx.compose.koinViewModel

@Composable
fun TransactionHistoryScreen(
    viewModel: TransactionHistoryViewModel = koinViewModel()
) {
    val uiState by viewModel.state.collectAsStateWithLifecycle()

    LaunchedEffect(Unit) {
        viewModel.loadTransactions()
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(horizontal = 16.dp, vertical = 12.dp)
    ) {
        Text(
            text = "Transactions",
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.Bold,
            color = MaterialTheme.colorScheme.onSurface
        )

        Spacer(Modifier.height(12.dp))

        if (uiState.isLoading && uiState.transactions.isEmpty()) {
            Box(Modifier.fillMaxSize(), Alignment.Center) {
                CircularProgressIndicator(Modifier.size(28.dp))
            }
        } else if (uiState.error != null && uiState.transactions.isEmpty()) {
            Box(Modifier.fillMaxSize(), Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(uiState.error ?: "Error", color = MaterialTheme.colorScheme.error)
                    Spacer(Modifier.height(8.dp))
                    TextButton(onClick = { viewModel.loadTransactions() }) { Text("Retry") }
                }
            }
        } else {
            LazyColumn(
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                items(uiState.transactions) { tx ->
                    TransactionCard(tx)
                }

                if (uiState.totalPages > 1) {
                    item {
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.Center,
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            TextButton(
                                onClick = { viewModel.loadTransactions(uiState.page - 1) },
                                enabled = uiState.page > 1
                            ) { Text("Prev") }
                            Text(
                                "${uiState.page}/${uiState.totalPages}",
                                modifier = Modifier.padding(horizontal = 12.dp),
                                style = MaterialTheme.typography.labelSmall,
                                color = MaterialTheme.colorScheme.onSurfaceVariant
                            )
                            TextButton(
                                onClick = { viewModel.loadTransactions(uiState.page + 1) },
                                enabled = uiState.page < uiState.totalPages
                            ) { Text("Next") }
                        }
                        Spacer(Modifier.height(16.dp))
                    }
                }
            }
        }
    }
}

@Composable
private fun TransactionCard(tx: TransactionItem) {
    val isSent = tx.fromAddress.equals(tx.toAddress, ignoreCase = true).not() // Simplified: all non-self are shown
    val typeColor = when (tx.type) {
        "BUY" -> Color(0xFF4CAF50)
        "SELL" -> Color(0xFFFF5722)
        else -> if (isSent) Color(0xFFFF9800) else Color(0xFF4CAF50)
    }
    val typeLabel = when (tx.type) {
        "BUY" -> "Buy"
        "SELL" -> "Sell"
        else -> "Transfer"
    }
    val sign = when (tx.type) {
        "SELL" -> "+"
        "BUY" -> "-"
        else -> ""
    }
    val amountColor = when (tx.type) {
        "BUY" -> Color(0xFF4CAF50)
        "SELL" -> Color(0xFFFF5722)
        else -> Color.Unspecified
    }

    Card(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.3f))
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(14.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Box(
                modifier = Modifier
                    .size(36.dp)
                    .padding(4.dp),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = if (tx.type == "BUY") "\u25B2" else if (tx.type == "SELL") "\u25BC" else "\u2192",
                    color = typeColor,
                    fontSize = 16.sp
                )
            }

            Spacer(Modifier.width(10.dp))

            Column(Modifier.weight(1f)) {
                Text(typeLabel, style = MaterialTheme.typography.labelMedium, fontWeight = FontWeight.SemiBold)
                Text(
                    if (tx.type == "TRANSFER") "${tx.fromAddress.take(6)}...${tx.fromAddress.takeLast(4)} \u2192 ${tx.toAddress.take(6)}...${tx.toAddress.takeLast(4)}"
                    else "Amount: ${tx.amount}",
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
            }

            Column(horizontalAlignment = Alignment.End) {
                Text(
                    "${sign}${"%.4f".format(tx.amount)} YTE",
                    style = MaterialTheme.typography.labelMedium,
                    fontWeight = FontWeight.SemiBold,
                    color = if (tx.type != "TRANSFER") amountColor else MaterialTheme.colorScheme.onSurface
                )
                Text(
                    tx.status.lowercase().replaceFirstChar { it.uppercase() },
                    style = MaterialTheme.typography.labelSmall,
                    color = if (tx.status == "SUCCESS") Color(0xFF4CAF50) else MaterialTheme.colorScheme.onSurfaceVariant,
                    fontSize = 10.sp
                )
            }
        }
    }
}
