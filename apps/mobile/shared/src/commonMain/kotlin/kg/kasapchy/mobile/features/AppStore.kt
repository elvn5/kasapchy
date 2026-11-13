package kg.kasapchy.mobile.features

import kg.kasapchy.mobile.entities.*

interface AppStore {
    val state: AppState
    fun startAuth(registering: Boolean)
    fun chooseRole(role: Role)
    fun toggleAuth()
    fun navigate(page: Page, id: Int? = null)
    fun back()
    fun selectTab(page: Page)
    fun passwordAuth(email: String, password: String): String?
    fun requestCode(email: String): String?
    fun confirmCode(code: String): String?
    fun startSocial(provider: AuthProvider)
    fun confirmSocial()
    fun logout()
    fun saveProfile(name: String, phone: String): String?
    fun search(query: String)
    fun setFilter(filter: MarketFilter)
    fun product(id: Int?): Product?
    fun order(id: Int?): Order?
    fun placeOrder(productId: Int, quantity: Int, delivery: Delivery, address: String): String?
    fun showCreatedOrder(id: Int)
    fun advanceOrder(id: Int)
    fun publishProduct(name: String, category: Category, price: String, stock: String, acceptedPolicy: Boolean = false): String?
    fun reportProduct(id: Int, reason: String) {}
    fun blockSeller(id: String) {}
    fun unblockSellers() {}
    fun clearError() {}
    fun refresh() {}
    fun deleteAccount() {}
}
