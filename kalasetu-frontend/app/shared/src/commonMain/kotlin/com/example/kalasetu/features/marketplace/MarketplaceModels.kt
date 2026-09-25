package com.example.kalasetu.features.marketplace

data class Product(
    val id: String,
    val name: String,
    val sellerName: String,
    val description: String,
    val category: String,
    val price: Double,
    val rating: Double,
    val imageUrl: String? = null,
    val imageBytes: ByteArray? = null,
) {
    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (other == null || this::class != other::class) return false

        other as Product

        if (id != other.id) return false
        if (name != other.name) return false
        if (sellerName != other.sellerName) return false
        if (description != other.description) return false
        if (category != other.category) return false
        if (price != other.price) return false
        if (rating != other.rating) return false
        if (imageUrl != other.imageUrl) return false
        if (imageBytes != null) {
            if (other.imageBytes == null) return false
            if (!imageBytes.contentEquals(other.imageBytes)) return false
        } else if (other.imageBytes != null) return false
        return true
    }

    override fun hashCode(): Int {
        var result = id.hashCode()
        result = 31 * result + name.hashCode()
        result = 31 * result + sellerName.hashCode()
        result = 31 * result + description.hashCode()
        result = 31 * result + category.hashCode()
        result = 31 * result + price.hashCode()
        result = 31 * result + rating.hashCode()
        result = 31 * result + (imageUrl?.hashCode() ?: 0)
        result = 31 * result + (imageBytes?.contentHashCode() ?: 0)
        return result
    }
}

val MarketplaceCategories = listOf(
    "All",
    "Potteries",
    "Paintings",
    "Sculptures",
    "Textiles",
    "Jewellery",
    "Woodcraft",
    "Metalwork",
    "Calligraphy",
)