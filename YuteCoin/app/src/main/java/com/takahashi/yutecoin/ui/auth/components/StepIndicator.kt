package com.takahashi.yutecoin.ui.auth.components

import androidx.compose.animation.animateColorAsState
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.unit.dp

@Composable
fun StepIndicator(
    currentStep: Int,
    totalSteps: Int = 3,
    modifier: Modifier = Modifier
) {
    Row(
        modifier = modifier,
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.Center
    ) {
        for (i in 0 until totalSteps) {
            val isActive = i <= currentStep
            val color by animateColorAsState(
                if (isActive) MaterialTheme.colorScheme.primary
                else MaterialTheme.colorScheme.outlineVariant,
                label = "stepColor"
            )
            val bgColor by animateColorAsState(
                if (isActive) MaterialTheme.colorScheme.primaryContainer
                else MaterialTheme.colorScheme.surfaceVariant,
                label = "stepBg"
            )

            Box(
                modifier = Modifier
                    .size(36.dp)
                    .clip(CircleShape)
                    .background(bgColor)
                    .border(2.dp, color, CircleShape),
                contentAlignment = Alignment.Center
            ) {
                androidx.compose.material3.Text(
                    text = "${i + 1}",
                    style = MaterialTheme.typography.labelMedium,
                    color = color
                )
            }

            if (i < totalSteps - 1) {
                Spacer(modifier = Modifier.width(12.dp))
                Box(
                    modifier = Modifier
                        .width(32.dp)
                        .size(2.dp, 2.dp)
                        .background(if (i < currentStep) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.outlineVariant)
                )
                Spacer(modifier = Modifier.width(12.dp))
            }
        }
    }
}
