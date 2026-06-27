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
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.BlendMode
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.unit.IntSize
import kotlin.math.sqrt

class RevealController {
    var center by mutableStateOf(Offset.Zero)
    var triggerCount by mutableIntStateOf(0)
    var maxRadius by mutableFloatStateOf(0f)
    var targetDark by mutableStateOf(false)

    fun trigger(position: Offset, viewSize: IntSize, darkTarget: Boolean) {
        center = position
        targetDark = darkTarget
        maxRadius = sqrt((viewSize.width * viewSize.width + viewSize.height * viewSize.height).toFloat())
        triggerCount++
    }
}

@Composable
fun rememberRevealController() = remember { RevealController() }

@Composable
fun ThemeRevealOverlay(
    controller: RevealController,
    onMidpoint: () -> Unit
) {
    var size by remember { mutableStateOf(IntSize.Zero) }
    val animRadius = remember { Animatable(0f) }
    var isDone by remember { mutableStateOf(true) }

    LaunchedEffect(controller.triggerCount) {
        if (controller.triggerCount > 0) {
            isDone = false
            animRadius.snapTo(0f)
            animRadius.animateTo(controller.maxRadius * 0.5f, tween(200))
            onMidpoint()
            animRadius.animateTo(controller.maxRadius, tween(300))
            isDone = true
        }
    }

    val r = animRadius.value
    if (!isDone && r > 1f) {
        val targetColor = if (controller.targetDark) Color(0xFF1C1B1F) else Color(0xFFFFFBFE)

        Canvas(
            modifier = Modifier
                .fillMaxSize()
                .onSizeChanged { size = it }
        ) {
            drawRect(
                color = targetColor,
                size = this.size
            )
            drawCircle(
                color = targetColor.copy(alpha = 0f),
                radius = r,
                center = controller.center,
                blendMode = BlendMode.Clear
            )
        }
    }
}
