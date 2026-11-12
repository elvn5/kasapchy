package kg.kasapchy.mobile.shared.api

import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.plugins.HttpTimeout
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.request.*
import io.ktor.http.*
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import kg.kasapchy.mobile.entities.*

@Serializable data class SessionResponse(val token: String, val user: Profile)
@Serializable data class ChallengeResponse(val challengeId: String, val nonce: String = "")
@Serializable data class FailureResponse(val error: String)
class ApiFailure(val code: Int, override val message: String) : Exception(message)

class MarketApi(val baseUrl: String) {
    var token: String? = null
    private val client = HttpClient {
        install(ContentNegotiation) { json(Json { ignoreUnknownKeys = true }) }
        install(HttpTimeout) { requestTimeoutMillis = 25000; connectTimeoutMillis = 10000 }
    }
    init { require(baseUrl.startsWith("https://") || baseUrl.startsWith("http://10.0.2.2:") || baseUrl.startsWith("http://127.0.0.1:")) }
    suspend inline fun <reified T> get(path: String): T = request(HttpMethod.Get, path).body()
    suspend inline fun <reified T> post(path: String, body: Map<String, String>): T = request(HttpMethod.Post, path, body).body()
    suspend fun request(method: HttpMethod, path: String, fields: Map<String, String>? = null, rawJson: String? = null): io.ktor.client.statement.HttpResponse {
        val response = client.request(baseUrl.trimEnd('/') + "/v1" + path) {
            this.method = method
            token?.let { bearerAuth(it) }
            contentType(ContentType.Application.Json)
            if (fields != null) setBody(fields)
            if (rawJson != null) setBody(rawJson)
        }
        if (!response.status.isSuccess()) {
            val message = runCatching { response.body<FailureResponse>().error }.getOrDefault("Сервис временно недоступен.")
            throw ApiFailure(response.status.value, message)
        }
        return response
    }
    fun close() { client.close() }
}

interface SessionStorage {
    fun read(): String?
    fun save(token: String?)
}
