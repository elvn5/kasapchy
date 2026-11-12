package kg.kasapchy.mobile

import androidx.compose.ui.test.*
import androidx.compose.ui.test.junit4.createAndroidComposeRule
import androidx.test.espresso.Espresso
import org.junit.Assume.assumeTrue
import org.junit.Before
import org.junit.Rule
import org.junit.Test

class WireflowTest {
    @Before fun demoOnly() { assumeTrue(BuildConfig.DEMO_MODE) }
    @get:Rule val compose = createAndroidComposeRule<MainActivity>()

    private fun tap(text: String) {
        compose.onNodeWithText(text).performScrollTo().performClick()
    }
    private fun fill(label: String, text: String) {
        compose.onNode(hasSetTextAction() and hasText(label)).performTextReplacement(text)
        Espresso.closeSoftKeyboard()
    }
    private fun signIn(seller: Boolean = false, social: String? = null) {
        tap("Создать аккаунт")
        if (seller) tap("Касапчы")
        tap("Продолжить")
        if (social == null) {
            fill("Email", "buyer@example.com")
            tap("Продолжить с email")
            fill("Код из письма", "1234")
            tap("Создать аккаунт")
        } else {
            tap("Продолжить с $social")
            tap("Продолжить с демо-аккаунтом")
        }
        compose.waitForIdle()
    }

    @Test fun buyerCanOrderAndEditOptionalPhone() {
        signIn()
        tap("Говяжья мякоть")
        tap("Оформить заказ")
        tap("Подтвердить заказ")
        tap("Перейти к заказу")
        compose.onNodeWithText("Ждём подтверждения").assertExists()
        compose.onNodeWithText("Профиль").performClick()
        tap("Добавить телефон")
        fill("Телефон · необязательно", "+996 700 123 456")
        tap("Сохранить")
        compose.onNodeWithText("+996 700 123 456").assertExists()
        tap("Изменить телефон")
        fill("Телефон · необязательно", "")
        tap("Сохранить")
        compose.onNodeWithText("Не добавлен").assertExists()
    }

    @Test fun sellerCanPublishAndAdvanceIncomingOrder() {
        signIn(seller = true, social = "Google")
        tap("Добавить товар")
        fill("Название", "Demo steak")
        fill("Цена за кг, сом", "800")
        fill("В наличии, кг", "10")
        tap("Опубликовать товар")
        compose.onNodeWithText("Demo steak").assertExists()
        compose.onNodeWithText("Мои заказы").performClick()
        tap("Открыть заказ")
        tap("Принять заказ")
        compose.onNodeWithText("Заказ принят").assertExists()
    }

    @Test fun appleLoginAndSystemBackWork() {
        signIn(social = "Apple")
        tap("Говяжья мякоть")
        Espresso.pressBack()
        compose.onNodeWithText("Предложения рядом").assertExists()
        compose.onNodeWithText("Профиль").performClick()
        compose.onNodeWithText("demo@privaterelay.example.com").assertExists()
    }
}
