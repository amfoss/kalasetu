package com.example.kalasetu.features.auth

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.ArrowForward
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.kalasetu.theme.SelectedPurple
import com.example.kalasetu.theme.SubtitleGray

@Composable
fun AuthSignupDetailsScreen(
    email: String,
    onContinue: (String, String) -> Unit,
    onBack: () -> Unit,
) {
    var name by remember { mutableStateOf("") }
    var password by remember { mutableStateOf("") }
    var confirmPassword by remember { mutableStateOf("") }
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

            Spacer(Modifier.height(24.dp))

            Text(
                text = "Complete your account",
                fontSize = 32.sp,
                style = MaterialTheme.typography.headlineMedium
            )

            Spacer(Modifier.height(8.dp))

            Text(
                text = email,
                color = SubtitleGray,
                fontSize = 15.sp
            )

            Spacer(Modifier.height(32.dp))

            OutlinedTextField(
                value = name,
                onValueChange = {
                    name = it
                    error = null
                },
                modifier = Modifier.fillMaxWidth(),
                label = { Text("Name") },
                singleLine = true
            )

            Spacer(Modifier.height(16.dp))

            OutlinedTextField(
                value = password,
                onValueChange = {
                    password = it
                    error = null
                },
                modifier = Modifier.fillMaxWidth(),
                label = { Text("Password") },
                singleLine = true,
                visualTransformation = PasswordVisualTransformation()
            )

            Spacer(Modifier.height(16.dp))

            OutlinedTextField(
                value = confirmPassword,
                onValueChange = {
                    confirmPassword = it
                    error = null
                },
                modifier = Modifier.fillMaxWidth(),
                label = { Text("Confirm Password") },
                singleLine = true,
                visualTransformation = PasswordVisualTransformation()
            )

            if (error != null) {
                Spacer(Modifier.height(12.dp))

                Text(
                    text = error!!,
                    color = MaterialTheme.colorScheme.error,
                    fontSize = 14.sp
                )
            }
        }

        FloatingActionButton(
            onClick = {
                when {
                    name.isBlank() ->
                        error = "Enter your name"

                    password.length < 6 ->
                        error = "Password must be at least 6 characters"

                    password != confirmPassword ->
                        error = "Passwords do not match"

                    else ->
                        onContinue(name.trim(), password)
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