package com.example.kalasetu.features.marketplace

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update

class MarketplaceViewModel(
    private val repository: ProductRepository = LocalProductRepository(),
) : ViewModel() {

    private val _products = MutableStateFlow(repository.products)
    val products: StateFlow<List<Product>> = _products.asStateFlow()

    private val _favouriteIds = MutableStateFlow(repository.favouriteIds())
    val favouriteIds: StateFlow<Set<String>> = _favouriteIds.asStateFlow()

    private val _selectedCategory = MutableStateFlow(MarketplaceCategories.first())
    val selectedCategory: StateFlow<String> = _selectedCategory.asStateFlow()

    private val _query = MutableStateFlow("")
    val query: StateFlow<String> = _query.asStateFlow()

    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()

    val filteredProducts: StateFlow<List<Product>> =
        combine(_products, _selectedCategory, _query) { products, category, rawQuery ->
            val q = rawQuery.trim()
            products.filter { product ->
                val matchesCategory =
                    category == "All" ||
                        product.category == category
                val matchesQuery =
                    q.isEmpty() ||
                        product.name.contains(q, ignoreCase = true) ||
                        product.sellerName.contains(q, ignoreCase = true) ||
                        product.category.contains(q, ignoreCase = true)
                matchesCategory && matchesQuery
            }
        }.stateIn(
            scope = viewModelScope,
            started = SharingStarted.Eagerly,
            initialValue = emptyList(),
        )

    fun productById(id: String): Product? =
        repository.productById(id)

    fun isFavourite(id: String): Boolean =
        id in _favouriteIds.value

    fun toggleFavourite(id: String) {
        repository.toggleFavourite(id)
        _favouriteIds.update { repository.favouriteIds() }
    }

    fun setCategory(category: String) {
        _selectedCategory.value = category
    }

    fun setQuery(query: String) {
        _query.value = query
    }
}