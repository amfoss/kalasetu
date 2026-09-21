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
import androidx.compose.ui.text.LinkAnnotation
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.TextLinkStyles
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

@Composable
fun AuthLoginScreen(
    onLogin: (String, String) -> Unit,
    onSignUp: () -> Unit,
    onBack: () -> Unit,
) {
    var username by remember { mutableStateOf("") }
    var password by remember { mutableStateOf("") }

    val canLogin = username.isNotBlank() && password.isNotBlank()

    val forgotPhrase = buildAnnotatedString {
        append("Forgot your password? ")
        pushLink(
            LinkAnnotation.Clickable(
                tag = "reset",
                styles = TextLinkStyles(
                    style = SpanStyle(
                        color = MaterialTheme.colorScheme.primary,
                        fontWeight = FontWeight.Bold,
                    )
                ),
                linkInteractionListener = { /* TODO: reset password */ },
            )
        )
        append("Reset")
        pop()
    }

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
                text = "Login to your account",
                fontSize = 32.sp,
                fontWeight = FontWeight.Bold,
                textAlign = TextAlign.Center,
            )

            Spacer(Modifier.height(32.dp))

            AuthTextField(value = username, onValueChange = { username = it }, placeholder = "Email address")
            Spacer(Modifier.height(16.dp))
            AuthTextField(value = password, onValueChange = { password = it }, placeholder = "Password", isPassword = true)

            Spacer(Modifier.height(24.dp))

            Button(
                onClick = {
                    onLogin(
                        username.trim(),
                        password
                    )
                },
                enabled = canLogin,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(56.dp),
                shape = RoundedCornerShape(8.dp),
            ) {
                Text("Login", fontSize = 16.sp)
            }

            Spacer(Modifier.height(16.dp))

            Text(
                text = forgotPhrase,
                fontSize = 14.sp,
            )

            Spacer(Modifier.height(16.dp))

            OrDivider()

            Spacer(Modifier.height(24.dp))

            TextButton(onClick = onSignUp) {
                Text("Create new account", fontSize = 16.sp)
            }
        }
    }
}
