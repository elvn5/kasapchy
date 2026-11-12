package kg.kasapchy.mobile.features

import androidx.compose.runtime.*
import io.ktor.client.call.body
import io.ktor.http.HttpMethod
import kotlinx.coroutines.*
import kotlinx.serialization.encodeToString
import kotlinx.serialization.json.*
import kotlin.random.Random
import kg.kasapchy.mobile.entities.*
import kg.kasapchy.mobile.shared.api.*

/** Server-backed state. The server owns roles, prices, stock, identity and order transitions. */
class NetworkStore(private val api: MarketApi, private val storage: SessionStorage, private val scope: CoroutineScope) : AppStore {
    override var state by mutableStateOf(AppState(isDemo = false, products = emptyList(), orders = emptyList()))
        private set
    var socialLogin: (suspend (AuthProvider, String, String) -> SessionResponse)? = null
    private var activeJob: Job? = null
    private var challenge = ""
    private var pendingOrder: Pair<String, String>? = null
    init {
        api.token = storage.read()
        if (api.token != null) runRequest {
            val profile = api.get<Profile>("/me")
            state = state.copy(profile = profile, history = listOf(Route(Page.Market)))
            loadData()
        }
    }
    private fun runRequest(block: suspend () -> Unit) {
        if (state.busy) return
        state = state.copy(busy = true, error = null)
        activeJob = scope.launch {
            try { block() }
            catch (e: CancellationException) { throw e }
            catch (e: Exception) {
                if (e is ApiFailure && e.code == 401 && api.token != null) {
                    api.token = null; storage.save(null)
                    state = state.copy(profile = null, orders = emptyList(), history = listOf(Route(Page.Welcome)))
                }
                state = state.copy(error = if (e is ApiFailure) e.message else "Не удалось подключиться. Проверьте интернет и повторите.")
            } finally { state = state.copy(busy = false) }
        }
    }
    private suspend fun loadData() {
        val products = api.get<List<Product>>("/products")
        val orders = if (state.profile != null) api.get<List<Order>>("/orders") else emptyList()
        state = state.copy(products = products, orders = orders)
    }
    private suspend fun signedIn(session: SessionResponse) {
        api.token = session.token; storage.save(session.token)
        state = state.copy(profile = session.user, products = emptyList(), orders = emptyList(), history = listOf(Route(Page.Market)))
        loadData()
    }
    override fun startAuth(registering: Boolean) { state = state.copy(registering = registering, error = null, history = listOf(Route(Page.Welcome), Route(Page.RoleChoice))) }
    override fun chooseRole(role: Role) { state = state.copy(selectedRole = role) }
    override fun toggleAuth() { state = state.copy(registering = !state.registering) }
    override fun navigate(page: Page, id: Int?) { if (!state.busy) state = state.copy(error = null, history = state.history + Route(page, id)) }
    override fun back() { if (state.route.page == Page.Social && state.busy) { activeJob?.cancel(); state = state.copy(busy = false) }; if (state.canGoBack && !state.busy) state = state.copy(error = null, history = state.history.dropLast(1)) }
    override fun selectTab(page: Page) { if (state.busy) return; state = state.copy(error = null, history = listOf(Route(page))); refresh() }
    override fun refresh() { runRequest { loadData() } }
    override fun clearError() { state = state.copy(error = null) }
    override fun requestCode(email: String): String? {
        val normalized = email.trim().lowercase()
        emailError(normalized)?.let { return it }
        runRequest {
            challenge = api.post<ChallengeResponse>("/auth/email/request", mapOf("email" to normalized, "role" to state.selectedRole.name)).challengeId
            state = state.copy(pendingEmail = normalized, history = state.history + Route(Page.Code))
        };return null
    }
    override fun confirmCode(code: String): String? {
        if (code.length != 6) return "Введите 6 цифр из письма."
        runRequest { signedIn(api.post("/auth/email/verify", mapOf("challengeId" to challenge, "code" to code))) }; return null
    }
    override fun startSocial(provider: AuthProvider) { state = state.copy(socialProvider = provider); navigate(Page.Social) }
    override fun confirmSocial() {
        val login = socialLogin ?: run { state = state.copy(error = "Этот способ входа ещё не настроен."); return }
        runRequest { signedIn(login(state.socialProvider, state.selectedRole.name, api.baseUrl)) }
    }
    override fun logout() { runRequest { api.request(HttpMethod.Post, "/auth/logout"); api.token = null; storage.save(null); state = AppState(isDemo = false, products = emptyList(), orders = emptyList()) } }
    override fun saveProfile(name: String, phone: String): String? {
        phoneError(phone)?.let { return it }; if (name.length > 60) return "Максимум 60 символов."
        runRequest { val profile = api.request(HttpMethod.Patch, "/me", mapOf("name" to name.trim(), "phone" to phone.trim())).body<Profile>(); state = state.copy(profile = profile, history = state.history.dropLast(1)) };return null
    }
    override fun deleteAccount() { runRequest { api.request(HttpMethod.Delete, "/me"); api.token = null; storage.save(null); state = AppState(isDemo = false, products = emptyList(), orders = emptyList()) } }
    override fun search(query: String) { state = state.copy(query = query) }
    override fun setFilter(filter: MarketFilter) { state = state.copy(filter = filter) }
    override fun product(id: Int?) = state.products.find { it.id == id }
    override fun order(id: Int?) = state.orders.find { it.id == id }
    override fun placeOrder(productId: Int, quantity: Int, delivery: Delivery, address: String): String? {
        if (delivery == Delivery.Courier && address.trim().length < 5) return "Укажите улицу и номер дома."
        val signature = "$productId:$quantity:$delivery:${address.trim()}"
        // Retain the key on network failure: a retry cannot duplicate an accepted order.
        val key = pendingOrder?.takeIf { it.first == signature }?.second ?: List(32) { Random.nextInt(16).toString(16) }.joinToString("")
        pendingOrder = signature to key
        val payload = buildJsonObject { put("productId", productId); put("quantity", quantity); put("delivery", delivery.name); put("address", address.trim()); put("requestId", key) }
        runRequest {
            val order = api.request(HttpMethod.Post, "/orders", rawJson = payload.toString()).body<Order>()
            pendingOrder = null
            state = state.copy(orders = listOf(order) + state.orders.filterNot { it.id == order.id }, history = listOf(Route(Page.Orders), Route(Page.Success, order.id)))
            loadData()
        }; return null
    }
    override fun showCreatedOrder(id: Int) { state = state.copy(history = listOf(Route(Page.Orders), Route(Page.Order, id))) }
    override fun advanceOrder(id: Int) {
        val order = order(id) ?: return
        runRequest { api.post<Order>("/orders/$id/advance", mapOf("expectedStatus" to order.status.name)); loadData() }
    }
    override fun publishProduct(name: String, category: Category, price: String, stock: String): String? {
        val amount = price.toIntOrNull()?.takeIf { it in 1..100000 } ?: return "Укажите цену от 1 до 100 000 сом."
        val available = stock.toIntOrNull()?.takeIf { it in 1..10000 } ?: return "Укажите остаток от 1 до 10 000 кг."
        if (name.trim().length !in 2..60) return "Название должно содержать от 2 до 60 символов."
        val payload = buildJsonObject { put("name", name.trim()); put("category", category.name); put("price", amount); put("stock", available); put("description", "") }
        runRequest { api.request(HttpMethod.Post, "/products", rawJson = payload.toString()); state = state.copy(history = listOf(Route(Page.Market)), query = "", filter = MarketFilter()); loadData() };return null
    }
    fun close() { api.close() }
}
