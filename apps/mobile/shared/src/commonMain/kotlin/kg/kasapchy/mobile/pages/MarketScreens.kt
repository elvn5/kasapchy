package kg.kasapchy.mobile.pages

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.LocationOn
import androidx.compose.material.icons.outlined.Search
import androidx.compose.material.icons.outlined.Tune
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import kg.kasapchy.mobile.entities.*
import kg.kasapchy.mobile.features.AppStore
import kg.kasapchy.mobile.features.Page
import kg.kasapchy.mobile.shared.ui.*

@OptIn(ExperimentalLayoutApi::class)
@Composable
fun MarketScreen(store: AppStore) = PageColumn {
    val state = store.state
    Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceBetween, verticalAlignment = Alignment.CenterVertically) {
        Column {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(Icons.Outlined.LocationOn, null, Modifier.size(16.dp), tint = MaterialTheme.colorScheme.onSurfaceVariant)
                Muted("Бишкек")
            }
            Heading("Рынок")
        }
        Badge(state.profile?.role?.title.orEmpty())
    }
    Panel {
        if (state.isSeller) {
            Text("Ваш прилавок", style = MaterialTheme.typography.titleLarge)
            Muted(if (state.isDemo) "Добавляйте товары и принимайте заказы. Демо-магазин: Азамат эт." else "Добавляйте товары и принимайте заказы.")
            PrimaryButton("Добавить товар") { store.navigate(Page.AddProduct) }
        } else {
            Eyebrow("К СЕМЕЙНОМУ СТОЛУ")
            Text("Свежий выбор\nкаждый день.", style = MaterialTheme.typography.titleLarge)
            Muted("Мясо от местных касапчы")
        }
    }
    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
        OutlinedTextField(state.query, store::search, Modifier.weight(1f), placeholder = { Text("Мясо или продавец") },
            leadingIcon = { Icon(Icons.Outlined.Search, null) }, singleLine = true, shape = RoundedCornerShape(15.dp))
        FilledTonalIconButton(onClick = { store.navigate(Page.Filters) }) { Icon(Icons.Outlined.Tune, "Фильтры") }
    }
    FlowRow(horizontalArrangement = Arrangement.spacedBy(6.dp)) {
        FilterChip(selected = state.filter.category == null, onClick = { store.setFilter(state.filter.copy(category = null)) }, label = { Text("Все") })
        Category.entries.forEach { category ->
            FilterChip(selected = state.filter.category == category, onClick = { store.setFilter(state.filter.copy(category = category)) }, label = { Text(category.title) })
        }
    }
    if (state.filter.maxPrice != null || state.filter.cheapestFirst) {
        Muted(listOfNotNull(state.filter.maxPrice?.let { "До ${money(it)} / кг" }, if (state.filter.cheapestFirst) "Сначала дешевле" else null).joinToString(" · "))
    }
    if (!state.isDemo) TextButton(onClick = store::refresh) { Text("Обновить предложения") }
    SummaryRow("Предложения рядом", "за 1 кг")
    val products = state.products.filtered(state.query, state.filter)
    if (products.isEmpty()) {
        Panel { Heading("Ничего не найдено", "Измените запрос или фильтры.") }
        SecondaryButton("Сбросить поиск и фильтры") { store.search(""); store.setFilter(MarketFilter()) }
    }
    products.chunked(2).forEach { row ->
        Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            row.forEach { product -> ProductCard(product, Modifier.weight(1f)) { store.navigate(Page.Product, product.id) } }
            if (row.size == 1) Spacer(Modifier.weight(1f))
        }
    }
}

@Composable
private fun ProductCard(product: Product, modifier: Modifier, onClick: () -> Unit) {
    Card(onClick, modifier, shape = RoundedCornerShape(18.dp),
        border = BorderStroke(1.dp, MaterialTheme.colorScheme.outlineVariant),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surface)) {
        Box(Modifier.fillMaxWidth().height(112.dp).background(MaterialTheme.colorScheme.surfaceVariant), contentAlignment = Alignment.Center) {
            Text(product.category.illustration, fontSize = 54.sp)
        }
        Column(Modifier.padding(12.dp), verticalArrangement = Arrangement.spacedBy(5.dp)) {
            Text(product.name, style = MaterialTheme.typography.titleSmall, minLines = 2)
            Muted("${product.seller} · Бишкек")
            Text(money(product.price), style = MaterialTheme.typography.titleMedium)
            Muted(if (product.stock > 0) "за кг · в наличии" else "Нет в наличии")
        }
    }
}

@OptIn(ExperimentalLayoutApi::class)
@Composable
fun FiltersScreen(store: AppStore) = PageColumn {
    var categoryName by rememberSaveable { mutableStateOf(store.state.filter.category?.name) }
    var maximum by rememberSaveable { mutableStateOf(store.state.filter.maxPrice?.toString().orEmpty()) }
    var cheapest by rememberSaveable { mutableStateOf(store.state.filter.cheapestFirst) }
    var error by rememberSaveable { mutableStateOf<String?>(null) }
    Heading("Ваш выбор")
    Text("Категория", style = MaterialTheme.typography.titleMedium)
    FlowRow(horizontalArrangement = Arrangement.spacedBy(6.dp)) {
        FilterChip(categoryName == null, { categoryName = null }, label = { Text("Все") })
        Category.entries.forEach { FilterChip(categoryName == it.name, { categoryName = it.name }, label = { Text(it.title) }) }
    }
    Field("Максимальная цена за кг, сом", maximum, { maximum = it; error = null }, "Без ограничения", KeyboardType.Number)
    Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
        Text("Сначала дешевле", Modifier.weight(1f)); Switch(cheapest, { cheapest = it })
    }
    ErrorMessage(error)
    PrimaryButton("Показать предложения") {
        val price = maximum.toIntOrNull()
        if (maximum.isNotBlank() && (price == null || price !in 1..100000)) error = "Укажите цену от 1 до 100 000 сом."
        else {
            store.setFilter(MarketFilter(categoryName?.let(Category::valueOf), price, cheapest))
            store.back()
        }
    }
    SecondaryButton("Сбросить фильтры") { categoryName = null; maximum = ""; cheapest = false; error = null }
}

@Composable
fun ProductScreen(store: AppStore, id: Int?) = PageColumn {
    val product = store.product(id)
    if (product == null) { Heading("Товар недоступен"); return@PageColumn }
    Box(Modifier.fillMaxWidth().height(210.dp).background(MaterialTheme.colorScheme.surfaceVariant, RoundedCornerShape(24.dp)), contentAlignment = Alignment.Center) {
        Text(product.category.illustration, fontSize = 104.sp)
    }
    Badge("Охлаждённое · ${product.category.title}")
    Heading(product.name)
    SummaryRow("${money(product.price)} / кг", "${product.stock} кг")
    Text(product.description, color = MaterialTheme.colorScheme.onSurfaceVariant)
    Panel {
        Text(product.seller, style = MaterialTheme.typography.titleMedium)
        Muted("Бишкек · Самовывоз с рынка")
    }
    if (!store.state.isSeller) PrimaryButton(if (product.stock > 0) "Оформить заказ" else "Нет в наличии", product.stock > 0) { store.navigate(Page.Checkout, product.id) }
    else SecondaryButton("Вернуться на рынок") { store.selectTab(Page.Market) }
}

@OptIn(ExperimentalLayoutApi::class)
@Composable
fun AddProductScreen(store: AppStore) = PageColumn {
    var name by rememberSaveable { mutableStateOf("") }
    var categoryName by rememberSaveable { mutableStateOf(Category.Beef.name) }
    var price by rememberSaveable { mutableStateOf("") }
    var stock by rememberSaveable { mutableStateOf("") }
    var error by rememberSaveable { mutableStateOf<String?>(null) }
    Heading("На ваш прилавок")
    Field("Название", name, { name = it }, "Например, говяжья мякоть")
    FlowRow(horizontalArrangement = Arrangement.spacedBy(6.dp)) {
        Category.entries.forEach { FilterChip(categoryName == it.name, { categoryName = it.name }, label = { Text(it.title) }) }
    }
    Field("Цена за кг, сом", price, { price = it }, "680", KeyboardType.Number)
    Field("В наличии, кг", stock, { stock = it }, "20", KeyboardType.Number)
    Muted("Используется иллюстрация категории.")
    ErrorMessage(error)
    PrimaryButton("Опубликовать товар") { error = store.publishProduct(name, Category.valueOf(categoryName), price, stock) }
}
