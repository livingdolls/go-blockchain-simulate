package com.takahashi.yutecoin.ui.auth

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.takahashi.yutecoin.ui.auth.components.BackupWarning
import com.takahashi.yutecoin.ui.auth.components.MnemonicDisplay
import com.takahashi.yutecoin.ui.auth.components.StepIndicator
import kotlinx.coroutines.flow.collectLatest
import org.koin.androidx.compose.koinViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun RegisterScreen(
    onNavigateHome: () -> Unit,
    onBack: () -> Unit,
    viewModel: RegisterViewModel = koinViewModel()
) {
    val state by viewModel.state.collectAsStateWithLifecycle()

    LaunchedEffect(Unit) {
        viewModel.events.collectLatest { event ->
            when (event) {
                is RegisterEvent.NavigateHome -> onNavigateHome()
                is RegisterEvent.ShowError -> {}
            }
        }
    }

    Scaffold(
        modifier = Modifier.fillMaxSize(),
        topBar = {
            if (state.currentStep > 0) {
                TopAppBar(
                    title = {},
                    navigationIcon = {
                        TextButton(onClick = viewModel::goBack) {
                            Text("\u2190 Back")
                        }
                    }
                )
            }
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .padding(padding)
                .padding(horizontal = 24.dp)
                .verticalScroll(rememberScrollState()),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Spacer(modifier = Modifier.height(24.dp))

            StepIndicator(currentStep = state.currentStep, totalSteps = 3)
            Spacer(modifier = Modifier.height(24.dp))

            when (state.currentStep) {
                0 -> PasswordStep(
                    state = state,
                    onPasswordChange = viewModel::onPasswordChange,
                    onConfirmPasswordChange = viewModel::onConfirmPasswordChange,
                    onNext = viewModel::goToMnemonicStep
                )

                1 -> MnemonicStep(
                    state = state,
                    onGenerate = viewModel::generateWallet,
                    onBackupConfirmed = viewModel::onBackupConfirmed,
                    onNext = viewModel::goToUsernameStep
                )

                2 -> UsernameStep(
                    state = state,
                    onUsernameChange = viewModel::onUsernameChange,
                    onSubmit = viewModel::submitRegistration
                )
            }

            Spacer(modifier = Modifier.height(24.dp))
        }
    }
}

@Composable
private fun PasswordStep(
    state: RegisterUiState,
    onPasswordChange: (String) -> Unit,
    onConfirmPasswordChange: (String) -> Unit,
    onNext: () -> Unit
) {
    Text(
        text = "Create Backup Password",
        style = MaterialTheme.typography.headlineSmall,
        textAlign = TextAlign.Center
    )
    Spacer(modifier = Modifier.height(8.dp))
    Text(
        text = "This password encrypts your wallet backup. It is never sent to the server.",
        style = MaterialTheme.typography.bodyMedium,
        color = MaterialTheme.colorScheme.onSurfaceVariant,
        textAlign = TextAlign.Center
    )
    Spacer(modifier = Modifier.height(24.dp))

    OutlinedTextField(
        value = state.password,
        onValueChange = onPasswordChange,
        label = { Text("Password") },
        visualTransformation = PasswordVisualTransformation(),
        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Password),
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(12.dp),
        singleLine = true
    )
    Spacer(modifier = Modifier.height(12.dp))

    OutlinedTextField(
        value = state.confirmPassword,
        onValueChange = onConfirmPasswordChange,
        label = { Text("Confirm Password") },
        visualTransformation = PasswordVisualTransformation(),
        keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Password),
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(12.dp),
        singleLine = true,
        isError = state.passwordError != null,
        supportingText = {
            if (state.passwordError != null) {
                Text(text = state.passwordError!!, color = MaterialTheme.colorScheme.error)
            }
        }
    )
    Spacer(modifier = Modifier.height(8.dp))

    Text(
        text = "Password must be at least 8 characters",
        style = MaterialTheme.typography.bodySmall,
        color = MaterialTheme.colorScheme.outline
    )

    Spacer(modifier = Modifier.height(32.dp))

    Button(
        onClick = onNext,
        modifier = Modifier
            .fillMaxWidth()
            .height(52.dp),
        shape = RoundedCornerShape(12.dp)
    ) {
        Text("Continue", style = MaterialTheme.typography.titleMedium)
    }
}

@Composable
private fun MnemonicStep(
    state: RegisterUiState,
    onGenerate: () -> Unit,
    onBackupConfirmed: () -> Unit,
    onNext: () -> Unit
) {
    Text(
        text = "Your Wallet",
        style = MaterialTheme.typography.headlineSmall,
        textAlign = TextAlign.Center
    )
    Spacer(modifier = Modifier.height(8.dp))
    Text(
        text = "Save these 12 words in a safe place. They are the only way to recover your wallet.",
        style = MaterialTheme.typography.bodyMedium,
        color = MaterialTheme.colorScheme.onSurfaceVariant,
        textAlign = TextAlign.Center
    )
    Spacer(modifier = Modifier.height(24.dp))

    if (state.mnemonic.isEmpty()) {
        Button(
            onClick = onGenerate,
            modifier = Modifier
                .fillMaxWidth()
                .height(52.dp),
            shape = RoundedCornerShape(12.dp),
            enabled = !state.isGenerating
        ) {
            if (state.isGenerating) {
                CircularProgressIndicator(
                    modifier = Modifier.size(24.dp),
                    color = MaterialTheme.colorScheme.onPrimary,
                    strokeWidth = 2.dp
                )
            } else {
                Text("Generate Wallet", style = MaterialTheme.typography.titleMedium)
            }
        }
    } else {
        Text(
            text = "Address: ${state.address}",
            style = MaterialTheme.typography.bodySmall,
            fontFamily = FontFamily.Monospace,
            color = MaterialTheme.colorScheme.outline,
            textAlign = TextAlign.Center,
            modifier = Modifier.fillMaxWidth()
        )
        Spacer(modifier = Modifier.height(12.dp))

        MnemonicDisplay(mnemonic = state.mnemonic)
        Spacer(modifier = Modifier.height(16.dp))

        OutlinedButton(
            onClick = { /* TODO: keystore JSON download */ },
            modifier = Modifier.fillMaxWidth(),
            shape = RoundedCornerShape(12.dp)
        ) {
            Text("Download Backup JSON")
        }
        Spacer(modifier = Modifier.height(8.dp))

        BackupWarning(
            onBackupConfirmed = state.hasBackedUp,
            onCheckedChange = { onBackupConfirmed() }
        )
        Spacer(modifier = Modifier.height(16.dp))

        Button(
            onClick = onNext,
            modifier = Modifier
                .fillMaxWidth()
                .height(52.dp),
            shape = RoundedCornerShape(12.dp),
            enabled = state.hasBackedUp
        ) {
            Text("Continue", style = MaterialTheme.typography.titleMedium)
        }
    }
}

@Composable
private fun UsernameStep(
    state: RegisterUiState,
    onUsernameChange: (String) -> Unit,
    onSubmit: () -> Unit
) {
    Text(
        text = "Choose Username",
        style = MaterialTheme.typography.headlineSmall,
        textAlign = TextAlign.Center
    )
    Spacer(modifier = Modifier.height(8.dp))
    Text(
        text = "This will be your identity on the network",
        style = MaterialTheme.typography.bodyMedium,
        color = MaterialTheme.colorScheme.onSurfaceVariant,
        textAlign = TextAlign.Center
    )
    Spacer(modifier = Modifier.height(24.dp))

    Text(
        text = "Address: ${state.address}",
        style = MaterialTheme.typography.bodySmall,
        fontFamily = FontFamily.Monospace,
        color = MaterialTheme.colorScheme.outline,
        textAlign = TextAlign.Center,
        modifier = Modifier.fillMaxWidth()
    )
    Spacer(modifier = Modifier.height(16.dp))

    OutlinedTextField(
        value = state.username,
        onValueChange = onUsernameChange,
        label = { Text("Username") },
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(12.dp),
        singleLine = true,
        isError = state.usernameError != null,
        supportingText = {
            if (state.usernameError != null) {
                Text(text = state.usernameError!!, color = MaterialTheme.colorScheme.error)
            }
        },
        enabled = !state.isLoading
    )
    Spacer(modifier = Modifier.height(8.dp))

    if (state.error != null) {
        Text(
            text = state.error!!,
            color = MaterialTheme.colorScheme.error,
            style = MaterialTheme.typography.bodySmall,
            textAlign = TextAlign.Center,
            modifier = Modifier.fillMaxWidth()
        )
        Spacer(modifier = Modifier.height(8.dp))
    }

    Spacer(modifier = Modifier.height(24.dp))

    Button(
        onClick = onSubmit,
        modifier = Modifier
            .fillMaxWidth()
            .height(52.dp),
        shape = RoundedCornerShape(12.dp),
        enabled = !state.isLoading
    ) {
        if (state.isLoading) {
            CircularProgressIndicator(
                modifier = Modifier.size(24.dp),
                color = MaterialTheme.colorScheme.onPrimary,
                strokeWidth = 2.dp
            )
        } else {
            Text("Complete", style = MaterialTheme.typography.titleMedium)
        }
    }
}
