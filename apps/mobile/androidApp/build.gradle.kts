import java.net.URI

plugins {
    id("com.android.application")
    kotlin("android")
    id("org.jetbrains.kotlin.plugin.compose")
}

// Upload key belongs to the publisher. Never fall back to a debug key for a release.
val uploadStore = providers.environmentVariable("KASAPCHY_UPLOAD_STORE_FILE")
val uploadPassword = providers.environmentVariable("KASAPCHY_UPLOAD_STORE_PASSWORD")
val uploadAlias = providers.environmentVariable("KASAPCHY_UPLOAD_KEY_ALIAS")
val uploadKeyPassword = providers.environmentVariable("KASAPCHY_UPLOAD_KEY_PASSWORD")
val hasUploadKey = listOf(uploadStore, uploadPassword, uploadAlias, uploadKeyPassword).all { it.isPresent }

fun literal(value: String) = "\"" + value.replace("\\", "\\\\").replace("\"", "\\\"") + "\""
val apiUrl = providers.gradleProperty("kasapchy.apiBaseUrl").orElse("https://not-configured.invalid")
val googleClientId = providers.gradleProperty("kasapchy.googleClientId").orElse("")
val appleClientId = providers.gradleProperty("kasapchy.appleClientId").orElse("")
val privacyUrl = providers.gradleProperty("kasapchy.privacyUrl").orElse("")
val deletionUrl = providers.gradleProperty("kasapchy.deletionUrl").orElse("")

android {
    namespace = "kg.kasapchy.mobile"
    testBuildType = providers.gradleProperty("kasapchy.testBuildType").orElse("debug").get()
    compileSdk = 36
    defaultConfig {
        applicationId = "kg.kasapchy.mobile"
        minSdk = 24
        targetSdk = 36
        versionCode = providers.gradleProperty("kasapchy.versionCode").orElse("1").get().toInt().also {
            require(it in 1..2100000000) { "kasapchy.versionCode must be a positive Play version code." }
        }
        versionName = providers.gradleProperty("kasapchy.versionName").orElse("0.1.0").get()
        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"
        resValue("string", "app_name", "Касапчы")
        buildConfigField("boolean", "DEMO_MODE", "false")
        buildConfigField("String", "API_BASE_URL", literal(apiUrl.get()))
        buildConfigField("String", "GOOGLE_CLIENT_ID", literal(googleClientId.get()))
        buildConfigField("String", "APPLE_CLIENT_ID", literal(appleClientId.get()))
        buildConfigField("String", "PRIVACY_URL", literal(privacyUrl.get()))
        buildConfigField("String", "DELETION_URL", literal(deletionUrl.get()))
    }
    signingConfigs {
        if (hasUploadKey) create("upload") {
            storeFile = file(uploadStore.get())
            storePassword = uploadPassword.get()
            keyAlias = uploadAlias.get()
            keyPassword = uploadKeyPassword.get()
        }
    }
    buildTypes {
        getByName("debug") {
            buildConfigField("boolean", "DEMO_MODE", "true")
            applicationIdSuffix = ".debug"
            versionNameSuffix = "-debug"
            resValue("string", "app_name", "Касапчы Debug")
        }
        getByName("release") {
            isDebuggable = false
            isMinifyEnabled = true
            isShrinkResources = true
            proguardFiles(getDefaultProguardFile("proguard-android-optimize.txt"), "proguard-rules.pro")
            if (hasUploadKey) signingConfig = signingConfigs.getByName("upload")
        }
        create("apiDebug") {
            initWith(getByName("debug"))
            buildConfigField("boolean", "DEMO_MODE", "false")
            buildConfigField("String", "API_BASE_URL", literal(providers.gradleProperty("kasapchy.localApiBaseUrl").orElse("http://10.0.2.2:18080").get()))
            applicationIdSuffix = ".apidebug"
            matchingFallbacks += listOf("debug")
        }
        // Exercises the same R8/resource-shrinking pipeline without masquerading as the live marketplace.
        create("demoRelease") {
            initWith(getByName("release"))
            buildConfigField("boolean", "DEMO_MODE", "true")
            applicationIdSuffix = ".demo"
            versionNameSuffix = "-demo"
            resValue("string", "app_name", "Касапчы Демо")
            matchingFallbacks += listOf("release")
        }
    }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
    buildFeatures { compose = true; buildConfig = true }
}

val verifyProductionReadiness by tasks.registering {
    group = "verification"
    description = "Validates the publisher configuration for a real production build."
    doLast {
        val url = URI(apiUrl.get())
        check(url.scheme == "https" && !url.host.isNullOrBlank() && !url.host.endsWith(".invalid") && url.host != "localhost") {
            "Production API HTTPS URL is not configured. Set kasapchy.apiBaseUrl after deployment."
        }
        check(googleClientId.get().isNotBlank() && appleClientId.get().isNotBlank()) { "Configure Google and Apple client IDs before production." }
        check(listOf(privacyUrl.get(), deletionUrl.get()).all { it.startsWith("https://") }) {
            "Configure published privacy and account-deletion HTTPS URLs."
        }
    }
}
val verifyUploadSigning by tasks.registering {
    group = "verification"
    doLast {
        check(hasUploadKey) { "Set all four KASAPCHY_UPLOAD_* variables locally; never send passwords in chat or commit them." }
        check(file(uploadStore.get()).isFile) { "Upload keystore file does not exist." }
    }
}
tasks.configureEach {
    if (name == "preReleaseBuild") dependsOn(verifyProductionReadiness, verifyUploadSigning)
}

dependencies {
    implementation(project(":shared"))
    implementation("io.ktor:ktor-client-core:3.3.1")
    implementation("androidx.credentials:credentials:1.5.0")
    implementation("androidx.credentials:credentials-play-services-auth:1.5.0")
    implementation("com.google.android.libraries.identity.googleid:googleid:1.1.1")
    implementation("androidx.activity:activity-compose:1.10.1")
    implementation("androidx.lifecycle:lifecycle-viewmodel-ktx:2.9.4")
    androidTestImplementation("androidx.compose.ui:ui-test-junit4:1.9.4")
    androidTestImplementation("androidx.test:runner:1.6.2")
    androidTestImplementation("androidx.test.ext:junit:1.2.1")
    androidTestImplementation("androidx.test.espresso:espresso-core:3.6.1")
}
kotlin { compilerOptions { jvmTarget.set(org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17) } }
