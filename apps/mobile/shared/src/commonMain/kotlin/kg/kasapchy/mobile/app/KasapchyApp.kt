package kg.kasapchy.mobile.app

import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.ArrowBack
import androidx.compose.material.icons.outlined.PersonOutline
import androidx.compose.material.icons.automirrored.outlined.ReceiptLong
import androidx.compose.material.icons.outlined.Storefront
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import kg.kasapchy.mobile.shared.ui.LocalBusy
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import kg.kasapchy.mobile.features.AppStore
import kg.kasapchy.mobile.features.Page
import kg.kasapchy.mobile.pages.*
import kg.kasapchy.mobile.shared.ui.KasapchyTheme

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun KasapchyApp(store: AppStore, privacyUrl: String = "", deletionUrl: String = "") {
    KasapchyTheme {
        val state = store.state
        val route = state.route
        CompositionLocalProvider(LocalBusy provides state.busy) {
        Surface(Modifier.fillMaxSize(), color = MaterialTheme.colorScheme.background) {
            Box(Modifier.fillMaxSize(), contentAlignment = Alignment.TopCenter) {
                Scaffold(
                    modifier = Modifier.widthIn(max = 600.dp).fillMaxSize().imePadding(),
                    containerColor = MaterialTheme.colorScheme.background,
                    topBar = { Column {
                        if (state.busy) LinearProgressIndicator(Modifier.fillMaxWidth())
                        state.error?.let { message ->
                            TextButton(onClick = store::clearError, modifier = Modifier.fillMaxWidth()) { Text(message, color = MaterialTheme.colorScheme.error) }
                        }
                        if (state.canGoBack) TopAppBar(
                            title = { Text(when (route.page) {
                                Page.RoleChoice -> if (state.registering) "Регистрация" else "Вход"
                                Page.Auth -> if (state.registering) "Создать аккаунт" else "Вход"
                                Page.Code -> "Подтверждение email"
                                Page.Social -> "Вход через ${state.socialProvider.title}"
                                Page.Filters -> "Фильтры"
                                Page.Product -> "Товар"
                                Page.Checkout -> "Оформление"
                                Page.Success -> "Заказ отправлен"
                                Page.Order -> "Заказ №${route.id}"
                                Page.EditProfile -> if (state.isSeller) "Данные магазина" else "Личные данные"
                                Page.AddProduct -> "Новый товар"
                                else -> "Касапчы"
                            }, style = MaterialTheme.typography.titleMedium) },
                            navigationIcon = { IconButton(onClick = store::back) { Icon(Icons.AutoMirrored.Outlined.ArrowBack, "Назад") } },
                            colors = TopAppBarDefaults.topAppBarColors(containerColor = MaterialTheme.colorScheme.background),
                        )
                    } },
                    bottomBar = {
                        if (state.profile != null) {
                            val selectedTab = when (route.page) {
                                Page.Orders, Page.Order, Page.Success -> Page.Orders
                                Page.Profile, Page.EditProfile -> Page.Profile
                                else -> Page.Market
                            }
                            NavigationBar(containerColor = MaterialTheme.colorScheme.surface, tonalElevation = 0.dp) {
                                listOf(Triple(Page.Market, "Рынок", Icons.Outlined.Storefront),
                                    Triple(Page.Orders, "Мои заказы", Icons.AutoMirrored.Outlined.ReceiptLong),
                                    Triple(Page.Profile, "Профиль", Icons.Outlined.PersonOutline)).forEach { (page, label, icon) ->
                                    NavigationBarItem(selected = selectedTab == page, onClick = { store.selectTab(page) },
                                        icon = { Icon(icon, null) }, label = { Text(label) })
                                }
                            }
                        }
                    },
                ) { padding ->
                    AnimatedContent(route, Modifier.padding(padding), transitionSpec = { fadeIn() togetherWith fadeOut() }, label = "screen") { current ->
                        when (current.page) {
                            Page.Welcome -> WelcomeScreen(store)
                            Page.RoleChoice -> RoleScreen(store)
                            Page.Auth -> AuthScreen(store, privacyUrl)
                            Page.Code -> CodeScreen(store)
                            Page.Social -> SocialScreen(store)
                            Page.Market -> MarketScreen(store)
                            Page.Filters -> FiltersScreen(store)
                            Page.Product -> ProductScreen(store, current.id)
                            Page.Checkout -> CheckoutScreen(store, current.id)
                            Page.Success -> SuccessScreen(store, current.id)
                            Page.Orders -> OrdersScreen(store)
                            Page.Order -> OrderScreen(store, current.id)
                            Page.Profile -> ProfileScreen(store, privacyUrl, deletionUrl)
                            Page.EditProfile -> EditProfileScreen(store)
                            Page.AddProduct -> AddProductScreen(store)
                        }
                    }
                }
            }
        }
    }
}
}
