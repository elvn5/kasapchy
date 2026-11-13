package kg.kasapchy.mobile

import androidx.compose.ui.test.*
import androidx.compose.ui.test.junit4.createAndroidComposeRule
import androidx.test.espresso.Espresso
import org.junit.Assume.assumeFalse
import org.junit.Rule
import org.junit.Test

/** Runs only with apiDebug against compose.local.yml + the real Go process, never external mail. */
class ApiWireflowTest {
    @get:Rule val compose = createAndroidComposeRule<MainActivity>()
    private fun tap(text: String) { compose.onNodeWithText(text).performScrollTo().performClick() }
    private fun fill(label: String, value: String) {
        compose.onNode(hasSetTextAction() and hasText(label)).performTextReplacement(value)
        Espresso.closeSoftKeyboard()
    }
    private fun waitFor(text: String) {
        compose.waitUntil(25000) { compose.onAllNodesWithText(text).fetchSemanticsNodes().isNotEmpty() }
    }
    private fun login(email: String, seller: Boolean) {
        tap("Создать аккаунт")
        if (seller) tap("Касапчы")
        tap("Продолжить")
        fill("Email", email)
        fill("Пароль", "test-only-strong-password")
        fill("Повторите пароль", "test-only-strong-password")
        tap("Создать аккаунт")
        waitFor("Предложения рядом")
    }
    @Test fun passwordSellerProductBuyerOrderAndAccountDeletion() {
        assumeFalse(BuildConfig.DEMO_MODE)
        val stamp = System.currentTimeMillis()
        login("seller-$stamp@example.com", true)
        tap("Добавить товар")
        fill("Название", "Steak $stamp")
        fill("Цена за кг, сом", "680")
        fill("В наличии, кг", "3")
        tap("Опубликовать товар")
        waitFor("Steak $stamp")
        compose.onNodeWithText("Профиль").performClick()
        waitFor("Удалить аккаунт")
        tap("Выйти из аккаунта")
        waitFor("Создать аккаунт")
        login("buyer-$stamp@example.com", false)
        tap("Steak $stamp")
        tap("Оформить заказ")
        tap("Подтвердить заказ")
        waitFor("Перейти к заказу")
        tap("Перейти к заказу")
        compose.onNodeWithText("Ждём подтверждения").assertExists()
        compose.onNodeWithText("Профиль").performClick()
        waitFor("Удалить аккаунт")
        tap("Удалить аккаунт")
        compose.onNodeWithText("Удалить", useUnmergedTree = true).performClick()
        waitFor("Создать аккаунт")
    }
}
