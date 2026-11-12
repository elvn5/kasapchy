package kg.kasapchy.mobile

import kg.kasapchy.mobile.entities.*
import kg.kasapchy.mobile.features.*
import kotlin.test.*

class DemoStoreTest {
    private fun buyer(email: String = "buyer@example.com") = DemoStore().apply {
        startAuth(true)
        chooseRole(Role.Buyer)
        assertNull(requestCode(email))
        assertNull(confirmCode("1234"))
    }

    @Test fun registrationRequiresOnlyValidEmailAndCode() {
        val store = DemoStore()
        store.startAuth(true)
        assertNotNull(store.requestCode("invalid"))
        assertNull(store.state.profile)
        assertNull(store.requestCode("  Buyer@Example.com "))
        assertNotNull(store.confirmCode("0000"))
        assertNull(store.state.profile)
        assertNull(store.confirmCode("1234"))
        assertEquals("buyer@example.com", store.state.profile?.email)
        assertEquals("", store.state.profile?.phone)
        assertEquals("", store.state.profile?.name)
        assertFalse(store.state.canGoBack)
    }

    @Test fun optionalPhoneCanBeAddedAndRemovedWithoutName() {
        val store = buyer()
        assertNull(store.saveProfile("", ""))
        assertNotNull(store.saveProfile("", "abc"))
        assertNull(store.saveProfile("", "+996 700 123 456"))
        assertEquals("+996 700 123 456", store.state.profile?.phone)
        assertNull(store.saveProfile("", ""))
        assertEquals("", store.state.profile?.phone)
    }

    @Test fun socialLoginPreservesSelectedRoleAndCanBeCancelled() {
        AuthProvider.entries.filter { it != AuthProvider.Email }.forEach { provider ->
            val store = DemoStore()
            store.startAuth(true)
            store.chooseRole(Role.Seller)
            store.navigate(Page.Auth)
            store.startSocial(provider)
            store.back()
            assertNull(store.state.profile)
            assertEquals(Page.Auth, store.state.route.page)
            store.startSocial(provider)
            store.confirmSocial()
            assertEquals(provider, store.state.profile?.provider)
            assertEquals(Role.Seller, store.state.profile?.role)
            assertEquals(Page.Market, store.state.route.page)
        }
    }

    @Test fun filtersCombineSearchCategoryBudgetAndSorting() {
        val result = DemoData.products.filtered("Азамат", MarketFilter(Category.Beef, 650, true))
        assertEquals(listOf(4), result.map { it.id })
        assertTrue(DemoData.products.filtered("несуществующий", MarketFilter()).isEmpty())
        assertEquals(listOf(340, 590, 680, 720), DemoData.products.filtered("", MarketFilter(cheapestFirst = true)).map { it.price })
    }

    @Test fun orderHasStableSnapshotTotalAndDeliveryAndReservesStock() {
        val store = buyer()
        assertNotNull(store.placeOrder(1, 0, Delivery.Pickup, ""))
        assertNotNull(store.placeOrder(1, 21, Delivery.Pickup, ""))
        assertNotNull(store.placeOrder(1, 2, Delivery.Courier, ""))
        assertNull(store.placeOrder(1, 2, Delivery.Courier, "Чуй 100"))
        val order = store.state.visibleOrders.single()
        assertEquals(1510, order.total)
        assertEquals("Чуй 100", order.address)
        assertEquals(18, store.product(1)?.stock)
        store.selectTab(Page.Market)
        store.navigate(Page.Product, 2)
        assertEquals("Говяжья мякоть", store.order(order.id)?.product?.name)
        assertEquals(2, store.order(order.id)?.quantity)
    }

    @Test fun cannotAdvanceOtherSellersOrdersOrBuyAsSeller() {
        val store = buyer()
        store.placeOrder(2, 1, Delivery.Pickup, "")
        val otherSellerId = store.state.visibleOrders.single().id
        store.advanceOrder(otherSellerId)
        assertEquals(OrderStatus.Pending, store.state.orders.last().status)
        store.logout()
        store.startAuth(false)
        store.chooseRole(Role.Seller)
        store.requestCode("seller@example.com")
        store.confirmCode("1234")
        assertNull(store.order(otherSellerId))
        store.advanceOrder(otherSellerId)
        assertEquals(OrderStatus.Pending, store.state.orders.last().status)
        assertNotNull(store.placeOrder(1, 1, Delivery.Pickup, ""))
        repeat(6) { store.advanceOrder(1041) }
        assertEquals(OrderStatus.Completed, store.order(1041)?.status)
    }

    @Test fun publishedProductAppearsAndInvalidInputsDoNotMutateCatalog() {
        val store = DemoStore()
        store.chooseRole(Role.Seller)
        store.requestCode("seller@example.com")
        store.confirmCode("1234")
        assertNotNull(store.publishProduct("Стейк", Category.Beef, "-1", "10"))
        assertNotNull(store.publishProduct("Стейк", Category.Beef, "680", "0"))
        assertEquals(4, store.state.products.size)
        assertNull(store.publishProduct("Стейк", Category.Beef, "800", "10"))
        assertEquals("Стейк", store.state.products.first().name)
        assertEquals(Page.Market, store.state.route.page)
    }

    @Test fun accountsDoNotSeeEachOthersPurchasesAndLogoutClearsBackstack() {
        val store = buyer()
        store.placeOrder(1, 1, Delivery.Pickup, "")
        store.logout()
        assertFalse(store.state.canGoBack)
        assertEquals(Page.Welcome, store.state.route.page)
        store.startAuth(false)
        store.requestCode("another@example.com")
        store.confirmCode("1234")
        assertTrue(store.state.visibleOrders.isEmpty())
    }

    @Test fun backAndTabsDoNotReturnToCompletedCheckout() {
        val store = buyer()
        store.navigate(Page.Product, 1)
        store.navigate(Page.Checkout, 1)
        store.back()
        assertEquals(Page.Product, store.state.route.page)
        store.placeOrder(1, 1, Delivery.Pickup, "")
        store.back()
        assertEquals(Page.Orders, store.state.route.page)
        assertFalse(store.state.canGoBack)
        store.selectTab(Page.Profile)
        assertEquals(1, store.state.history.size)
    }
}
