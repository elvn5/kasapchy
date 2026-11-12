package kg.kasapchy.mobile

import androidx.compose.ui.test.*
import androidx.compose.ui.test.junit4.createAndroidComposeRule
import androidx.test.espresso.Espresso
import org.json.JSONObject
import org.junit.Assume.assumeFalse
import org.junit.Rule
import org.junit.Test
import java.net.URL

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
    private fun mailCode(email: String): String {
        repeat(20) {
            val messages = JSONObject(URL("http://10.0.2.2:18025/api/v1/messages").readText()).getJSONArray("messages")
            for (i in 0 until messages.length()) {
                val item = messages.getJSONObject(i)
                if (item.getJSONArray("To").getJSONObject(0).getString("Address") == email) {
                    val message = JSONObject(URL("http://10.0.2.2:18025/api/v1/message/" + item.getString("ID")).readText())
                    return Regex("\\b[0-9]{6}\\b").find(message.getString("Text"))!!.value
                }
            }
            Thread.sleep(300)
        }
        error("Local Mailpit did not receive the test code")
    }
    private fun login(email: String, seller: Boolean) {
        tap("Создать аккаунт")
        if (seller) tap("Касапчы")
        tap("Продолжить")
        fill("Email", email)
        tap("Продолжить с email")
        waitFor("Проверьте почту")
        fill("Код из письма", mailCode(email))
        tap("Создать аккаунт")
        waitFor("Предложения рядом")
    }
    @Test fun realEmailSellerProductBuyerOrderAndAccountDeletion() {
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
