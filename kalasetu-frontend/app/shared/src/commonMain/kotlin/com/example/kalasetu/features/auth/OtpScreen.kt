package com.example.kalasetu.features.auth

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.automirrored.filled.ArrowForward
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import kotlinx.coroutines.delay
import com.example.kalasetu.theme.SelectedPurple
import com.example.kalasetu.theme.SubtitleGray
import com.example.kalasetu.theme.UnselectedBorder

@Composable
fun AuthOtpScreen(
    email: String,
    onVerify: (String) -> Unit,
    onResend: () -> Unit,
    onLogin: () -> Unit,
    onBack: () -> Unit,

) {
    var otp by remember { mutableStateOf("") }
    val focusRequester = remember { FocusRequester() }
    var remainingSeconds by remember { mutableStateOf(60) }

    LaunchedEffect(Unit) {
        focusRequester.requestFocus()
    }

    LaunchedEffect(remainingSeconds) {
        while (remainingSeconds > 0) {
            delay(1000)
            remainingSeconds--
        }
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
    ) {
        Column(
            modifier = Modifier.fillMaxSize(),
        ) {

            IconButton(
                onClick = onBack,
                modifier = Modifier.align(Alignment.Start),
            ) {
                Icon(
                    Icons.AutoMirrored.Filled.ArrowBack,
                    contentDescription = "Back",
                )
            }

            Spacer(Modifier.height(24.dp))

            Text(
                text = "6-digit Code",
                fontSize = 32.sp,
                fontWeight = FontWeight.Bold,
                modifier = Modifier.fillMaxWidth(),
                textAlign = TextAlign.Start,
            )

            Spacer(Modifier.height(8.dp))

            Text(
                text = "Enter the 6-digit code sent to\n$email",
                fontSize = 16.sp,
                textAlign = TextAlign.Start,
                color = SubtitleGray,
                modifier = Modifier.fillMaxWidth(),
            )

            Spacer(Modifier.height(48.dp))

            Box(
                modifier = Modifier.fillMaxWidth(),
                contentAlignment = Alignment.CenterStart,
            ) {
                Row(
                    horizontalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    repeat(6) { index ->
                        val char = otp.getOrNull(index)?.toString() ?: ""
                        val isActive = otp.length == index

                        Box(
                            modifier = Modifier
                                .size(56.dp)
                                .clip(RoundedCornerShape(12.dp))
                                .background(
                                    if (isActive)
                                        SelectedPurple.copy(alpha = 0.08f)
                                    else
                                        MaterialTheme.colorScheme.surfaceVariant,
                                )
                                .border(
                                    width = 1.dp,
                                    color = if (isActive)
                                        SelectedPurple
                                    else
                                        UnselectedBorder,
                                    shape = RoundedCornerShape(12.dp),
                                ),
                            contentAlignment = Alignment.Center,
                        ) {
                            Text(
                                text = char,
                                fontSize = 24.sp,
                                fontWeight = FontWeight.Bold,
                            )
                        }
                    }
                }

                BasicTextField(
                    value = otp,
                    onValueChange = {
                        otp = it.filter(Char::isDigit).take(6)
                    },
                    modifier = Modifier
                        .matchParentSize()
                        .focusRequester(focusRequester)
                        .alpha(0f),
                    singleLine = true,
                    textStyle = TextStyle(
                        color = Color.Transparent,
                        fontSize = 24.sp,
                    ),
                    keyboardOptions = KeyboardOptions(
                        keyboardType = KeyboardType.Number,
                    ),
                )
            }

            Spacer(Modifier.height(32.dp))

            if (remainingSeconds > 0) {
                Text(
                    text = "Resend code in 00:${remainingSeconds.toString().padStart(2, '0')}",
                    fontSize = 14.sp,
                    color = SubtitleGray,
                )
            } else {
                TextButton(
                    onClick = {
                        remainingSeconds = 60
                        onResend()},
                    contentPadding = PaddingValues(0.dp),
                ) {
                    Text(
                        text = "Didn't receive the code? Resend",
                        fontSize = 14.sp,
                        color = SelectedPurple,
                    )
                }
            }


            TextButton(
                onClick = onLogin,
                contentPadding = PaddingValues(0.dp)
            ) {
                Text(
                    text = "Already have an account? Login",
                    fontSize = 16.sp
                )
            }
        }

        IconButton(
            onClick = {
                if (otp.length == 6) {
                    onVerify(otp)
                }
            },
            modifier = Modifier
                .align(Alignment.BottomEnd)
                .size(56.dp)
                .clip(RoundedCornerShape(28.dp))
                .background(
                    if (otp.length == 6)
                        MaterialTheme.colorScheme.primary
                    else
                        MaterialTheme.colorScheme.surfaceVariant,
                ),
        ) {
            Icon(
                imageVector = Icons.AutoMirrored.Filled.ArrowForward,
                contentDescription = "Next",
                tint = if (otp.length == 6)
                    MaterialTheme.colorScheme.onPrimary
                else
                    MaterialTheme.colorScheme.onSurfaceVariant,
            )
        }
    }
}