package kg.kasapchy.mobile.pages

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.outlined.CheckCircle
import androidx.compose.material.icons.outlined.MailOutline
import androidx.compose.material.icons.outlined.ShoppingBag
import androidx.compose.material.icons.outlined.Storefront
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalUriHandler
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import kg.kasapchy.mobile.entities.AuthProvider
import kg.kasapchy.mobile.entities.Role
import kg.kasapchy.mobile.features.AppStore
import kg.kasapchy.mobile.features.Page
import kg.kasapchy.mobile.shared.ui.*

@Composable
fun WelcomeScreen(store: AppStore) = PageColumn {
    Text("касапчы", style = MaterialTheme.typography.titleLarge, color = MaterialTheme.colorScheme.primary)
    Panel {
        Eyebrow("СВОЁ. СВЕЖЕЕ. РЯДОМ.")
        Text("🥩 🌿", Modifier.align(Alignment.CenterHorizontally).padding(vertical = 22.dp), fontSize = 70.sp)
        Heading("Хорошее мясо.\nЗнакомые люди.", "Местные касапчы и свежие продукты — в одном приложении.")
    }
    Badge("Бишкек · От местных продавцов")
    Spacer(Modifier.height(12.dp))
    PrimaryButton("Создать аккаунт") { store.startAuth(true) }
    SecondaryButton("Уже есть аккаунт") { store.startAuth(false) }
    if (store.state.isDemo) Muted("Демонстрационная версия · без реальных покупок")
}

@Composable
fun RoleScreen(store: AppStore) = PageColumn {
    Heading("Как вы будете\nпользоваться?", "Выберите роль для этого аккаунта.")
    Role.entries.forEach { role ->
        val selected = store.state.selectedRole == role
        Card(onClick = { store.chooseRole(role) }, modifier = Modifier.fillMaxWidth(), shape = RoundedCornerShape(18.dp),
            border = BorderStroke(if (selected) 2.dp else 1.dp, if (selected) MaterialTheme.colorScheme.primary else MaterialTheme.colorScheme.outlineVariant),
            colors = CardDefaults.cardColors(containerColor = if (selected) MaterialTheme.colorScheme.surfaceVariant else MaterialTheme.colorScheme.surface)) {
            Row(Modifier.padding(20.dp), horizontalArrangement = Arrangement.spacedBy(14.dp), verticalAlignment = Alignment.CenterVertically) {
                Icon(if (role == Role.Buyer) Icons.Outlined.ShoppingBag else Icons.Outlined.Storefront, null)
                Column(Modifier.weight(1f)) {
                    Text(role.title, style = MaterialTheme.typography.titleMedium)
                    Muted(if (role == Role.Buyer) "Найти мясо и оформить заказ" else "Продавать и принимать заказы")
                }
                if (selected) Icon(Icons.Outlined.CheckCircle, "Выбрано", tint = MaterialTheme.colorScheme.primary)
            }
        }
    }
    PrimaryButton("Продолжить") { store.navigate(Page.Auth) }
}

@Composable
fun AuthScreen(store: AppStore, privacyUrl: String = "") = PageColumn {
    var email by rememberSaveable { mutableStateOf(store.state.pendingEmail) }
    // Passwords must never be persisted in saved instance state.
    var password by remember(store.state.registering) { mutableStateOf("") }
    var confirmation by remember(store.state.registering) { mutableStateOf("") }
    var error by rememberSaveable { mutableStateOf<String?>(null) }
    Badge(store.state.selectedRole.title)
    Heading(if (store.state.registering) "Начнём с email." else "С возвращением.", if (store.state.isDemo) "Без пароля. Подтвердите email одноразовым кодом." else "Вход по email и паролю. Код из письма не нужен.")
    Field("Email", email, { email = it; error = null }, "you@example.com", KeyboardType.Email)
    if (!store.state.isDemo) {
        Field("Пароль", password, { password = it; error = null }, keyboard = KeyboardType.Password, password = true)
        if (store.state.registering) {
            Field("Повторите пароль", confirmation, { confirmation = it; error = null }, keyboard = KeyboardType.Password, password = true)
            Muted("Не менее 12 символов латиницей. Сохраните пароль: восстановление по почте пока недоступно.")
        } else Muted("Используйте пароль, заданный при регистрации. Восстановление по почте пока недоступно.")
    }
    ErrorMessage(error)
    PrimaryButton(if (store.state.isDemo) "Продолжить с email" else if (store.state.registering) "Создать аккаунт" else "Войти") {
        error = if (store.state.isDemo) store.requestCode(email)
        else if (store.state.registering && password != confirmation) "Пароли не совпадают."
        else store.passwordAuth(email, password)
    }
    if (store.state.isDemo) {
    Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(12.dp)) {
        HorizontalDivider(Modifier.weight(1f)); Muted("или"); HorizontalDivider(Modifier.weight(1f))
    }
    SecondaryButton("Продолжить с Google") { store.startSocial(AuthProvider.Google) }
    SecondaryButton("Продолжить с Apple") { store.startSocial(AuthProvider.Apple) }
    }
    Muted("Телефон можно добавить позже в профиле — это необязательно.")
    TextButton(onClick = store::toggleAuth, modifier = Modifier.align(Alignment.CenterHorizontally)) {
        Text(if (store.state.registering) "Уже есть аккаунт? Войти" else "Нет аккаунта? Зарегистрироваться")
    }
    if (store.state.isDemo) Muted("Демо: письма не отправляются.")
    val uriHandler = LocalUriHandler.current
    if (privacyUrl.startsWith("https://")) TextButton(onClick = { uriHandler.openUri(privacyUrl) }) { Text("Политика конфиденциальности") }
}

@Composable
fun CodeScreen(store: AppStore) = PageColumn {
    var code by rememberSaveable { mutableStateOf("") }
    var error by rememberSaveable { mutableStateOf<String?>(null) }
    Icon(Icons.Outlined.MailOutline, null, tint = MaterialTheme.colorScheme.primary)
    Heading("Проверьте почту", "Введите код для ${store.state.pendingEmail}")
    Field("Код из письма", code, { code = it.filter(Char::isDigit).take(if (store.state.isDemo) 4 else 6); error = null }, if (store.state.isDemo) "1234" else "000000", KeyboardType.Number)
    if (store.state.isDemo) Muted("Демо-код: 1234") else Muted("Код действует 10 минут.")
    ErrorMessage(error)
    PrimaryButton(if (store.state.registering) "Создать аккаунт" else "Войти") { error = store.confirmCode(code) }
    SecondaryButton("Изменить email", store::back)
}

@Composable
fun SocialScreen(store: AppStore) = PageColumn {
    val provider = store.state.socialProvider
    Heading("Продолжить\nс ${provider.title}", "${if (store.state.registering) "Создать аккаунт" else "Войти"} как ${store.state.selectedRole.title.lowercase()}.")
    if (store.state.isDemo) {
        Panel {
            Text("Демо-аккаунт", style = MaterialTheme.typography.titleMedium)
            Text(if (provider == AuthProvider.Apple) "demo@privaterelay.example.com" else "demo@example.com")
        }
        Muted("Имитация авторизации. Подключение к ${provider.title} не выполняется.")
    } else {
        Muted(if (provider == AuthProvider.Apple) "Откроется защищённая страница Apple. После подтверждения вернитесь в приложение." else "Выберите Google-аккаунт в системном окне.")
    }
    PrimaryButton(if (store.state.isDemo) "Продолжить с демо-аккаунтом" else "Войти через ${provider.title}", onClick = store::confirmSocial)
    SecondaryButton("Отмена", store::back)
}
