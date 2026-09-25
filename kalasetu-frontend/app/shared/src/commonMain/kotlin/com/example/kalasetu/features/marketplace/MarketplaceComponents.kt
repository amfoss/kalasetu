package com.example.kalasetu.features.marketplace

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Favorite
import androidx.compose.material.icons.filled.FavoriteBorder
import androidx.compose.material.icons.filled.Star
import androidx.compose.material3.Icon
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

@Composable
fun ProductRatingRow(
    rating: Double,
    starSize: Dp = 16.dp,
    textSize: Int = 13,
    textColor: Color = TextGray,
) {
    Row(verticalAlignment = Alignment.CenterVertically) {
        Icon(
            imageVector = Icons.Filled.Star,
            contentDescription = "Rating",
            tint = StarYellow,
            modifier = Modifier.size(starSize),
        )
        Spacer(Modifier.width(4.dp))
        Text(
            text = rating.toString(),
            fontSize = textSize.sp,
            fontWeight = FontWeight.Medium,
            color = textColor,
        )
    }
}

@Composable
fun FavouriteHeart(
    isFavourite: Boolean,
    onClick: () -> Unit,
    size: Dp = 20.dp,
    tint: Color = TextGray,
) {
    Icon(
        imageVector = if (isFavourite) Icons.Filled.Favorite else Icons.Filled.FavoriteBorder,
        contentDescription = if (isFavourite) "Remove from favourites" else "Add to favourites",
        tint = if (isFavourite) FavRed else tint,
        modifier = Modifier
            .size(size)
            .clip(CircleShape)
            .clickable(onClick = onClick),
    )
}