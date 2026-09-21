package com.example.kalasetu.features.auth

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.text.input.VisualTransformation
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.kalasetu.theme.SubtitleGray

@Composable
fun AuthLogoHeader() {
    Column(horizontalAlignment = Alignment.CenterHorizontally) {
        Text("KalaSetu", fontSize = 36.sp, fontWeight = FontWeight.Bold)
    }
}

@Composable
fun AuthTextField(
    value: String,
    onValueChange: (String) -> Unit,
    placeholder: String,
    isPassword: Boolean = false,
    keyboardOptions: KeyboardOptions = KeyboardOptions.Default
) {
    OutlinedTextField(
        value = value,
        onValueChange = onValueChange,
        placeholder = {
            Text(
                placeholder,
                modifier = Modifier.fillMaxWidth(),
                textAlign = TextAlign.Center,
            )
        },
        singleLine = true,
        shape = RoundedCornerShape(8.dp),

        keyboardOptions = keyboardOptions,

        visualTransformation = if (isPassword)
            PasswordVisualTransformation()
        else
            VisualTransformation.None,

        modifier = Modifier.fillMaxWidth(),
    )
}

@Composable
fun OrDivider() {
    Row(verticalAlignment = Alignment.CenterVertically) {
        HorizontalDivider(modifier = Modifier.weight(1f))
        Text(
            text = "OR",
            modifier = Modifier.padding(horizontal = 16.dp),
            color = SubtitleGray,
        )
        HorizontalDivider(modifier = Modifier.weight(1f))
    }
}
