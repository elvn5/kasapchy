package kg.kasapchy.mobile

import android.content.ContentValues
import android.graphics.Bitmap
import android.provider.MediaStore
import androidx.compose.ui.test.*
import androidx.compose.ui.test.junit4.createAndroidComposeRule
import androidx.test.platform.app.InstrumentationRegistry
import org.junit.Assume.assumeTrue
import org.junit.Rule
import org.junit.Test

/** Opt-in captures of actual app UI. No illustrations of unimplemented screens. */
class StoreScreenshotsTest {
    @get:Rule val compose = createAndroidComposeRule<MainActivity>()
    private fun tap(text: String) = compose.onNodeWithText(text).performScrollTo().performClick()
    private fun capture(name: String) {
        compose.waitForIdle()
        val bitmap = checkNotNull(InstrumentationRegistry.getInstrumentation().uiAutomation.takeScreenshot())
        val resolver = InstrumentationRegistry.getInstrumentation().targetContext.contentResolver
        val values = ContentValues().apply {
            put(MediaStore.Images.Media.DISPLAY_NAME, name)
            put(MediaStore.Images.Media.MIME_TYPE, "image/png")
            put(MediaStore.Images.Media.RELATIVE_PATH, "Pictures/Kasapchy")
        }
        val uri = checkNotNull(resolver.insert(MediaStore.Images.Media.EXTERNAL_CONTENT_URI, values))
        resolver.openOutputStream(uri)!!.use { bitmap.compress(Bitmap.CompressFormat.PNG, 100, it) }
    }
    @Test fun captureCurrentWireflow() {
        assumeTrue(BuildConfig.DEMO_MODE)
        assumeTrue(InstrumentationRegistry.getArguments().getString("captureStore") == "true")
        tap("Создать аккаунт"); tap("Продолжить"); tap("Продолжить с Google"); tap("Продолжить с демо-аккаунтом")
        capture("01-market.png")
        tap("Говяжья мякоть"); capture("02-product.png")
        tap("Оформить заказ"); capture("03-checkout.png")
        tap("Подтвердить заказ"); tap("Перейти к заказу"); capture("04-order.png")
    }
}
