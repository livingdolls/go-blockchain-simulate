package com.takahashi.yutecoin.ui.theme

import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.unit.IntSize
import kotlin.math.sqrt

class RevealController {
    var center by mutableStateOf(Offset.Zero)
    var triggerCount by mutableIntStateOf(0)
    var maxRadius by mutableFloatStateOf(0f)

    fun trigger(position: Offset, viewSize: IntSize) {
        center = position
        maxRadius = sqrt((viewSize.width * viewSize.width + viewSize.height * viewSize.height).toFloat())
        triggerCount++
    }
}

@Composable
fun rememberRevealController() = remember { RevealController() }

@Composable
fun ThemeRevealOverlay(
    controller: RevealController,
    isDark: Boolean,
    onMidpoint: () -> Unit
) {
    var size by remember { mutableStateOf(IntSize.Zero) }
    val animRadius = remember { Animatable(0f) }

    LaunchedEffect(controller.triggerCount) {
        if (controller.triggerCount > 0) {
            animRadius.snapTo(0f)
            animRadius.animateTo(controller.maxRadius * 0.5f, tween(200))
            onMidpoint()
            animRadius.animateTo(controller.maxRadius, tween(300))
        }
    }

    val r = animRadius.value
    if (r > 1f && controller.triggerCount > 0) {
        Canvas(
            modifier = Modifier
                .fillMaxSize()
                .onSizeChanged { size = it }
        ) {
            drawCircle(
                color = if (isDark) Color(0xFF1C1B1F) else Color(0xFFFFFBFE),
                radius = r,
                center = controller.center
            )
        }
    }
}
