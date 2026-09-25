package com.example.kalasetu.features.marketplace

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Favorite
import androidx.compose.material.icons.filled.FavoriteBorder
import androidx.compose.material.icons.filled.Share
import androidx.compose.material.icons.filled.Star
import androidx.compose.material.icons.filled.StarBorder
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.kalasetu.features.feed.KalaTopBar
import com.example.kalasetu.navigation.BackHandler

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ProductOverviewScreen(
    productId: String,
    viewModel: MarketplaceViewModel,
    avatarUrl: String?,
    avatarBytes: ByteArray?,
    userName: String?,
    onBack: () -> Unit,
    onProfileClick: () -> Unit,
) {
    val products by viewModel.products.collectAsState()
    val favouriteIds by viewModel.favouriteIds.collectAsState()
    val product = products.firstOrNull { it.id == productId }

    var userRating by remember { mutableIntStateOf(0) }

    BackHandler(onBack = onBack)

    Scaffold(
        topBar = {
            KalaTopBar(
                avatarUrl = avatarUrl,
                avatarBytes = avatarBytes,
                userName = userName,
                title = product?.name ?: "Product",
                onBack = onBack,
                onProfileClick = onProfileClick,
                onMenuClick = onBack,
            )
        },
        containerColor = CardWhite,
    ) { padding ->
        if (product == null) {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding),
                contentAlignment = Alignment.Center,
            ) {
                Text("Product not found", color = TextGray)
            }
            return@Scaffold
        }

        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .verticalScroll(rememberScrollState()),
        ) {

            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .aspectRatio(1f),
            ) {
                val imageModel = product.imageUrl ?: product.imageBytes
                if (imageModel != null) {
                    coil3.compose.AsyncImage(
                        model = imageModel,
                        contentDescription = product.name,
                        contentScale = ContentScale.Crop,
                        modifier = Modifier.fillMaxSize(),
                    )
                } else {
                    Box(
                        modifier = Modifier
                            .fillMaxSize()
                            .background(LightPurpleBg),
                        contentAlignment = Alignment.Center,
                    ) {
                        Text(
                            text = product.name.take(2).uppercase(),
                            fontSize = 48.sp,
                            fontWeight = FontWeight.Bold,
                            color = MarketPurple,
                        )
                    }
                }


                Row(
                    modifier = Modifier
                        .align(Alignment.BottomEnd)
                        .padding(end = 12.dp, bottom = 12.dp),
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Box(
                        modifier = Modifier
                            .size(40.dp)
                            .clip(CircleShape)
                            .background(CardWhite)
                            .border(1.dp, DividerGray, CircleShape),
                        contentAlignment = Alignment.Center,
                    ) {
                        Icon(
                            imageVector = Icons.Default.Share,
                            contentDescription = "Share",
                            tint = TextGray,
                            modifier = Modifier.size(18.dp),
                        )
                    }

                    Box(
                        modifier = Modifier
                            .size(40.dp)
                            .clip(CircleShape)
                            .background(CardWhite)
                            .border(1.dp, DividerGray, CircleShape)
                            .clickable { viewModel.toggleFavourite(product.id) },
                        contentAlignment = Alignment.Center,
                    ) {
                        Icon(
                            imageVector = if (product.id in favouriteIds) {
                                Icons.Filled.Favorite
                            } else {
                                Icons.Filled.FavoriteBorder
                            },
                            contentDescription = "Favourite",
                            tint = if (product.id in favouriteIds) FavRed else TextGray,
                            modifier = Modifier.size(20.dp),
                        )
                    }
                }
            }


            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 20.dp)
                    .padding(top = 20.dp),
            ) {
                Text(
                    text = product.name,
                    fontSize = 24.sp,
                    fontWeight = FontWeight.Bold,
                    color = TextDark,
                )

                Spacer(Modifier.height(8.dp))

                ProductRatingRow(rating = product.rating, starSize = 18.dp, textSize = 14)

                Spacer(Modifier.height(20.dp))

                Text(
                    text = product.description,
                    fontSize = 14.sp,
                    color = TextGray,
                    lineHeight = 20.sp,
                )

                // Rating input using stars
                Spacer(Modifier.height(28.dp))

                Text(
                    text = "Rate this product",
                    fontSize = 16.sp,
                    fontWeight = FontWeight.Bold,
                    color = TextDark,
                )

                Spacer(Modifier.height(12.dp))

                Row(
                    horizontalArrangement = Arrangement.spacedBy(6.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    repeat(5) { index ->
                        val isFilled = index < userRating
                        Icon(
                            imageVector = if (isFilled) Icons.Filled.Star else Icons.Filled.StarBorder,
                            contentDescription = "Rate ${index + 1} star${if (index == 0) "" else "s"}",
                            tint = StarYellow,
                            modifier = Modifier
                                .size(36.dp)
                                .clip(CircleShape)
                                .clickable { userRating = if (userRating == index + 1) 0 else index + 1 },
                        )
                    }
                }


                Spacer(Modifier.height(32.dp))

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    OutlinedButton(
                        onClick = { /* cart logic comes later */ },
                        modifier = Modifier
                            .weight(1f)
                            .height(52.dp),
                        shape = RoundedCornerShape(12.dp),
                        border = BorderStroke(1.dp, MarketPurple),
                        colors = ButtonDefaults.outlinedButtonColors(contentColor = MarketPurple),
                    ) {
                        Text("Add to Cart", fontSize = 15.sp, fontWeight = FontWeight.Bold)
                    }

                    Button(
                        onClick = { /* order logic comes later */ },
                        modifier = Modifier
                            .weight(1f)
                            .height(52.dp),
                        shape = RoundedCornerShape(12.dp),
                        colors = ButtonDefaults.buttonColors(containerColor = MarketPurple),
                    ) {
                        Text("Order Now", fontSize = 15.sp, fontWeight = FontWeight.Bold)
                    }
                }

                Spacer(Modifier.height(32.dp))
            }
        }
    }
}