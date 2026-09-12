package com.example.kalasetu.features.marketplace

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

object MarketplaceStore {
    private val _products = MutableStateFlow<List<Product>>(sampleProducts)
    val products: StateFlow<List<Product>> = _products.asStateFlow()

    private val _favouriteIds = MutableStateFlow<Set<String>>(emptySet())
    val favouriteIds: StateFlow<Set<String>> = _favouriteIds.asStateFlow()

    fun productById(id: String): Product? =
        _products.value.firstOrNull { it.id == id }

    fun isFavourite(id: String): Boolean =
        id in _favouriteIds.value

    fun toggleFavourite(id: String) {
        _favouriteIds.value = if (id in _favouriteIds.value) {
            _favouriteIds.value - id
        } else {
            _favouriteIds.value + id
        }
    }
}

private val sampleProducts = listOf(
    Product(
        id = "p1",
        name = "Terracotta Vase",
        sellerName = "Ramesh Crafts",
        description = "Hand-thrown terracotta vase with a warm rustic finish, made using traditional potter's wheel techniques passed down through generations.",
        category = "Potteries",
        price = 1200.0,
        rating = 4.6,
    ),
    Product(
        id = "p2",
        name = "Blue Pottery Plate",
        sellerName = "Meena Arts",
        description = "Delicately hand-painted blue pottery plate with intricate floral motifs, glazed and kiln-fired for a lasting shine.",
        category = "Potteries",
        price = 850.0,
        rating = 4.3,
    ),
    Product(
        id = "p3",
        name = "Madhubani Wall Art",
        sellerName = "Sita Devi Studio",
        description = "Vibrant Madhubani painting depicting folk tales, hand-painted with natural dyes on handmade paper.",
        category = "Paintings",
        price = 3500.0,
        rating = 4.8,
    ),
    Product(
        id = "p4",
        name = "Tanpura Landscape",
        sellerName = "Farhan Arts",
        description = "A serene watercolour landscape capturing the banks of the Ganga at dawn, signed and titled by the artist.",
        category = "Paintings",
        price = 2500.0,
        rating = 4.1,
    ),
    Product(
        id = "p5",
        name = "Bronze Dancing Figure",
        sellerName = "Karthik Bronze Works",
        description = "Solid brass figurine of a classical dancer in motion, cast using the ancient lost-wax technique.",
        category = "Sculptures",
        price = 5400.0,
        rating = 4.7,
    ),
    Product(
        id = "p6",
        name = "Stone Elephant Candle Stand",
        sellerName = "Om Craft House",
        description = "Hand-carved soapstone elephant candle holder with a soft matte finish, ideal for home decor.",
        category = "Sculptures",
        price = 1750.0,
        rating = 4.2,
    ),
    Product(
        id = "p7",
        name = "Kalamkari Sari",
        sellerName = "Weaves of India",
        description = "Hand-printed Kalamkari cotton sari with earthy natural dyes and mythological border art.",
        category = "Textiles",
        price = 4600.0,
        rating = 4.5,
    ),
    Product(
        id = "p8",
        name = "Block Printed Stole",
        sellerName = "Anandi Textiles",
        description = "Lightweight hand-block-printed cotton stole in warm indigo tones with tassel edges.",
        category = "Textiles",
        price = 1290.0,
        rating = 4.0,
    ),
    Product(
        id = "p9",
        name = "Necklace Set",
        sellerName = "Shailu Jewels",
        description = "Traditional necklace set in gold-plated finish, studded with hand-set coloured stones.",
        category = "Jewellery",
        price = 8900.0,
        rating = 4.9,
    ),
    Product(
        id = "p10",
        name = "Silver Anklet Pair",
        sellerName = "Noor Silvers",
        description = "Hand-beaten sterling silver anklets with bell charms, crafted by family silversmiths.",
        category = "Jewellery",
        price = 2300.0,
        rating = 4.4,
    ),
    Product(
        id = "p11",
        name = "Rosewood Serving Bowl",
        sellerName = "Teak & Brush",
        description = "Carved rosewood serving bowl with smooth oiled finish, safe for dry snacks and fruits.",
        category = "Woodcraft",
        price = 1450.0,
        rating = 4.3,
    ),
    Product(
        id = "p12",
        name = "Sandalwood Elephant",
        sellerName = "Arun Woodcrafts",
        description = "Intricately carved sandalwood elephant ornament with natural aromatic wood grain.",
        category = "Woodcraft",
        price = 3200.0,
        rating = 4.6,
    ),
    Product(
        id = "p13",
        name = "Bell Metal Utensil Set",
        sellerName = "Bellmakers of Palakkad",
        description = "Traditional bell metal bowls for water and cooking, hammered and polished to a golden sheen.",
        category = "Metalwork",
        price = 4800.0,
        rating = 4.5,
    ),
    Product(
        id = "p14",
        name = "Dhokra Home Décor",
        sellerName = "Tribal Craft Co-op",
        description = "Hand-crafted dhokra tribal décor piece made with the ancient lost-wax metal casting art.",
        category = "Metalwork",
        price = 2100.0,
        rating = 4.2,
    ),
    Product(
        id = "p15",
        name = "Handwritten Sanskrit Sloka",
        sellerName = "Devika Calligraphy",
        description = "Beautifully hand-lettered Sanskrit sloka on handmade paper framed in a walnut border.",
        category = "Calligraphy",
        price = 990.0,
        rating = 4.7,
    ),
    Product(
        id = "p16",
        name = "Marathi Poem Frame",
        sellerName = "Graphé Ink",
        description = "Custom calligraphy frame of a classic Marathi poem, rendered in flowing ink letters.",
        category = "Calligraphy",
        price = 1500.0,
        rating = 4.0,
    ),
)