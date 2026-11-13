package kg.kasapchy.mobile.features

import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import kg.kasapchy.mobile.entities.AuthProvider
import kg.kasapchy.mobile.entities.Delivery
import kg.kasapchy.mobile.entities.DemoData
import kg.kasapchy.mobile.entities.Category
import kg.kasapchy.mobile.entities.MarketFilter
import kg.kasapchy.mobile.entities.Order
import kg.kasapchy.mobile.entities.Product
import kg.kasapchy.mobile.entities.Profile
import kg.kasapchy.mobile.entities.Role
import kg.kasapchy.mobile.entities.emailError
import kg.kasapchy.mobile.entities.phoneError

enum class Page {
    Welcome, RoleChoice, Auth, Code, Social, Market, Filters, Product, Checkout, Success, Orders, Order, Profile, EditProfile, AddProduct
}
data class Route(val page: Page, val id: Int? = null)
data class AppState(
    val isDemo: Boolean = true,
    val busy: Boolean = false,
    val error: String? = null,
    val notice: String? = null,
    val history: List<Route> = listOf(Route(Page.Welcome)),
    val selectedRole: Role = Role.Buyer,
    val registering: Boolean = true,
    val pendingEmail: String = "",
    val socialProvider: AuthProvider = AuthProvider.Google,
    val profile: Profile? = null,
    val products: List<Product> = DemoData.products,
    val orders: List<Order> = listOf(DemoData.incomingOrder),
    val query: String = "",
    val filter: MarketFilter = MarketFilter(),
) {
    val route: Route get() = history.last()
    val canGoBack: Boolean get() = history.size > 1
    val isSeller: Boolean get() = profile?.role == Role.Seller
    val visibleOrders: List<Order> get() = if (!isDemo) orders else orders.filter {
        if (isSeller) it.product.sellerId == DemoData.sellerId else it.buyerEmail == profile?.email
    }.reversed()
}

/** Session-scoped demo data. No network requests, credentials, or real OAuth tokens. */
class DemoStore : AppStore {
    override var state by mutableStateOf(AppState())
        private set
    private val profiles = mutableMapOf<Pair<String, Role>, Profile>()

    override fun startAuth(registering: Boolean) {
        state = state.copy(registering = registering, history = listOf(Route(Page.Welcome), Route(Page.RoleChoice)))
    }
    override fun chooseRole(role: Role) { state = state.copy(selectedRole = role) }
    override fun toggleAuth() { state = state.copy(registering = !state.registering) }
    override fun navigate(page: Page, id: Int?) {
        val route = Route(page, id)
        if (route != state.route) state = state.copy(history = state.history + route)
    }
    override fun back() {
        if (state.canGoBack) state = state.copy(history = state.history.dropLast(1))
    }
    override fun selectTab(page: Page) {
        require(page in listOf(Page.Market, Page.Orders, Page.Profile))
        state = state.copy(history = listOf(Route(page)))
    }
    override fun passwordAuth(email: String, password: String): String? {
        return "В демо используйте демонстрационный код."
    }
    override fun requestCode(email: String): String? {
        val normalized = email.trim().lowercase()
        emailError(normalized)?.let { return it }
        state = state.copy(pendingEmail = normalized)
        navigate(Page.Code)
        return null
    }
    override fun confirmCode(code: String): String? {
        if (code != DemoData.code) return "Код не подошёл. Для демо введите ${DemoData.code}."
        emailError(state.pendingEmail)?.let { return "Сначала укажите email." }
        signIn(state.pendingEmail, AuthProvider.Email)
        return null
    }
    override fun startSocial(provider: AuthProvider) {
        require(provider != AuthProvider.Email)
        state = state.copy(socialProvider = provider)
        navigate(Page.Social)
    }
    override fun confirmSocial() {
        val email = if (state.socialProvider == AuthProvider.Apple) "demo@privaterelay.example.com" else "demo@example.com"
        signIn(email, state.socialProvider)
    }
    private fun signIn(email: String, provider: AuthProvider) {
        val key = email to state.selectedRole
        val profile = profiles[key]?.copy(provider = provider) ?: Profile(email, state.selectedRole, provider)
        profiles[key] = profile
        state = state.copy(profile = profile, history = listOf(Route(Page.Market)), query = "", filter = MarketFilter())
    }
    override fun logout() {
        state = state.copy(profile = null, pendingEmail = "", history = listOf(Route(Page.Welcome)), query = "", filter = MarketFilter())
    }
    override fun saveProfile(name: String, phone: String): String? {
        val profile = state.profile ?: return "Сначала войдите в аккаунт."
        phoneError(phone)?.let { return it }
        if (name.trim().length > 60) return "Максимум 60 символов."
        val updated = profile.copy(name = name.trim(), phone = phone.trim())
        profiles[updated.email to updated.role] = updated
        state = state.copy(profile = updated)
        back()
        return null
    }
    override fun search(query: String) { state = state.copy(query = query) }
    override fun setFilter(filter: MarketFilter) { state = state.copy(filter = filter) }
    override fun product(id: Int?): Product? = state.products.find { it.id == id }
    override fun order(id: Int?): Order? = state.visibleOrders.find { it.id == id }

    override fun placeOrder(productId: Int, quantity: Int, delivery: Delivery, address: String): String? {
        val profile = state.profile ?: return "Сначала войдите в аккаунт."
        if (profile.role != Role.Buyer) return "Заказы оформляют покупатели."
        val product = product(productId) ?: return "Товар больше недоступен."
        if (quantity !in 1..product.stock) return "Недостаточно товара в наличии."
        if (delivery == Delivery.Courier && address.trim().length < 5) return "Укажите улицу и номер дома."
        val order = Order((state.orders.maxOfOrNull { it.id } ?: 1041) + 1, product, quantity, profile.email,
            delivery, if (delivery == Delivery.Courier) address.trim() else "")
        state = state.copy(
            products = state.products.map { if (it.id == productId) it.copy(stock = it.stock - quantity) else it },
            orders = state.orders + order,
            history = listOf(Route(Page.Orders), Route(Page.Success, order.id)),
        )
        return null
    }
    override fun showCreatedOrder(id: Int) { state = state.copy(history = listOf(Route(Page.Orders), Route(Page.Order, id))) }
    override fun advanceOrder(id: Int) {
        if (!state.isSeller) return
        state = state.copy(orders = state.orders.map {
            if (it.id == id && it.product.sellerId == DemoData.sellerId) it.copy(status = it.status.next()) else it
        })
    }
    override fun publishProduct(name: String, category: Category, price: String, stock: String, acceptedPolicy: Boolean): String? {
        val profile = state.profile ?: return "Сначала войдите в аккаунт."
        if (profile.role != Role.Seller) return "Товары добавляют касапчы."
        if (name.trim().length !in 2..60) return "Название должно содержать от 2 до 60 символов."
        val amount = price.toIntOrNull()?.takeIf { it in 1..100000 } ?: return "Укажите цену от 1 до 100 000 сом."
        val available = stock.toIntOrNull()?.takeIf { it in 1..10000 } ?: return "Укажите остаток от 1 до 10 000 кг."
        val product = Product((state.products.maxOfOrNull { it.id } ?: 0) + 1, name.trim(), category, amount,
            profile.name.ifBlank { "Азамат эт" }, DemoData.sellerId, available, "Свежий товар от местного продавца.")
        state = state.copy(products = listOf(product) + state.products, query = "", filter = MarketFilter(),
            history = listOf(Route(Page.Market)))
        return null
    }
}
