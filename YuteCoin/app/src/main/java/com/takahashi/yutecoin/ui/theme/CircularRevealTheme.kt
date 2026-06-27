package com.takahashi.yutecoin.ui.theme

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clipToBounds
import androidx.compose.ui.draw.drawWithContent
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.geometry.Size
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.drawscope.clipPath
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.unit.IntSize
import kotlin.math.sqrt

@Composable
fun CircularRevealTheme(
    isDark: Boolean,
    revealState: RevealState,
    modifier: Modifier = Modifier,
    content: @Composable () -> Unit
) {
    var containerSize by remember { mutableStateOf(IntSize.Zero) }

    YuteCoinTheme(
        darkTheme = isDark,
        content = {
            Box(
                modifier = modifier
                    .fillMaxSize()
                    .onSizeChanged { containerSize = it }
            ) {
                content()
            }
        }
    )

    if (revealState.isRevealing && containerSize != IntSize.Zero) {
        val animatedRadius by animateFloatAsState(
            targetValue = sqrt((containerSize.width * containerSize.width + containerSize.height * containerSize.height).toFloat()),
            animationSpec = tween(500),
            label = "reveal",
            finishedListener = { revealState.finishReveal() }
        )

        if (animatedRadius > 1f) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .drawWithContent {
                        val path = Path().apply {
                            addOval(
                                androidx.compose.ui.geometry.Rect(
                                    revealState.center.x - animatedRadius,
                                    revealState.center.y - animatedRadius,
                                    revealState.center.x + animatedRadius,
                                    revealState.center.y + animatedRadius
                                )
                            )
                        }
                        clipPath(path) {
                            this@drawWithContent.drawContent()
                        }
                    }
            ) {
                YuteCoinTheme(
                    darkTheme = !isDark,
                    content = { content() }
                )
            }
        }
    }
}
