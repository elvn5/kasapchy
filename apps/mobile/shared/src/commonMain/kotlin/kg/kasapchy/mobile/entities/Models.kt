package kg.kasapchy.mobile.entities

@kotlinx.serialization.Serializable
enum class Role(val title: String) { Buyer("Покупатель"), Seller("Касапчы") }
@kotlinx.serialization.Serializable
enum class AuthProvider(val title: String) { Email("Email"), Google("Google"), Apple("Apple") }
@kotlinx.serialization.Serializable
enum class Category(val title: String, val illustration: String) {
    Beef("Говядина", "🥩"), Lamb("Баранина", "🍖"), Poultry("Птица", "🍗")
}
@kotlinx.serialization.Serializable
enum class OrderStatus(val title: String) {
    Pending("Ожидает подтверждения"), Accepted("Принят"), Ready("Готов к выдаче"), Completed("Завершён");
    fun next(): OrderStatus = entries.getOrElse(ordinal + 1) { this }
}
@kotlinx.serialization.Serializable
enum class Delivery(val title: String, val price: Int) { Pickup("Самовывоз", 0), Courier("Доставка", 150) }

@kotlinx.serialization.Serializable
data class Profile(
    val email: String,
    val role: Role,
    val provider: AuthProvider,
    val name: String = "",
    val phone: String = "",
    val id: String = "",
) {
    val displayName: String get() = name.ifBlank { if (role == Role.Seller) "Мой магазин" else "Покупатель" }
}
@kotlinx.serialization.Serializable
data class Product(
    val id: Int,
    val name: String,
    val category: Category,
    val price: Int,
    val seller: String,
    val sellerId: String,
    val stock: Int,
    val description: String,
)
@kotlinx.serialization.Serializable
data class Order(
    val id: Int,
    val product: Product, // Snapshot: edits to the catalog must never alter an existing order.
    val quantity: Int,
    val buyerEmail: String,
    val delivery: Delivery,
    val address: String,
    val status: OrderStatus = OrderStatus.Pending,
) {
    val total: Int get() = product.price * quantity + delivery.price
}
@kotlinx.serialization.Serializable
data class MarketFilter(val category: Category? = null, val maxPrice: Int? = null, val cheapestFirst: Boolean = false)

fun List<Product>.filtered(query: String, filter: MarketFilter): List<Product> = filter {
    (filter.category == null || it.category == filter.category) &&
        (filter.maxPrice == null || it.price <= filter.maxPrice) &&
        ("${it.name} ${it.seller}".contains(query.trim(), ignoreCase = true))
}.let { if (filter.cheapestFirst) it.sortedBy(Product::price) else it }

fun emailError(email: String): String? =
    if (email.length <= 254 && Regex("^[^\\s@]+@[^\\s@]+\\.[^\\s@]+$").matches(email.trim())) null
    else "Введите корректный email, например you@example.com."

fun phoneError(phone: String): String? = when {
    phone.isBlank() -> null
    !Regex("^\\+?[\\d ()-]+$").matches(phone.trim()) || phone.count(Char::isDigit) !in 7..15 ->
        "Проверьте номер телефона или оставьте поле пустым."
    else -> null
}

fun money(amount: Int): String = amount.toString().reversed().chunked(3).joinToString(" ").reversed() + " с"
