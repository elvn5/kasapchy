package kg.kasapchy.mobile.shared.ui

import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Typography
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.sp

// Derived from nsp_logo.svg. Darker blue is used for accessible text/button contrast.
val LogoBlue = Color(0xFF007EC5)
val LogoSilver = Color(0xFFA7AFB2)
private val Light = lightColorScheme(
    primary = Color(0xFF0076B8), onPrimary = Color.White,
    primaryContainer = Color(0xFFDFEFF9), onPrimaryContainer = Color(0xFF005582),
    secondary = Color(0xFF5F717C), secondaryContainer = Color(0xFFEDF5FA),
    onSecondaryContainer = Color(0xFF172D3B),
    background = Color.White, surface = Color.White, onSurface = Color(0xFF172D3B),
    onBackground = Color(0xFF172D3B), surfaceVariant = Color(0xFFEDF5FA),
    onSurfaceVariant = Color(0xFF5F717C), outline = Color(0xFF83959F), outlineVariant = Color(0xFFDCE4E8),
)
private val Dark = darkColorScheme(
    primary = Color(0xFF6BC6FA), onPrimary = Color(0xFF082838),
    primaryContainer = Color(0xFF203D50), onPrimaryContainer = Color(0xFFC2E6FD),
    secondary = LogoSilver, secondaryContainer = Color(0xFF203441),
    onSecondaryContainer = Color(0xFFEAF2F7),
    background = Color(0xFF131D24), surface = Color(0xFF131D24), onSurface = Color(0xFFEAF2F7),
    onBackground = Color(0xFFEAF2F7), surfaceVariant = Color(0xFF203441),
    onSurfaceVariant = Color(0xFFABB9C3), outline = Color(0xFF8498A4), outlineVariant = Color(0xFF354651),
)
private val Type = Typography(
    headlineLarge = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.SemiBold, fontSize = 32.sp, lineHeight = 36.sp, letterSpacing = (-1).sp),
    headlineMedium = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.SemiBold, fontSize = 28.sp, lineHeight = 33.sp, letterSpacing = (-0.7).sp),
    titleLarge = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.SemiBold, fontSize = 21.sp, lineHeight = 27.sp),
    titleMedium = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.SemiBold, fontSize = 16.sp, lineHeight = 22.sp),
    bodyLarge = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.Normal, fontSize = 16.sp, lineHeight = 24.sp),
    bodyMedium = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.Normal, fontSize = 14.sp, lineHeight = 21.sp),
    bodySmall = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.Normal, fontSize = 12.sp, lineHeight = 18.sp),
    labelLarge = TextStyle(fontFamily = FontFamily.SansSerif, fontWeight = FontWeight.SemiBold, fontSize = 14.sp, lineHeight = 20.sp),
)
@Composable
fun KasapchyTheme(content: @Composable () -> Unit) {
    MaterialTheme(colorScheme = if (isSystemInDarkTheme()) Dark else Light, typography = Type, content = content)
}
