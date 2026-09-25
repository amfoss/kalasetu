package com.example.kalasetu.features.marketplace

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.kalasetu.features.feed.KalaBottomNav
import com.example.kalasetu.features.feed.KalaTopBar

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun MarketplaceScreen(
    viewModel: MarketplaceViewModel,
    avatarUrl: String?,
    avatarBytes: ByteArray?,
    userName: String?,
    onProductClick: (String) -> Unit,
    onProfileClick: () -> Unit,
    onMenuClick: () -> Unit,
    onHomeClick: () -> Unit,
    onEventsClick: () -> Unit,
    onStoreClick: () -> Unit,
) {
    val products by viewModel.filteredProducts.collectAsState()
    val favouriteIds by viewModel.favouriteIds.collectAsState()
    val selectedCategory by viewModel.selectedCategory.collectAsState()
    val query by viewModel.query.collectAsState()

    Scaffold(
        topBar = {
            KalaTopBar(
                avatarUrl = avatarUrl,
                avatarBytes = avatarBytes,
                userName = userName,
                onProfileClick = onProfileClick,
                onMenuClick = onMenuClick,
            )
        },
        bottomBar = {
            KalaBottomNav(
                selectedIndex = 2,
                onStoreClick = onStoreClick,
                onEventsClick = onEventsClick,
                onHomeClick = onHomeClick,
                onProfileClick = onProfileClick,
            )
        },
        containerColor = CardWhite,
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding),
        ) {
            //  Search bar in the main marketplace screen
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp)
                    .padding(top = 8.dp, bottom = 12.dp)
                    .height(50.dp)
                    .clip(RoundedCornerShape(12.dp))
                    .background(CardWhite)
                    .border(1.dp, BorderGray, RoundedCornerShape(12.dp))
                    .padding(horizontal = 16.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Icon(
                    imageVector = Icons.Filled.Search,
                    contentDescription = "Search",
                    tint = TextDark,
                    modifier = Modifier.size(22.dp),
                )
                Spacer(Modifier.width(10.dp))
                BasicTextField(
                    value = query,
                    onValueChange = viewModel::setQuery,
                    singleLine = true,
                    textStyle = TextStyle(
                        fontSize = 15.sp,
                        color = TextDark,
                    ),
                    cursorBrush = SolidColor(MarketPurple),
                    modifier = Modifier.weight(1f),
                    decorationBox = { innerTextField ->
                        if (query.isEmpty()) {
                            Text(
                                text = "Search",
                                fontSize = 15.sp,
                                color = TextGray,
                            )
                        }
                        innerTextField()
                    },
                )
            }

            // the Category Labels
            LazyRow(
                modifier = Modifier.fillMaxWidth(),
                contentPadding = PaddingValues(horizontal = 16.dp),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                items(MarketplaceCategories) { category ->
                    CategoryChip(
                        label = category,
                        isSelected = category == selectedCategory,
                        onClick = { viewModel.setCategory(category) },
                    )
                }
            }

            Spacer(Modifier.height(12.dp))

            // product cards grid
            if (products.isEmpty()) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center,
                ) {
                    Text("No products found", fontSize = 14.sp, color = TextGray)
                }
            } else {
                LazyVerticalGrid(
                    columns = GridCells.Adaptive(minSize = 150.dp),
                    modifier = Modifier.fillMaxSize(),
                    contentPadding = PaddingValues(horizontal = 16.dp, vertical = 4.dp),
                    horizontalArrangement = Arrangement.spacedBy(12.dp),
                    verticalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    items(products, key = { it.id }) { product ->
                        ProductCard(
                            product = product,
                            isFavourite = product.id in favouriteIds,
                            onFavouriteClick = { viewModel.toggleFavourite(product.id) },
                            onClick = { onProductClick(product.id) },
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun CategoryChip(
    label: String,
    isSelected: Boolean,
    onClick: () -> Unit,
) {
    Box(
        modifier = Modifier
            .clip(RoundedCornerShape(50))
            .background(
                if (isSelected) MarketPurple else LightPurpleBg,
            )
            .clickable(onClick = onClick)
            .padding(horizontal = 18.dp, vertical = 8.dp),
    ) {
        Text(
            text = label,
            fontSize = 14.sp,
            fontWeight = if (isSelected) FontWeight.SemiBold else FontWeight.Normal,
            color = if (isSelected) Color.White else TextDark,
        )
    }
}

@Composable
private fun ProductCard(
    product: Product,
    isFavourite: Boolean,
    onFavouriteClick: () -> Unit,
    onClick: () -> Unit,
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = CardWhite),
        border = BorderStroke(1.dp, BorderGray),
        elevation = CardDefaults.cardElevation(2.dp),
    ) {
        Column {

            // ─── Product image (17:20) ───
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
                            fontSize = 30.sp,
                            fontWeight = FontWeight.Bold,
                            color = MarketPurple,
                        )
                    }
                }
            }


            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(12.dp),
            ) {
                Text(
                    text = product.name,
                    fontSize = 15.sp,
                    fontWeight = FontWeight.SemiBold,
                    color = TextDark,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                Spacer(Modifier.height(4.dp))
                Text(
                    text = product.sellerName,
                    fontSize = 12.sp,
                    color = TextGray,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                Spacer(Modifier.height(8.dp))
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    ProductRatingRow(rating = product.rating)
                    Spacer(Modifier.weight(1f))
                    FavouriteHeart(
                        isFavourite = isFavourite,
                        onClick = onFavouriteClick,
                    )
                }
            }
        }
    }
}