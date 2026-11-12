package kg.kasapchy.mobile.pages

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.Check
import androidx.compose.material.icons.outlined.ShoppingBag
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import kg.kasapchy.mobile.entities.*
import kg.kasapchy.mobile.features.AppStore
import kg.kasapchy.mobile.features.Page
import kg.kasapchy.mobile.shared.ui.*

@OptIn(ExperimentalLayoutApi::class)
@Composable
fun CheckoutScreen(store: AppStore, id: Int?) = PageColumn {
    val product = store.product(id)
    if (product == null) { Heading("Товар недоступен"); return@PageColumn }
    var quantity by rememberSaveable { mutableIntStateOf(1) }
    var deliveryName by rememberSaveable { mutableStateOf(Delivery.Pickup.name) }
    var address by rememberSaveable { mutableStateOf("") }
    var error by rememberSaveable { mutableStateOf<String?>(null) }
    val delivery = Delivery.valueOf(deliveryName)
    Heading("Ваш заказ")
    Panel {
        Text(product.name, style = MaterialTheme.typography.titleMedium)
        Muted("${product.seller} · ${money(product.price)} / кг")
        Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
            Text("Количество", Modifier.weight(1f))
            OutlinedIconButton({ quantity--; error = null }, enabled = quantity > 1) { Text("−") }
            Text("$quantity кг", Modifier.padding(horizontal = 8.dp))
            OutlinedIconButton({ quantity++; error = null }, enabled = quantity < product.stock) { Text("+") }
        }
    }
    Text("Получение", style = MaterialTheme.typography.titleMedium)
    FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
        Delivery.entries.forEach { FilterChip(delivery == it, { deliveryName = it.name }, label = { Text(it.title) }) }
    }
    if (delivery == Delivery.Courier) Field("Адрес доставки", address, { address = it; error = null }, "Улица, дом, квартира")
    else Muted(if (store.state.isDemo) "Ошский рынок, мясной павильон. Точку выдачи подтвердит продавец." else "Точку самовывоза подтвердит продавец по email вашего заказа.")
    Panel {
        SummaryRow("Товары", money(product.price * quantity))
        if (delivery == Delivery.Courier) SummaryRow("Доставка", money(delivery.price))
        HorizontalDivider()
        SummaryRow("Итого", money(product.price * quantity + delivery.price))
    }
    Muted("Оплата при получении. Вес и сумму подтвердит касапчы.")
    ErrorMessage(error)
    PrimaryButton("Подтвердить заказ", product.stock > 0) { error = store.placeOrder(product.id, quantity, delivery, address) }
}

@Composable
fun SuccessScreen(store: AppStore, id: Int?) = PageColumn {
    val order = store.order(id)
    if (order == null) { Heading("Заказ не найден"); return@PageColumn }
    Surface(shape = CircleShape, color = MaterialTheme.colorScheme.surfaceVariant, modifier = Modifier.align(Alignment.CenterHorizontally).padding(top = 24.dp)) {
        Icon(Icons.Outlined.Check, null, Modifier.padding(24.dp).size(36.dp), tint = MaterialTheme.colorScheme.primary)
    }
    Heading("Заказ отправлен", "Касапчы подтвердит наличие и подготовит ваш заказ.")
    Panel {
        Text("Заказ №${order.id}", style = MaterialTheme.typography.titleMedium)
        Text("${order.product.name} · ${order.quantity} кг")
        SummaryRow("Итого", money(order.total))
    }
    PrimaryButton("Перейти к заказу") { store.showCreatedOrder(order.id) }
    SecondaryButton("Продолжить покупки") { store.selectTab(Page.Market) }
}

@Composable
fun OrdersScreen(store: AppStore) = PageColumn {
    Eyebrow(if (store.state.isSeller) "ВАШ МАГАЗИН" else "ВСЁ ПОД РУКОЙ")
    Heading("Мои заказы")
    Badge(if (store.state.isSeller) "Входящие заказы" else "Покупки")
    if (!store.state.isDemo) TextButton(onClick = store::refresh) { Text("Обновить заказы") }
    if (store.state.visibleOrders.isEmpty()) {
        Icon(Icons.Outlined.ShoppingBag, null, Modifier.align(Alignment.CenterHorizontally).padding(top = 40.dp).size(50.dp), tint = MaterialTheme.colorScheme.primary)
        Heading("Пока нет заказов", "Выберите мясо на рынке — здесь появится ваш первый заказ.")
        PrimaryButton("На рынок") { store.selectTab(Page.Market) }
    }
    store.state.visibleOrders.forEach { order ->
        Card(shape = RoundedCornerShape(20.dp), border = BorderStroke(1.dp, MaterialTheme.colorScheme.outlineVariant),
            colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)) {
            Column(Modifier.padding(18.dp), verticalArrangement = Arrangement.spacedBy(12.dp)) {
                SummaryRow("№${order.id}", money(order.total))
                Badge(order.status.title)
                Text("${order.product.name} · ${order.quantity} кг", style = MaterialTheme.typography.titleMedium)
                Muted(if (store.state.isSeller) "Покупатель: ${order.buyerEmail}" else order.product.seller)
                SecondaryButton("Открыть заказ") { store.navigate(Page.Order, order.id) }
            }
        }
    }
}

@Composable
fun OrderScreen(store: AppStore, id: Int?) = PageColumn {
    val order = store.order(id)
    if (order == null) { Heading("Заказ не найден"); return@PageColumn }
    Heading(when (order.status) {
        OrderStatus.Pending -> "Ждём подтверждения"
        OrderStatus.Accepted -> "Заказ принят"
        OrderStatus.Ready -> if (order.delivery == Delivery.Pickup) "Можно забирать" else "Готов к доставке"
        OrderStatus.Completed -> "Заказ завершён"
    })
    Badge(order.status.title)
    Panel {
        Text(order.product.name, style = MaterialTheme.typography.titleMedium)
        SummaryRow("Количество", "${order.quantity} кг")
        SummaryRow("Итого", money(order.total))
        HorizontalDivider()
        Text(if (store.state.isSeller) "Покупатель · ${order.buyerEmail}" else "Продавец · ${order.product.seller}")
        Muted(if (order.delivery == Delivery.Courier) "Доставка · ${order.address}" else if (store.state.isDemo) "Самовывоз · Ошский рынок" else "Самовывоз · точку выдачи подтвердит продавец")
    }
    Text("Статус заказа", style = MaterialTheme.typography.titleMedium)
    OrderStatus.entries.forEach { status ->
        Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            Surface(shape = CircleShape, color = if (status.ordinal <= order.status.ordinal) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.surfaceVariant) {
                Text("${status.ordinal + 1}", Modifier.padding(horizontal = 12.dp, vertical = 5.dp),
                    color = if (status.ordinal <= order.status.ordinal) MaterialTheme.colorScheme.onPrimary else MaterialTheme.colorScheme.onSurfaceVariant)
            }
            Text(status.title, color = if (status == order.status) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.onSurfaceVariant)
        }
    }
    if (store.state.isSeller && order.status != OrderStatus.Completed) {
        PrimaryButton(when (order.status) {
            OrderStatus.Pending -> "Принять заказ"
            OrderStatus.Accepted -> "Готов к выдаче"
            else -> "Завершить заказ"
        }) { store.advanceOrder(order.id) }
    } else SecondaryButton("Все заказы") { store.selectTab(Page.Orders) }
}
