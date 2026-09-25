package com.example.kalasetu.features.auth

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.ArrowForward
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.kalasetu.theme.SelectedPurple
import com.example.kalasetu.theme.SubtitleGray

@Composable
fun AuthEmailScreen(
    onContinue: (String) -> Unit,
    onBack: () -> Unit,
) {
    var email by remember { mutableStateOf("") }
    var error by remember { mutableStateOf<String?>(null) }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp)
    ) {
        Column(
            modifier = Modifier.fillMaxSize()
        ) {
            IconButton(
                onClick = onBack,
                modifier = Modifier.align(Alignment.Start)
            ) {
                Icon(
                    Icons.AutoMirrored.Filled.ArrowBack,
                    contentDescription = "Back"
                )
            }

            Spacer(Modifier.height(32.dp))

            Text(
                text = "Create your account",
                fontSize = 32.sp,
                style = MaterialTheme.typography.headlineMedium
            )

            Spacer(Modifier.height(8.dp))

            Text(
                text = "Enter your email to get a verification code.",
                color = SubtitleGray,
                fontSize = 16.sp
            )

            Spacer(Modifier.height(40.dp))

            OutlinedTextField(
                value = email,
                onValueChange = {
                    email = it
                    error = null
                },
                modifier = Modifier.fillMaxWidth(),
                label = { Text("Email") },
                singleLine = true,
                keyboardOptions = KeyboardOptions(
                    keyboardType = KeyboardType.Email
                ),
                isError = error != null
            )

            if (error != null) {
                Spacer(Modifier.height(8.dp))

                Text(
                    text = error!!,
                    color = MaterialTheme.colorScheme.error,
                    fontSize = 14.sp
                )
            }
        }

        FloatingActionButton(
            onClick = {
                val cleanEmail = email.trim()

                when {
                    cleanEmail.isBlank() -> {
                        error = "Enter your email"
                    }

                    !cleanEmail.contains("@") -> {
                        error = "Enter a valid email"
                    }

                    else -> {
                        onContinue(cleanEmail)
                    }
                }
            },
            modifier = Modifier
                .align(Alignment.BottomEnd)
                .size(56.dp),
            containerColor = SelectedPurple
        ) {
            Icon(
                Icons.AutoMirrored.Filled.ArrowForward,
                contentDescription = "Continue"
            )
        }
    }
}