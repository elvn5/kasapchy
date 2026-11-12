package kg.kasapchy.mobile.pages

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.outlined.KeyboardArrowRight
import androidx.compose.material.icons.outlined.PersonOutline
import androidx.compose.material.icons.outlined.Storefront
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalUriHandler
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import kg.kasapchy.mobile.features.AppStore
import kg.kasapchy.mobile.features.Page
import kg.kasapchy.mobile.shared.ui.*

@Composable
fun ProfileScreen(store: AppStore, privacyUrl: String = "", deletionUrl: String = "") = PageColumn {
    val profile = store.state.profile ?: return@PageColumn
    val uriHandler = LocalUriHandler.current
    if (privacyUrl.startsWith("https://")) TextButton(onClick = { uriHandler.openUri(privacyUrl) }) { Text("Конфиденциальность") }
    if (deletionUrl.startsWith("https://")) TextButton(onClick = { uriHandler.openUri(deletionUrl) }) { Text("Удаление данных на сайте") }
    Heading("Профиль")
    Row(horizontalArrangement = Arrangement.spacedBy(16.dp), verticalAlignment = Alignment.CenterVertically) {
        Surface(shape = CircleShape, color = MaterialTheme.colorScheme.surfaceVariant) {
            Icon(if (store.state.isSeller) Icons.Outlined.Storefront else Icons.Outlined.PersonOutline,
                null, Modifier.padding(20.dp).size(28.dp), tint = MaterialTheme.colorScheme.primary)
        }
        Column(Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(5.dp)) {
            Text(profile.displayName, style = MaterialTheme.typography.titleLarge)
            Muted(profile.email)
            Badge("${profile.role.title} · ${profile.provider.title}")
        }
    }
    Panel {
        SummaryRow("Телефон", profile.phone.ifBlank { "Не добавлен" })
        SecondaryButton(if (profile.phone.isBlank()) "Добавить телефон" else "Изменить телефон") { store.navigate(Page.EditProfile) }
        Muted("По желанию, для связи по заказу.")
    }
    MenuRow(if (store.state.isSeller) "Данные магазина" else "Личные данные") { store.navigate(Page.EditProfile) }
    MenuRow("Мои заказы") { store.selectTab(Page.Orders) }
    if (store.state.isSeller) MenuRow("Добавить товар") { store.navigate(Page.AddProduct) }
    SecondaryButton("Выйти из аккаунта", store::logout)
    if (store.state.isDemo) Muted("Демо-данные хранятся до закрытия приложения.")
    else {
        var confirmDelete by rememberSaveable { mutableStateOf(false) }
        TextButton(onClick = { confirmDelete = true }) { Text("Удалить аккаунт", color = MaterialTheme.colorScheme.error) }
        if (confirmDelete) AlertDialog(
            onDismissRequest = { confirmDelete = false }, title = { Text("Удалить аккаунт?") },
            text = { Text("Профиль, товары и связанные заказы будут удалены без возможности восстановления. Для подтверждения требуется недавний вход (до 15 минут).") },
            confirmButton = { TextButton(onClick = { confirmDelete = false; store.deleteAccount() }) { Text("Удалить") } },
            dismissButton = { TextButton(onClick = { confirmDelete = false }) { Text("Отмена") } },
        )
    }
}

@Composable
private fun MenuRow(label: String, onClick: () -> Unit) {
    Column {
        Row(Modifier.fillMaxWidth().clickable(onClick = onClick).padding(vertical = 17.dp), verticalAlignment = Alignment.CenterVertically) {
            Text(label, Modifier.weight(1f))
            Icon(Icons.AutoMirrored.Outlined.KeyboardArrowRight, null)
        }
        HorizontalDivider()
    }
}

@Composable
fun EditProfileScreen(store: AppStore) = PageColumn {
    val profile = store.state.profile ?: return@PageColumn
    var name by rememberSaveable { mutableStateOf(profile.name) }
    var phone by rememberSaveable { mutableStateOf(profile.phone) }
    var error by rememberSaveable { mutableStateOf<String?>(null) }
    Field("Email", profile.email, {}, readOnly = true)
    Field(if (store.state.isSeller) "Название магазина · необязательно" else "Имя · необязательно", name, { name = it; error = null })
    Field("Телефон · необязательно", phone, { phone = it; error = null }, "+996 700 123 456", KeyboardType.Phone)
    Muted("Можно оставить пустым или удалить номер в любой момент.")
    Field("Город", "Бишкек", {}, readOnly = true)
    ErrorMessage(error)
    PrimaryButton("Сохранить") { error = store.saveProfile(name, phone) }
}
