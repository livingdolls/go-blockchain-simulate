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
import androidx.compose.foundation.shape.CircleShape
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
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.takahashi.yutecoin.data.dto.NotificationItem
import org.koin.androidx.compose.koinViewModel
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

@Composable
fun NotificationScreen(
    viewModel: NotificationViewModel = koinViewModel(),
    onDismiss: () -> Unit = {}
) {
    val uiState by viewModel.state.collectAsStateWithLifecycle()

    LaunchedEffect(Unit) { viewModel.load() }

    Column(modifier = Modifier.fillMaxSize().padding(16.dp)) {
        Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
            Text("Notifications", style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold)
            Row {
                if (uiState.unreadCount > 0) {
                    TextButton(onClick = { viewModel.markAllAsRead() }) { Text("Read all", style = MaterialTheme.typography.labelSmall) }
                }
                if (onDismiss != {}) {
                    TextButton(onClick = onDismiss) { Text("Close", style = MaterialTheme.typography.labelSmall) }
                }
            }
        }

        Spacer(Modifier.height(8.dp))

        if (uiState.isLoading) {
            Box(Modifier.fillMaxWidth().height(200.dp), Alignment.Center) { CircularProgressIndicator(Modifier.size(28.dp)) }
        } else if (uiState.notifications.isEmpty()) {
            Box(Modifier.fillMaxWidth().weight(1f), Alignment.Center) {
                Text("No notifications", style = MaterialTheme.typography.bodyMedium, color = MaterialTheme.colorScheme.onSurfaceVariant)
            }
        } else {
            LazyColumn(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                items(uiState.notifications) { notif ->
                    NotificationCard(notif, onMarkRead = { viewModel.markAsRead(notif.id) }, onDelete = { viewModel.delete(notif.id) })
                }
                item { Spacer(Modifier.height(16.dp)) }
            }
        }
    }
}

@Composable
fun NotificationBadge(
    unreadCount: Int,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Box(modifier = modifier) {
        TextButton(onClick = onClick) {
            Text("\uD83D\uDD14", fontSize = 18.sp)
        }
        if (unreadCount > 0) {
            Box(
                modifier = Modifier
                    .align(Alignment.TopEnd)
                    .padding(top = 2.dp, end = 2.dp)
                    .size(16.dp)
                    .clip(CircleShape)
                    .background(Color(0xFFFF5722)),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    if (unreadCount > 9) "9+" else "$unreadCount",
                    fontSize = 9.sp,
                    color = Color.White,
                    fontWeight = FontWeight.Bold
                )
            }
        }
    }
}

@Composable
private fun NotificationCard(notif: NotificationItem, onMarkRead: () -> Unit, onDelete: () -> Unit) {
    val dateFormat = SimpleDateFormat("HH:mm", Locale.getDefault())
    val bgAlpha = if (notif.isRead) 0.2f else 0.5f

    Card(
        Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(10.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = bgAlpha))
    ) {
        Row(Modifier.padding(10.dp)) {
            if (!notif.isRead) {
                Box(Modifier.padding(top = 6.dp).size(8.dp).clip(CircleShape).background(MaterialTheme.colorScheme.primary))
                Spacer(Modifier.width(8.dp))
            }
            Column(Modifier.weight(1f)) {
                Text(notif.title, style = MaterialTheme.typography.labelMedium, fontWeight = if (notif.isRead) FontWeight.Normal else FontWeight.Bold)
                if (!notif.message.isNullOrEmpty()) {
                    Text(notif.message, style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurfaceVariant, maxLines = 2)
                }
                Text(dateFormat.format(Date(notif.createdAt * 1000)), style = MaterialTheme.typography.labelSmall, color = MaterialTheme.colorScheme.onSurfaceVariant)
            }
            Column {
                if (!notif.isRead) {
                    TextButton(onClick = onMarkRead) { Text("Read", style = MaterialTheme.typography.labelSmall, fontSize = 10.sp) }
                }
                TextButton(onClick = onDelete) { Text("Del", style = MaterialTheme.typography.labelSmall, fontSize = 10.sp, color = MaterialTheme.colorScheme.error) }
            }
        }
    }
}
