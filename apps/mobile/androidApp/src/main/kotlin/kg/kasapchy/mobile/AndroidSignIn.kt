package kg.kasapchy.mobile

import android.app.Activity
import android.content.Intent
import android.net.Uri
import androidx.credentials.CredentialManager
import androidx.credentials.GetCredentialRequest
import com.google.android.libraries.identity.googleid.GetSignInWithGoogleOption
import com.google.android.libraries.identity.googleid.GoogleIdTokenCredential
import io.ktor.client.call.body
import io.ktor.http.HttpMethod
import kg.kasapchy.mobile.entities.AuthProvider
import kg.kasapchy.mobile.shared.api.*
import kotlinx.coroutines.delay
import java.security.MessageDigest
import java.security.SecureRandom

class AndroidSignIn(private val activity: Activity) {
    suspend fun login(provider: AuthProvider, role: String, baseUrl: String): SessionResponse {
        val api = MarketApi(baseUrl)
        try {
            if (provider == AuthProvider.Google) {
                if (BuildConfig.GOOGLE_CLIENT_ID.isBlank()) throw ApiFailure(503, "Вход через Google ещё не настроен.")
                val challenge = api.post<ChallengeResponse>("/auth/oidc/challenge", mapOf("provider" to "Google", "role" to role))
                val option = GetSignInWithGoogleOption.Builder(BuildConfig.GOOGLE_CLIENT_ID).setNonce(challenge.nonce).build()
                val credential = CredentialManager.create(activity).getCredential(activity, GetCredentialRequest.Builder().addCredentialOption(option).build()).credential
                val idToken = GoogleIdTokenCredential.createFrom(credential.data).idToken
                return api.post("/auth/oidc/verify", mapOf("provider" to "Google", "challengeId" to challenge.challengeId, "idToken" to idToken))
            }
            if (BuildConfig.APPLE_CLIENT_ID.isBlank()) throw ApiFailure(503, "Вход через Apple ещё не настроен.")
            val bytes = ByteArray(32).also { SecureRandom().nextBytes(it) }
            val proof = bytes.joinToString("") { "%02x".format(it) }
            val hash = MessageDigest.getInstance("SHA-256").digest(proof.toByteArray()).joinToString("") { "%02x".format(it) }
            val challenge = api.post<ChallengeResponse>("/auth/oidc/challenge", mapOf("provider" to "Apple", "role" to role, "proofHash" to hash))
            val url = Uri.parse("https://appleid.apple.com/auth/authorize").buildUpon()
                .appendQueryParameter("client_id", BuildConfig.APPLE_CLIENT_ID)
                .appendQueryParameter("redirect_uri", baseUrl.trimEnd('/') + "/v1/auth/apple/callback")
                .appendQueryParameter("response_type", "code id_token").appendQueryParameter("response_mode", "form_post")
                .appendQueryParameter("scope", "email name").appendQueryParameter("state", challenge.challengeId)
                .appendQueryParameter("nonce", challenge.nonce).build()
            activity.startActivity(Intent(Intent.ACTION_VIEW, url))
            repeat(150) {
                delay(2000)
                val response = api.request(HttpMethod.Post, "/auth/apple/poll", mapOf("challengeId" to challenge.challengeId, "proof" to proof))
                if (response.status.value == 200) return response.body()
            }
            throw ApiFailure(408, "Время входа истекло. Попробуйте ещё раз.")
        } finally { api.close() }
    }
}
