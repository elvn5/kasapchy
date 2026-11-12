import org.jetbrains.kotlin.gradle.dsl.JvmTarget

plugins {
    kotlin("multiplatform")
    kotlin("plugin.serialization")
    id("com.android.library")
    id("org.jetbrains.compose")
    id("org.jetbrains.kotlin.plugin.compose")
}
kotlin {
    androidTarget { compilerOptions { jvmTarget.set(JvmTarget.JVM_17) } }
    jvm() // Runs the platform-independent domain and navigation tests without a simulator.
    iosArm64()
    iosSimulatorArm64()
    targets.withType<org.jetbrains.kotlin.gradle.plugin.mpp.KotlinNativeTarget>().configureEach {
        binaries.framework {
            baseName = "KasapchyShared"
            isStatic = true
        }
    }
    sourceSets {
        commonMain.dependencies {
            implementation("io.ktor:ktor-client-core:3.3.1")
            implementation("io.ktor:ktor-client-content-negotiation:3.3.1")
            implementation("io.ktor:ktor-serialization-kotlinx-json:3.3.1")
            implementation(compose.runtime)
            implementation(compose.foundation)
            implementation(compose.material3)
            implementation(compose.materialIconsExtended)
        }
        androidMain.dependencies { implementation("io.ktor:ktor-client-okhttp:3.3.1") }
        iosMain.dependencies { implementation("io.ktor:ktor-client-darwin:3.3.1") }
        jvmMain.dependencies { implementation("io.ktor:ktor-client-cio:3.3.1") }
        commonTest.dependencies { implementation(kotlin("test")) }
    }
}
android {
    namespace = "kg.kasapchy.mobile.shared"
    compileSdk = 36
    defaultConfig { minSdk = 24 }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
}
