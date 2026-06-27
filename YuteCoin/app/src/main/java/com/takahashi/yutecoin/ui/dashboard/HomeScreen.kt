package com.takahashi.yutecoin.ui.dashboard

import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
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
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.PathEffect
import androidx.compose.ui.graphics.nativeCanvas
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.takahashi.yutecoin.data.dto.CandleResponse
import org.koin.androidx.compose.koinViewModel
import kotlin.math.max
import kotlin.math.min

@Composable
fun HomeScreen(
    onLogout: () -> Unit,
    viewModel: HomeViewModel = koinViewModel()
) {
    val uiState by viewModel.state.collectAsStateWithLifecycle()

    Scaffold(modifier = Modifier.fillMaxSize()) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .verticalScroll(rememberScrollState()),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            TopBar(
                name = uiState.name,
                onRefresh = { viewModel.loadMarket(); viewModel.loadBalance() },
                onLogout = {
                    viewModel.logout()
                    onLogout()
                }
            )

            Spacer(modifier = Modifier.height(8.dp))

            if (uiState.error != null && uiState.price == 0.0 && uiState.yteBalance == 0.0) {
                ErrorCard(
                    error = uiState.error ?: "",
                    onRetry = { viewModel.loadMarket(); viewModel.loadBalance() }
                )
            } else {
                PriceTicker(
                    price = uiState.price,
                    isLoading = uiState.isLoadingMarket
                )

                Spacer(modifier = Modifier.height(12.dp))

                if (uiState.candles.isNotEmpty()) {
                    CandleChart(
                        candles = uiState.candles,
                        isLiveConnected = uiState.isLiveConnected,
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 16.dp)
                            .height(220.dp)
                    )
                } else if (uiState.isLoadingMarket) {
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(160.dp),
                        contentAlignment = Alignment.Center
                    ) {
                        CircularProgressIndicator(modifier = Modifier.size(28.dp))
                    }
                } else {
                    Card(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 16.dp)
                            .height(100.dp),
                        shape = RoundedCornerShape(16.dp),
                        colors = CardDefaults.cardColors(
                            containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.3f)
                        )
                    ) {
                        Box(
                            modifier = Modifier.fillMaxSize(),
                            contentAlignment = Alignment.Center
                        ) {
                            Text(
                                text = "No price data yet\nSend a transaction to start",
                                style = MaterialTheme.typography.bodyMedium,
                                color = MaterialTheme.colorScheme.onSurfaceVariant,
                                textAlign = androidx.compose.ui.text.style.TextAlign.Center
                            )
                        }
                    }
                }

                Spacer(modifier = Modifier.height(16.dp))

                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 16.dp),
                    horizontalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    BalanceCard(
                        modifier = Modifier.weight(1f),
                        label = "YTE Balance",
                        amount = formatBalance(uiState.yteBalance),
                        symbol = "YTE",
                        isLoading = uiState.isLoadingBalance
                    )
                    BalanceCard(
                        modifier = Modifier.weight(1f),
                        label = "USD Balance",
                        amount = formatBalance(uiState.usdBalance),
                        symbol = "USD",
                        isLoading = uiState.isLoadingBalance
                    )
                }

                Spacer(modifier = Modifier.height(16.dp))

                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 16.dp),
                    horizontalArrangement = Arrangement.spacedBy(12.dp)
                ) {
                    StatItem(
                        modifier = Modifier.weight(1f),
                        label = "Liquidity",
                        value = formatNumber(uiState.liquidity)
                    )
                    StatItem(
                        modifier = Modifier.weight(1f),
                        label = "Last Block",
                        value = uiState.lastBlock.toString()
                    )
                }

                Spacer(modifier = Modifier.height(24.dp))
            }
        }
    }
}

@Composable
private fun TopBar(
    name: String,
    onRefresh: () -> Unit,
    onLogout: () -> Unit
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 20.dp, vertical = 12.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Box(
            modifier = Modifier
                .size(40.dp)
                .clip(CircleShape)
                .background(MaterialTheme.colorScheme.primary.copy(alpha = 0.15f)),
            contentAlignment = Alignment.Center
        ) {
            Text(
                text = name.firstOrNull()?.uppercase() ?: "?",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold,
                color = MaterialTheme.colorScheme.primary
            )
        }

        Spacer(modifier = Modifier.width(12.dp))

        Column(modifier = Modifier.weight(1f)) {
            Text(
                text = "Hello, $name",
                style = MaterialTheme.typography.titleSmall,
                fontWeight = FontWeight.SemiBold,
                color = MaterialTheme.colorScheme.onSurface
            )
            Text(
                text = "Welcome back",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }

        IconButton(onClick = onRefresh) {
            Text(
                text = "\u27F3",
                style = MaterialTheme.typography.titleLarge,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }

        IconButton(onClick = onLogout) {
            Text(
                text = "Exit",
                style = MaterialTheme.typography.labelMedium,
                color = MaterialTheme.colorScheme.error
            )
        }
    }
}

@Composable
private fun PriceTicker(
    price: Double,
    isLoading: Boolean
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f)
        )
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 20.dp, vertical = 16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = "YTE / USD",
                    style = MaterialTheme.typography.labelMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                if (isLoading && price == 0.0) {
                    CircularProgressIndicator(
                        modifier = Modifier.size(16.dp),
                        strokeWidth = 2.dp
                    )
                } else {
                    Text(
                        text = "$${formatPrice(price)}",
                        style = MaterialTheme.typography.headlineMedium,
                        fontWeight = FontWeight.Bold,
                        color = MaterialTheme.colorScheme.onSurface
                    )
                }
            }

            PriceBadge(price = price)
        }
    }
}

@Composable
private fun PriceBadge(price: Double) {
    val color = if (price >= 1.0) Color(0xFF4CAF50) else Color(0xFFFF5722)
    val text = if (price >= 1.0) "+${formatPrice(price - 1.0)}" else formatPrice(price - 1.0)
    Box(
        modifier = Modifier
            .clip(RoundedCornerShape(8.dp))
            .background(color.copy(alpha = 0.12f))
            .padding(horizontal = 10.dp, vertical = 4.dp)
    ) {
        Text(
            text = "$$text",
            style = MaterialTheme.typography.labelMedium,
            fontWeight = FontWeight.SemiBold,
            color = color
        )
    }
}

@Composable
private fun CandleChart(
    candles: List<CandleResponse>,
    isLiveConnected: Boolean = false,
    modifier: Modifier = Modifier
) {
    val greenColor = Color(0xFF26A69A)
    val redColor = Color(0xFFEF5350)
    val gridColor = Color(0xFF2A2A2A)

    Card(
        modifier = modifier,
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.3f)
        )
    ) {
        Column(modifier = Modifier.padding(8.dp)) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 4.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "YTE Price",
                    style = MaterialTheme.typography.labelMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    modifier = Modifier.weight(1f)
                )
                if (isLiveConnected) {
                    Box(
                        modifier = Modifier
                            .size(8.dp)
                            .clip(CircleShape)
                            .background(Color(0xFF4CAF50))
                    )
                    Spacer(modifier = Modifier.width(4.dp))
                    Text(
                        text = "LIVE",
                        style = MaterialTheme.typography.labelSmall,
                        color = Color(0xFF4CAF50),
                        fontWeight = FontWeight.Bold
                    )
                }
            }

            Spacer(modifier = Modifier.height(4.dp))

            Canvas(
                modifier = Modifier
                    .fillMaxWidth()
                    .weight(1f)
            ) {
                val canvasWidth = size.width
                val canvasHeight = size.height
                val textAreaH = 28f
                val chartAreaH = canvasHeight - textAreaH
                val leftPad = 4f
                val rightPad = 4f
                val topPad = 8f
                val bottomPad = 4f

                val chartW = canvasWidth - leftPad - rightPad
                val chartH = chartAreaH - topPad - bottomPad

                val minPrice = candles.minOf { minOf(it.lowPrice, it.openPrice, it.closePrice) }
                val maxPrice = candles.maxOf { maxOf(it.highPrice, it.openPrice, it.closePrice) }
                val dataRange = if (maxPrice - minPrice < 0.00001) 0.001 else maxPrice - minPrice
                val padding = dataRange * 0.15
                val displayMin = minPrice - padding
                val displayMax = maxPrice + padding
                val range = displayMax - displayMin

                val candleCount = candles.size
                val candleSpacing = 1.5f
                val totalSpacing = (candleCount - 1) * candleSpacing
                val candleWidth = ((chartW - totalSpacing) / candleCount).coerceIn(1f, 16f)
                val stepX = candleWidth + candleSpacing

                // grid lines
                val gridCount = 4
                for (i in 0..gridCount) {
                    val y = topPad + chartH * i / gridCount
                    drawLine(
                        color = gridColor.copy(alpha = 0.3f),
                        start = Offset(leftPad, y),
                        end = Offset(leftPad + chartW, y),
                        strokeWidth = 0.5f,
                        pathEffect = PathEffect.dashPathEffect(floatArrayOf(4f, 4f))
                    )
                }

                // candles
                candles.forEachIndexed { index, candle ->
                    val x = leftPad + index * stepX
                    val openY = topPad + chartH - ((candle.openPrice - displayMin) / range * chartH).toFloat()
                    val closeY = topPad + chartH - ((candle.closePrice - displayMin) / range * chartH).toFloat()
                    val highY = topPad + chartH - ((candle.highPrice - displayMin) / range * chartH).toFloat()
                    val lowY = topPad + chartH - ((candle.lowPrice - displayMin) / range * chartH).toFloat()

                    val isBullish = candle.closePrice >= candle.openPrice
                    val color = if (isBullish) greenColor else redColor
                    val bodyTop = min(openY, closeY)
                    val bodyBottom = max(openY, closeY)
                    val bodyHeight = (bodyBottom - bodyTop).coerceAtLeast(1f)

                    val centerX = x + candleWidth / 2f

                    // wick
                    drawLine(
                        color = color,
                        start = Offset(centerX, highY),
                        end = Offset(centerX, lowY),
                        strokeWidth = 1f
                    )

                    // body
                    drawRect(
                        color = color,
                        topLeft = Offset(x, bodyTop),
                        size = Size(candleWidth, bodyHeight)
                    )
                }

                // price labels
                val labelPaint = android.graphics.Paint().apply {
                    textSize = 9.sp.toPx()
                    color = 0xFF888888.toInt()
                    textAlign = android.graphics.Paint.Align.LEFT
                    isAntiAlias = true
                }

                for (i in 0..gridCount) {
                    val price = displayMax - range * i / gridCount
                    val y = topPad + chartH * i / gridCount + 4f
                    drawContext.canvas.nativeCanvas.drawText(
                        "$${formatPrice(price)}",
                        leftPad,
                        y,
                        labelPaint
                    )
                }
            }
        }
    }
}

@Composable
private fun BalanceCard(
    modifier: Modifier = Modifier,
    label: String,
    amount: String,
    symbol: String,
    isLoading: Boolean
) {
    Card(
        modifier = modifier,
        shape = RoundedCornerShape(14.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surface
        ),
        elevation = CardDefaults.cardElevation(defaultElevation = 1.dp)
    ) {
        Column(
            modifier = Modifier.padding(14.dp),
            horizontalAlignment = Alignment.Start
        ) {
            Text(
                text = label,
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            Spacer(modifier = Modifier.height(6.dp))
            if (isLoading && amount == "0") {
                CircularProgressIndicator(
                    modifier = Modifier.size(14.dp),
                    strokeWidth = 2.dp
                )
            } else {
                Text(
                    text = amount,
                    style = MaterialTheme.typography.titleLarge,
                    fontWeight = FontWeight.Bold,
                    color = MaterialTheme.colorScheme.onSurface
                )
                Text(
                    text = symbol,
                    style = MaterialTheme.typography.labelSmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }
    }
}

@Composable
private fun StatItem(
    modifier: Modifier = Modifier,
    label: String,
    value: String
) {
    Card(
        modifier = modifier,
        shape = RoundedCornerShape(12.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.4f)
        )
    ) {
        Column(modifier = Modifier.padding(14.dp)) {
            Text(
                text = label,
                style = MaterialTheme.typography.labelSmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                text = value,
                style = MaterialTheme.typography.bodyLarge,
                fontWeight = FontWeight.SemiBold,
                color = MaterialTheme.colorScheme.onSurface
            )
        }
    }
}

@Composable
private fun ErrorCard(
    error: String,
    onRetry: () -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 16.dp, vertical = 8.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.errorContainer
        ),
        shape = RoundedCornerShape(12.dp)
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Text(
                text = "Unable to load data",
                style = MaterialTheme.typography.titleSmall,
                color = MaterialTheme.colorScheme.onErrorContainer
            )
            Spacer(modifier = Modifier.height(4.dp))
            Text(
                text = error,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onErrorContainer
            )
            Spacer(modifier = Modifier.height(12.dp))
            OutlinedButton(onClick = onRetry) {
                Text("Retry")
            }
        }
    }
}

private fun formatBalance(value: Double): String {
    return when {
        value >= 1_000_000 -> "%.2fM".format(value / 1_000_000)
        value >= 1_000 -> "%.2fK".format(value / 1_000)
        value == value.toLong().toDouble() -> "%.0f".format(value)
        else -> "%.4f".format(value)
    }
}

private fun formatPrice(value: Double): String {
    return when {
        value == value.toLong().toDouble() -> "%.2f".format(value)
        value >= 10 -> "%.4f".format(value)
        value >= 1 -> "%.6f".format(value)
        else -> "%.8f".format(value)
    }
}

private fun formatNumber(value: Double): String {
    return when {
        value >= 1_000_000 -> "%.1fM".format(value / 1_000_000)
        value >= 1_000 -> "%.1fK".format(value / 1_000)
        value == value.toLong().toDouble() -> "%.0f".format(value)
        else -> "%.2f".format(value)
    }
}
