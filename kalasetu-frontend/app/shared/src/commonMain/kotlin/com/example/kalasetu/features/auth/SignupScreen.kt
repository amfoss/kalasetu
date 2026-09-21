package com.example.kalasetu.features.auth

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.text.input.VisualTransformation
@Composable
fun AuthSignupScreen(
    onSignUp: (String, String, String) -> Unit,
    onLogin: () -> Unit,
    onBack: () -> Unit,
) {
    var name by remember { mutableStateOf("") }
    var email by remember { mutableStateOf("") }
    var password by remember { mutableStateOf("") }
    var confirmPassword by remember { mutableStateOf("") }

    val isEmailValid = email.matches(Regex("^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$"))
    val passwordsMatch = password == confirmPassword
    val canSignUp = name.isNotBlank() && isEmailValid && password.isNotBlank() && passwordsMatch

    Box(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState()),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            IconButton(onClick = onBack, modifier = Modifier.align(Alignment.Start)) {
                Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "Back")
            }

            Spacer(Modifier.height(16.dp))

            AuthLogoHeader()

            Spacer(Modifier.height(32.dp))

            Text(
                text = "Create your account",
                fontSize = 32.sp,
                fontWeight = FontWeight.Bold,
                textAlign = TextAlign.Center,
            )

            Spacer(Modifier.height(32.dp))
            AuthTextField(value = name, onValueChange = { name = it }, placeholder = "Name")
            Spacer(Modifier.height(16.dp))
            AuthTextField(
                value = email,
                onValueChange = { email = it },
                placeholder = "Email",
                keyboardOptions = KeyboardOptions(
                    keyboardType = KeyboardType.Email
                )
            )
            if (email.isNotBlank() && !isEmailValid) {
                Text(
                    text = "Enter a valid email address",
                    color = MaterialTheme.colorScheme.error,
                    fontSize = 12.sp,
                    modifier = Modifier.padding(top = 4.dp, start = 4.dp),
                )
            }
            Spacer(Modifier.height(16.dp))
            AuthTextField(value = password, onValueChange = { password = it }, placeholder = "Password", isPassword = true)
            Spacer(Modifier.height(16.dp))
            AuthTextField(value = confirmPassword, onValueChange = { confirmPassword = it }, placeholder = "Confirm Password", isPassword = true)
            if (confirmPassword.isNotBlank() && !passwordsMatch) {
                Text(
                    text = "Passwords do not match",
                    color = MaterialTheme.colorScheme.error,
                    fontSize = 12.sp,
                    modifier = Modifier.padding(top = 4.dp, start = 4.dp),
                )
            }

            Spacer(Modifier.height(24.dp))

            Button(
                onClick = {
                    onSignUp(
                        name.trim(),
                        email.trim(),
                        password
                    )
                },
                enabled = canSignUp,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(56.dp),
                shape = RoundedCornerShape(8.dp),
            ) {
                Text("Sign Up", fontSize = 16.sp)
            }

            Spacer(Modifier.height(24.dp))

            OrDivider()

            Spacer(Modifier.height(24.dp))

            TextButton(onClick = onLogin) {
                Text("Login to your account", fontSize = 16.sp)
            }
        }
    }
}
