package com.takahashi.yutecoin.ui.dashboard

import android.graphics.Color as AndroidColor
import android.graphics.Paint
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import com.github.mikephil.charting.charts.CandleStickChart
import com.github.mikephil.charting.components.XAxis
import com.github.mikephil.charting.components.YAxis
import com.github.mikephil.charting.data.CandleData
import com.github.mikephil.charting.data.CandleDataSet
import com.github.mikephil.charting.data.CandleEntry
import com.takahashi.yutecoin.data.dto.CandleResponse
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

@Composable
fun CandleStickChartView(
    candles: List<CandleResponse>,
    modifier: Modifier = Modifier
) {
    val greenColor = AndroidColor.rgb(38, 166, 154)
    val redColor = AndroidColor.rgb(239, 83, 80)

    var chartRef by remember { mutableStateOf<CandleStickChart?>(null) }

    Card(
        modifier = modifier,
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.3f)
        )
    ) {
        AndroidView(
            factory = { context ->
                CandleStickChart(context).apply {
                    description.isEnabled = false
                    legend.isEnabled = false
                    setDrawBorders(false)
                    setMaxVisibleValueCount(0)
                    setPinchZoom(true)
                    isAutoScaleMinMaxEnabled = false
                    setScaleEnabled(true)
                    setExtraOffsets(4f, 0f, 4f, 12f)

                    xAxis.apply {
                        position = XAxis.XAxisPosition.BOTTOM
                        setDrawGridLines(false)
                        setDrawAxisLine(true)
                        textColor = AndroidColor.GRAY
                        textSize = 9f
                        granularity = 5f
                    }

                    axisLeft.apply {
                        setDrawGridLines(true)
                        setDrawAxisLine(false)
                        gridColor = AndroidColor.argb(40, 255, 255, 255)
                        textColor = AndroidColor.GRAY
                        textSize = 9f
                        setPosition(YAxis.YAxisLabelPosition.INSIDE_CHART)
                        setDrawTopYLabelEntry(true)
                    }

                    axisRight.isEnabled = false

                    chartRef = this
                }
            },
            update = { chart ->
                if (candles.isEmpty()) return@AndroidView

                val sorted = candles.sortedBy { it.startTime }

                val entries = sorted.mapIndexed { index, candle ->
                    val prevClose = if (index > 0) sorted[index - 1].closePrice else candle.openPrice
                    val trend = candle.closePrice >= prevClose
                    val adjOpen = if (candle.openPrice != candle.closePrice) candle.openPrice.toFloat()
                        else if (trend) (candle.openPrice - 0.000001).toFloat()
                        else (candle.openPrice + 0.000001).toFloat()

                    CandleEntry(
                        index.toFloat(),
                        candle.highPrice.toFloat(),
                        candle.lowPrice.toFloat(),
                        adjOpen,
                        candle.closePrice.toFloat()
                    )
                }

                val dataSet = CandleDataSet(entries, "YTE").apply {
                    setDrawIcons(false)
                    axisDependency = YAxis.AxisDependency.LEFT
                    shadowColor = AndroidColor.DKGRAY
                    shadowWidth = 1.8f
                    barSpace = 0.2f
                    decreasingColor = redColor
                    decreasingPaintStyle = Paint.Style.FILL
                    increasingColor = greenColor
                    increasingPaintStyle = Paint.Style.FILL
                    neutralColor = AndroidColor.LTGRAY
                    setDrawValues(false)
                }

                chart.data = CandleData(dataSet)
                chart.notifyDataSetChanged()

                val minPrice = sorted.minOf { it.lowPrice }
                val maxPrice = sorted.maxOf { it.highPrice }
                val range = (maxPrice - minPrice).coerceAtLeast(0.004)
                val padding = range * 0.2
                chart.axisLeft.apply {
                    axisMinimum = (minPrice - padding).toFloat()
                    axisMaximum = (maxPrice + padding).toFloat()
                }

                chart.xAxis.valueFormatter = TimeValueFormatter(sorted)
                chart.setVisibleXRangeMaximum(30f)
                chart.moveViewToX(sorted.size.toFloat())
                chart.invalidate()
            },
            modifier = Modifier
                .fillMaxSize()
                .padding(end = 4.dp)
        )
    }
}

private class TimeValueFormatter(
    private val candles: List<CandleResponse>
) : com.github.mikephil.charting.formatter.ValueFormatter() {
    private val dateFormat = SimpleDateFormat("HH:mm", Locale.getDefault())

    override fun getFormattedValue(value: Float): String {
        val index = value.toInt()
        if (index < 0 || index >= candles.size) return ""
        val timestamp = candles[index].startTime * 1000L
        return dateFormat.format(Date(timestamp))
    }
}
